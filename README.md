# AgriFinance Platform

Financial platform for agricultural lending and payments.

## Architecture

```
main.go → handlers → services → repositories → PostgreSQL
                     ↓
                Double-entry ledger
```

## Week 1 (Active)

**Endpoints:**
```
POST /wallet/create    # Create wallet for user
GET  /wallet/:id       # Get wallet
GET  /wallet/:id/transactions?limit=20&offset=0  # List transactions
GET  /health           # Health check
```

## Setup & Testing

### 1. PostgreSQL (agri_finance database)
```bash
# Create database
createdb agri_finance

# Run migration
psql -d agri_finance -f migrations/001_initial_schema.sql
```

### 2. Dependencies
```bash
go mod tidy
```

### 3. Config (`config.yaml`)
Update `database.password`, `redis.password` if needed.

### 4. Build & Run
```bash
go build -o agri-finance .
./agri-finance
# or
go run .
```

Server runs on `http://0.0.0.0:8080`

### 5. Test Week 1 API

```bash
# 1. Create wallet
curl -X POST http://localhost:8080/wallet/create \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "currency": "NGN",
    "tier": 1
  }'

# 2. Get wallet (use returned ID)
curl http://localhost:8080/wallet/550e8400-e29b-41d4-a716-446655440001

# 3. List transactions
curl "http://localhost:8080/wallet/550e8400-e29b-41d4-a716-446655440001/transactions?limit=5"
```

## Future Weeks (Implemented, commented out)

**Week 2:** Transfers, deposits, withdrawals, idempotency middleware  
**Week 3:** Escrow, KYC integration, loans  
**Week 4:** Admin APIs, JWT auth, rate limiting, audit logs

## Tech Stack

- Go 1.23
- Gin (HTTP)
- pgx (PostgreSQL)
- go-redis (Redis)
- JWT (auth)
- UUID (IDs)

## Local Development

```bash
# Install deps
go mod tidy

# Build
go build

# Test
go test ./...

# Run
go run main.go
```

## Docker (optional)

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o agri-finance .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/agri-finance .
COPY config.yaml .
EXPOSE 8080
CMD ["./agri-finance"]
```

