# Hylo Wallet Tracker API

Read-only REST API for tracking Solana wallet activity and metrics for the Hylo protocol. Provides real-time wallet balances (hyUSD, sHYUSD, xSOL), price data (SOL/USD, xSOL pricing), and transaction history.

## Quick Start

1. **Set up environment variables:**

```bash
cp .env.example .env
# Edit .env with your Helios RPC endpoints
```

2. **Run the API:**

```bash
make run
# Or: go run ./cmd/api/main.go
```

3. **Check health:**

```bash
curl http://localhost:8080/health
```

## Environment Variables

```bash
PORT=8080
RPC_HTTP_URL=https://mainnet.helius-rpc.com
RPC_WS_URL=wss://mainnet.helius-rpc.com
```

## API Documentation

### Swagger/OpenAPI

Interactive API documentation is available via Swagger UI:

- **Swagger UI**: http://localhost:8080/swagger/index.html (when server is running)
- **API Specification**:
  - JSON: `docs/api/swagger.json`
  - YAML: `docs/api/swagger.yaml`

### Generate Documentation

```bash
make swagger      # Generate Swagger documentation
make swagger-fmt  # Format Swagger annotations
```

## API Endpoints

### Current Endpoints

- `GET /health` - Service health and Solana RPC connectivity status
- `GET /swagger/*` - Swagger UI and API documentation

### Planned Endpoints

- `GET /price` - Current SOL/USD, xSOL/SOL, and xSOL/USD prices
- `GET /wallet/:address/balances` - Wallet balances for hyUSD, sHYUSD, xSOL
- `GET /wallet/:address/trades` - xSOL trade history with pagination
- `GET /events` - Server-Sent Events for real-time updates

## Development

```bash
make test         # Run tests
make build        # Build binary
make clean        # Clean artifacts
make swagger      # Generate API documentation
make swagger-fmt  # Format Swagger annotations
```
