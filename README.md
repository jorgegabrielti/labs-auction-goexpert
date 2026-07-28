# Go Expert: Desafio Concorrência (Sistema de Leilão)

Este repositório contém a resolução do desafio prático de **Concorrência com Golang** do treinamento **Go Expert** da **Full Cycle**.

O objetivo do desafio é implementar uma rotina concorrente em background (Goroutines) que monitore a duração dos leilões e realize o **fechamento automático** dos leilões expirados, atualizando o status correspondente no banco de dados MongoDB.

---

## 🛠️ Tecnologias Utilizadas

- **Linguagem**: Go (Golang) 1.20
- **Framework Web**: Gin-Gonic (HTTP)
- **Banco de Dados**: MongoDB (Persistência e Cache de Verificação)
- **Containerização**: Docker e Docker Compose

---

## ⚙️ Configuração das Variáveis de Ambiente

As configurações de tempo do leilão e gravação de lances em lote são definidas no arquivo `.env` localizado em `cmd/auction/.env`.

Principais variáveis relacionadas à duração e concorrência:
- `AUCTION_DURATION` (ou `AUCTION_INTERVAL`): Determina a duração do ciclo de vida de um leilão ativo (ex: `10s`, `1m`, `5m`).
- `MAX_BATCH_SIZE`: Quantidade máxima de lances agrupados no canal de concorrência antes de fazer a inserção em massa no banco.
- `BATCH_INSERT_INTERVAL`: Tempo máximo de espera para fazer a inserção em lote no MongoDB de lances pendentes em memória (ex: `20s`).

---

## 🚀 Como Rodar o Projeto

Toda a infraestrutura do banco de dados MongoDB e a aplicação Go estão configuradas para subir via Docker Compose.

### 1. Iniciar os Serviços
Execute o seguinte comando na raiz do projeto para construir a imagem e subir os containers em segundo plano:
```bash
docker compose up -d --build
```

A aplicação ficará disponível na porta local `:8080`.

### 2. Parar os Serviços
Para derrubar os containers e limpar os recursos criados:
```bash
docker compose down
```

---

## 🧪 Como Rodar os Testes Automatizados

Implementamos um teste de integração de ponta a ponta (`create_auction_test.go`) para testar o fechamento automático assíncrono. O teste cria um leilão de teste com expiração curta (`1s`), aguarda o tempo e verifica se o status foi alterado no banco.

Caso você não tenha um banco de dados MongoDB de testes de pé localmente na máquina host, o teste detecta a ausência de conexão e realiza um `t.Skip` gracioso em vez de falhar.

Para executar os testes na máquina host:
```bash
go test -v ./internal/infra/database/auction/...
```

---

## 🌐 Endpoints REST da API

### Leilões (Auctions)
- **Criar Leilão**: `POST /auction`
  - Body (JSON): `{"product_name": "Nome", "category": "Categoria", "description": "Descrição", "condition": 1}`
- **Listar Leilões**: `GET /auction` (Permite filtros via query string: `?status=0&category=categoria`)
- **Buscar por ID**: `GET /auction/:auctionId`
- **Leilão Vencedor**: `GET /auction/winner/:auctionId` (Retorna os dados do leilão e o maior lance ofertado).

### Lances (Bids)
- **Registrar Lance**: `POST /bid`
  - Body (JSON): `{"user_id": "UUID_USER", "auction_id": "UUID_AUCTION", "amount": 150.0}`
- **Listar Lances por Leilão**: `GET /bid/:auctionId`
