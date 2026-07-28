package auction

import (
	"context"
	"os"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCreateAuction_AutomaticClose(t *testing.T) {
	// Configurar conexão com o MongoDB para teste
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		// Fallback para conectar de fora do docker se rodado localmente
		mongoURL = "mongodb://admin:admin@localhost:27017/auctions?authSource=admin"
	}
	mongoDBName := "auctions_test"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Skipf("Skipping integration test: MongoDB not connected: %v", err)
		return
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("Skipping integration test: MongoDB not running (ping failed): %v", err)
		return
	}

	database := client.Database(mongoDBName)
	// Limpa a coleção de teste antes de iniciar
	database.Collection("auctions").Drop(ctx)

	// Configura o timer para 1 segundo
	t.Setenv("AUCTION_DURATION", "1s")

	// Inicializa o repositório
	repo := NewAuctionRepository(database)

	// Cria a entidade de leilão
	auction, errInternal := auction_entity.CreateAuction(
		"Playstation 5",
		"Eletrônicos",
		"Console semi-novo com 2 controles e 4 jogos",
		auction_entity.New,
	)
	if errInternal != nil {
		t.Fatalf("Error creating auction entity: %v", errInternal)
	}

	// Executa a criação no repositório
	errCreate := repo.CreateAuction(ctx, auction)
	if errCreate != nil {
		t.Fatalf("Error inserting auction in DB: %v", errCreate)
	}

	// Verifica se inicialmente o leilão está com status Active (0)
	var auctionMongo AuctionEntityMongo
	errFind := repo.Collection.FindOne(ctx, bson.M{"_id": auction.Id}).Decode(&auctionMongo)
	if errFind != nil {
		t.Fatalf("Error finding created auction: %v", errFind)
	}

	if auctionMongo.Status != auction_entity.Active {
		t.Errorf("Expected initial status to be Active (0), got %v", auctionMongo.Status)
	}

	// Aguarda o timeout configurado (1s) mais uma margem de segurança (500ms)
	time.Sleep(1500 * time.Millisecond)

	// Verifica se o status mudou para Completed (1) automaticamente via goroutine
	errFind = repo.Collection.FindOne(ctx, bson.M{"_id": auction.Id}).Decode(&auctionMongo)
	if errFind != nil {
		t.Fatalf("Error finding auction after sleep: %v", errFind)
	}

	if auctionMongo.Status != auction_entity.Completed {
		t.Errorf("Expected final status to be Completed (1) after timeout, got %v", auctionMongo.Status)
	}
}
