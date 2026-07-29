# Go Expert: Desafio Concorrência — Sistema de Leilão

Resolução do desafio prático de **Concorrência com Golang** do treinamento **Go Expert (Full Cycle)**.

## 🎯 Objetivo do Desafio

Implementar o **fechamento automático de leilões** utilizando **Goroutines** em background. Quando um leilão é criado, uma goroutine é disparada assincronamente e — ao expirar o tempo definido em `AUCTION_DURATION` — atualiza o status do leilão para `Completed` no MongoDB sem nenhuma intervenção manual.

### O que foi implementado
- Goroutine de fechamento automático disparada a cada criação de leilão em `create_auction.go`
- Verificador de boot que reencerra leilões órfãos ao reiniciar a aplicação
- Variável de ambiente `AUCTION_DURATION` para controlar o tempo de expiração
- Teste de integração automatizado que valida o comportamento

---

## ⚙️ Variáveis de Ambiente

Configuradas em `cmd/auction/.env`:

| Variável | Descrição | Exemplo |
|---|---|---|
| `AUCTION_DURATION` | **Duração do leilão** — altere para testar o fechamento | `20s`, `1m`, `5m` |
| `BATCH_INSERT_INTERVAL` | Intervalo de inserção de lances em lote | `20s` |
| `MAX_BATCH_SIZE` | Tamanho máximo do lote de lances | `4` |
| `MONGODB_URL` | URL de conexão com o MongoDB | `mongodb://...` |
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

### 2. Configurar o tempo do leilão (opcional)
Para testar mais rapidamente, edite `cmd/auction/.env` e reduza `AUCTION_DURATION`:
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

## ✅ Como Validar o Fechamento Automático (Passo a Passo)

Esta é a validação completa do requisito principal do desafio:

### Passo 1 — Criar um leilão
```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{"product_name":"Teclado Mecanico","category":"Perifericos","description":"Teclado sem fio para teste","condition":1}'
```
**Resultado esperado:** HTTP `201 Created`

### Passo 2 — Listar os leilões ativos (status=0)
```bash
curl http://localhost:8080/auction?status=0
```
**Resultado esperado:** o leilão criado aparece com `"status": 0` (Active).

### Passo 3 — Aguardar o tempo configurado em `AUCTION_DURATION`
Aguarde o tempo definido (padrão: `20s`).

### Passo 4 — Verificar o fechamento automático
```bash
curl http://localhost:8080/auction?status=1
```
**Resultado esperado:** o leilão que estava ativo agora aparece com `"status": 1` (Completed) — **sem nenhuma intervenção manual**.

---

## 🧪 Testes Automatizados

O teste de integração (`create_auction_test.go`) valida o ciclo completo de fechamento. Ele requer o MongoDB em execução.

### Com Docker (recomendado)
```bash
# 1. Suba apenas o banco de dados
docker compose up -d mongodb

# 2. Execute os testes a partir da raiz do projeto
AUCTION_DURATION=1s MONGODB_URL=mongodb://admin:admin@localhost:27017/auctions?authSource=admin go test -v ./internal/infra/database/auction/...
```

### O que o teste verifica
1. Cria um leilão com `AUCTION_DURATION=1s`
2. Confirma que o status inicial é `Active` (0)
3. Aguarda `1.5s`
4. Confirma que o status foi automaticamente alterado para `Completed` (1)

---

## 🌐 Endpoints da API

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/auction` | Criar um novo leilão |
| `GET` | `/auction?status=0` | Listar leilões ativos |
| `GET` | `/auction?status=1` | Listar leilões finalizados |
| `GET` | `/auction/:auctionId` | Buscar leilão por ID |
| `GET` | `/auction/winner/:auctionId` | Buscar vencedor do leilão |
| `POST` | `/bid` | Registrar um lance |
| `GET` | `/bid/:auctionId` | Listar lances de um leilão |
| `GET` | `/user/:userId` | Buscar usuário por ID |

### Exemplo de corpo para criação de leilão
```json
{
  "product_name": "Produto Teste",
  "category": "Categoria",
  "description": "Descrição com ao menos 10 caracteres",
  "condition": 1
}
```

> **Condition**: `1` = Novo, `2` = Usado, `3` = Recondicionado

### Exemplo de corpo para lance
```json
{
  "user_id": "uuid-do-usuario",
  "auction_id": "uuid-do-leilao",
  "amount": 150.00
}
```
