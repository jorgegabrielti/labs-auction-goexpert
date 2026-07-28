package auction

import (
	"context"
	"os"
	"time"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	repo := &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
	repo.checkAndCloseExpiredAuctionsOnStartup()
	return repo
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	// Dispara goroutine em background para fechamento automático do leilão
	duration := getAuctionDuration()
	go func(id string, d time.Duration) {
		time.Sleep(d)
		updateCtx := context.Background()
		filter := bson.M{"_id": id}
		update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}
		_, err := ar.Collection.UpdateOne(updateCtx, filter, update)
		if err != nil {
			logger.Error("Error trying to automatically close auction: "+id, err)
		} else {
			logger.Info("Auction automatically closed: " + id)
		}
	}(auctionEntity.Id, duration)

	return nil
}

func getAuctionDuration() time.Duration {
	auctionInterval := os.Getenv("AUCTION_DURATION")
	if auctionInterval == "" {
		auctionInterval = os.Getenv("AUCTION_INTERVAL")
	}
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5
	}
	return duration
}

func (ar *AuctionRepository) checkAndCloseExpiredAuctionsOnStartup() {
	go func() {
		ctx := context.Background()
		filter := bson.M{"status": auction_entity.Active}
		cursor, err := ar.Collection.Find(ctx, filter)
		if err != nil {
			logger.Error("Error on startup finding active auctions", err)
			return
		}
		defer cursor.Close(ctx)

		duration := getAuctionDuration()
		now := time.Now()

		for cursor.Next(ctx) {
			var auctionMongo AuctionEntityMongo
			if err := cursor.Decode(&auctionMongo); err != nil {
				logger.Error("Error on startup decoding auction", err)
				continue
			}

			createdTime := time.Unix(auctionMongo.Timestamp, 0)
			expirationTime := createdTime.Add(duration)

			if now.After(expirationTime) {
				updateFilter := bson.M{"_id": auctionMongo.Id}
				update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}
				_, err := ar.Collection.UpdateOne(ctx, updateFilter, update)
				if err != nil {
					logger.Error("Error closing expired auction on startup", err)
				} else {
					logger.Info("Auction closed on startup: " + auctionMongo.Id)
				}
			} else {
				remainingTime := time.Until(expirationTime)
				go func(id string, sleepTime time.Duration) {
					time.Sleep(sleepTime)
					updateFilter := bson.M{"_id": id}
					update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}
					_, err := ar.Collection.UpdateOne(context.Background(), updateFilter, update)
					if err != nil {
						logger.Error("Error automatically closing scheduled auction", err)
					}
				}(auctionMongo.Id, remainingTime)
			}
		}
	}()
}

