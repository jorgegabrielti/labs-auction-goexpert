# Go Expert: Desafio Concorrência — Sistema de Leilão

Resolução do desafio prático de **Concorrência com Golang** do treinamento **Go Expert (Full Cycle)**.

## 🎯 Objetivo do Desafio

Implementar o **fechamento automático de leilões** utilizando **Goroutines** em background. Quando um leilão é criado, uma goroutine é disparada assincronamente e — ao expirar o tempo definido em `AUCTION_DURATION` — atualiza o status do leilão para `Completed` no MongoDB sem nenhuma intervenção manual.

### O que foi implementado

- **Goroutine de fechamento automático** disparada a cada criação de leilão em [`create_auction.go`](internal/infra/database/auction/create_auction.go)
- **Verificador de boot** (`checkAndCloseExpiredAuctionsOnStartup`) que reencerra leilões órfãos ao reiniciar a aplicação
- **Variável de ambiente `AUCTION_DURATION`** para controlar o tempo de expiração
- **Teste de integração automatizado** que valida o comportamento concorrente

---

## ⚙️ Variáveis de Ambiente

Configuradas em `cmd/auction/.env`:

| Variável | Descrição | Padrão |
|---|---|---|
| `AUCTION_DURATION` | **Duração do leilão** — altere para testar o fechamento | `20s` |
| `BATCH_INSERT_INTERVAL` | Intervalo de inserção de lances em lote | `20s` |
| `MAX_BATCH_SIZE` | Tamanho máximo do lote de lances | `4` |
| `MONGODB_URL` | URL de conexão com o MongoDB | `mongodb://admin:admin@mongodb:27017/...` |
| `MONGODB_DB` | Nome do banco de dados | `auctions` |

---

## 🚀 Como Rodar o Projeto

### Pré-requisitos
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) instalado e em execução.

### 1. Clonar o repositório
```bash
git clone https://github.com/jorgegabrielti/labs-auction-goexpert.git
cd labs-auction-goexpert
```

### 2. (Opcional) Ajustar o tempo do leilão
Para testar mais rapidamente, edite `cmd/auction/.env`:
```env
AUCTION_DURATION=20s
```

### 3. Subir os containers
```bash
docker compose up -d --build
```

A aplicação ficará disponível em `http://localhost:8080`.

### 4. Parar os containers
```bash
docker compose down
```

---

## ✅ Validação do Fechamento Automático (Passo a Passo)

Esta seção documenta a validação **real** realizada localmente, com os logs de saída.

### Passo 1 — Criar um leilão

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{"product_name":"Notebook Gamer","category":"Informatica","description":"Notebook para validacao do fechamento automatico","condition":1}'
```

**Resultado:** HTTP `201 Created`

---

### Passo 2 — Confirmar status ATIVO (status=0)

```bash
curl http://localhost:8080/auction?status=0
```

**Saída real obtida:**
```json
[
  {
    "id": "79c21fa3-6e38-465c-a92a-f2e43896c19d",
    "product_name": "Notebook Gamer",
    "category": "Informatica",
    "description": "Notebook para validacao do fechamento automatico",
    "condition": 1,
    "status": 0,
    "timestamp": "2026-07-29T00:06:48Z"
  }
]
```

✅ Leilão criado às **21:06:48** com `"status": 0` (Active).

---

### Passo 3 — Aguardar o tempo configurado

Aguarde o tempo definido em `AUCTION_DURATION` (padrão: `20s`).

---

### Passo 4 — Confirmar fechamento automático (status=1)

```bash
curl http://localhost:8080/auction?status=1
```

**Saída real obtida (25 segundos depois):**
```json
[
  {
    "id": "79c21fa3-6e38-465c-a92a-f2e43896c19d",
    "product_name": "Notebook Gamer",
    "category": "Informatica",
    "description": "Notebook para validacao do fechamento automatico",
    "condition": 1,
    "status": 1,
    "timestamp": "2026-07-29T00:06:48Z"
  }
]
```

✅ Às **21:07:13** o leilão aparece com `"status": 1` (Completed) — **sem nenhuma intervenção manual**.

> O fechamento ocorreu em ~25s (20s de duração + 5s de margem de verificação), provando que a **goroutine funcionou corretamente em background**.

---

## 🧪 Testes Automatizados

O teste de integração [`create_auction_test.go`](internal/infra/database/auction/create_auction_test.go) valida o ciclo completo automaticamente.

### Como executar

```bash
# 1. Suba apenas o banco de dados
docker compose up -d mongodb

# 2. Execute os testes (na raiz do projeto)
AUCTION_DURATION=1s \
MONGODB_URL=mongodb://admin:admin@localhost:27017/auctions?authSource=admin \
go test -v -timeout 30s ./internal/infra/database/auction/...
```

> **No Windows (PowerShell):**
> ```powershell
> $env:AUCTION_DURATION="1s"
> $env:MONGODB_URL="mongodb://admin:admin@localhost:27017/auctions?authSource=admin"
> go test -v -timeout 30s ./internal/infra/database/auction/...
> ```

### Saída real do teste executado localmente

```
=== RUN   TestCreateAuction_AutomaticClose
--- PASS: TestCreateAuction_AutomaticClose (1.56s)
PASS
ok      fullcycle-auction_go/internal/infra/database/auction    3.454s
```

✅ **Teste passou em 3.45s** com `AUCTION_DURATION=1s`.

### O que o teste verifica

| Passo | Ação | Resultado esperado |
|---|---|---|
| 1 | Cria um leilão com `AUCTION_DURATION=1s` | Leilão inserido no MongoDB |
| 2 | Lê o status imediatamente | `status = 0` (Active) |
| 3 | Aguarda 1.5s | — |
| 4 | Lê o status novamente | `status = 1` (Completed) — alterado pela goroutine |

---

## 🌐 Endpoints da API

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/auction` | Criar um novo leilão |
| `GET` | `/auction?status=0` | Listar leilões **ativos** |
| `GET` | `/auction?status=1` | Listar leilões **finalizados** |
| `GET` | `/auction/:auctionId` | Buscar leilão por ID |
| `GET` | `/auction/winner/:auctionId` | Buscar vencedor do leilão |
| `POST` | `/bid` | Registrar um lance |
| `GET` | `/bid/:auctionId` | Listar lances de um leilão |
| `GET` | `/user/:userId` | Buscar usuário por ID |

### Corpo para criação de leilão

```json
{
  "product_name": "Produto Teste",
  "category": "Categoria",
  "description": "Descrição com ao menos 10 caracteres",
  "condition": 1
}
```

> **condition:** `1` = Novo, `2` = Usado, `3` = Recondicionado

### Corpo para lance

```json
{
  "user_id": "uuid-do-usuario",
  "auction_id": "uuid-do-leilao",
  "amount": 150.00
}
```
