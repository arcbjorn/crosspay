# Oracle Service

Comprehensive oracle integration service providing FTSO price feeds, secure random numbers, and FDC external proof verification for CrossPay Protocol.

## Features

### FTSO (Flare Time Series Oracle)
- Real-time price feeds for major trading pairs
- Historical price data with 100-point history
- Staleness detection and circuit breaker
- Multi-currency support (ETH/USD, BTC/USD, cBTC/USD, etc.)

### Secure Random Numbers
- Cryptographically secure random generation
- Commit-reveal pattern with minimum delay
- Grant winner selection algorithms
- Auto-fulfillment after delay period

### FDC (Flare Data Connector)  
- External proof submission and verification
- Merkle proof validation
- Payment confirmation webhooks
- Proof status tracking and management

### Health Monitoring
- Service health checks across all protocols
- Circuit breaker for emergency pause/resume
- Response time monitoring
- Error rate tracking

## API Endpoints

### FTSO Price Feeds
- `GET /api/ftso/price/:symbol` - Get current price
- `GET /api/ftso/price/:symbol/history` - Get price history
- `POST /api/ftso/price/update` - Update price (admin)
- `GET /api/ftso/symbols` - List supported symbols

### Random Number Generation
- `POST /api/random/request` - Request random number
- `GET /api/random/status/:requestId` - Check request status
- `POST /api/random/fulfill` - Fulfill request (admin)
- `POST /api/random/winners` - Select random winners

### FDC External Proofs
- `POST /api/fdc/proof/submit` - Submit Merkle proof
- `GET /api/fdc/proof/verify/:proofId` - Verify proof
- `POST /api/fdc/proof/confirm` - Confirm verification
- `POST /api/fdc/webhook/payment` - Payment confirmation
- `GET /api/fdc/proofs` - Get proofs by transaction

### Health & Circuit Breaker
- `GET /api/oracle/status` - Overall oracle status
- `POST /api/oracle/healthcheck` - Trigger health check
- `POST /api/oracle/circuit-breaker/pause` - Emergency pause
- `POST /api/oracle/circuit-breaker/resume` - Resume operations

## Usage Examples

### Get Current Price
```bash
curl http://localhost:8081/api/ftso/price/ETH/USD
```

### Request Random Number
```bash
curl -X POST http://localhost:8081/api/random/request \
  -H "Content-Type: application/json" \
  -d '{"requester": "payment-app"}'
```

### Submit External Proof
```bash
curl -X POST http://localhost:8081/api/fdc/proof/submit \
  -H "Content-Type: application/json" \
  -d '{
    "merkle_root": "0xabc123...",
    "proof": ["0xdef456...", "0x789ghi..."],
    "data": "payment_confirmation_data"
  }'
```

### Check Oracle Health
```bash
curl http://localhost:8081/api/oracle/status
```

## Supported Trading Pairs

- ETH/USD - Ethereum to US Dollar
- BTC/USD - Bitcoin to US Dollar  
- cBTC/USD - Citrea Bitcoin to US Dollar
- FLR/USD - Flare to US Dollar
- USDC/USD - USD Coin to US Dollar

## Configuration

Environment variables:
- `FLARE_RPC_URL`: Flare network RPC endpoint
- `FTSO_API_URL`: FTSO API endpoint
- `FDC_API_URL`: FDC API endpoint

## Security Features

### Price Feed Protection
- Staleness threshold: 10 minutes
- Circuit breaker on consecutive failures
- Price deviation limits
- Fallback mechanisms

### Random Number Security  
- Minimum 1-minute fulfillment delay
- Cryptographically secure generation
- Request ID collision prevention
- Deterministic winner selection

### Proof Verification
- Merkle proof validation
- Content hash verification
- Timestamp validation
- Replay attack prevention

## Performance Targets

- Price feed update: < 2 seconds
- Random fulfillment: 60-120 seconds
- Proof verification: < 1 second
- Health check: < 500ms

## Development

```bash
# Install dependencies
go mod tidy

# Run locally
go run .

# Build Docker image
docker build -t oracle-service .

# Run tests
go test ./...
```

## Monitoring

The service exposes metrics for:
- Price feed accuracy and latency
- Random request fulfillment rates
- Proof submission success rates
- Circuit breaker activations
- Error rates by endpoint