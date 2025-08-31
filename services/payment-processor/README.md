# Payment Processor Service

Unified API gateway and orchestration service that coordinates between storage, oracle, ENS, and blockchain services for CrossPay Protocol.

## Features

### Service Orchestration
- **API Gateway**: Single entry point for all payment operations
- **Service Discovery**: Dynamic routing to microservices
- **Load Balancing**: Request distribution across service instances
- **Circuit Breaker**: Fault tolerance and graceful degradation

### Payment Workflows
- **End-to-end Processing**: From creation to receipt generation
- **Multi-chain Support**: Lisk, Base, Citrea coordination
- **Oracle Integration**: Price locking and verification
- **ENS Resolution**: Automatic name-to-address resolution

### Data Aggregation
- **Unified API**: Consistent interface across all features
- **Cross-service Queries**: Aggregate data from multiple sources
- **Analytics Collection**: Payment volume and usage metrics
- **Real-time Updates**: Live payment status tracking

## API Endpoints

### Core Payment Operations
- `POST /api/payments/create` - Create payment with full integration
- `GET /api/payments/:id` - Get payment with all associated data
- `POST /api/payments/complete/:id` - Complete payment
- `POST /api/payments/refund/:id` - Process refund
- `GET /api/payments/user/:address` - Get user payment history

### Integrated Receipt Management
- `POST /api/receipts/generate/:paymentId` - Generate receipt with storage
- `GET /api/receipts/download/:id` - Download receipt via CID
- `GET /api/receipts/verify/:cid` - Verify receipt authenticity
- `GET /api/receipts/payment/:paymentId` - Get all receipts for payment

### Oracle Integration
- `GET /api/oracle/price/:symbol` - Get current oracle price
- `POST /api/oracle/random/request` - Request random number
- `GET /api/oracle/random/status/:requestId` - Check random status
- `POST /api/oracle/proof/submit` - Submit external proof
- `GET /api/oracle/proof/verify/:proofId` - Verify external proof

### ENS Integration  
- `GET /api/ens/resolve/:name` - Resolve ENS name
- `GET /api/ens/reverse/:address` - Reverse resolve address
- `POST /api/ens/resolve/batch` - Batch ENS resolution

### Storage Integration
- `POST /api/storage/upload` - Upload files via storage worker
- `GET /api/storage/retrieve/:cid` - Retrieve files by CID
- `GET /api/storage/cost/:size` - Get storage cost estimate

### Analytics & Monitoring
- `GET /api/analytics/stats` - Overall system statistics
- `GET /api/analytics/payments/volume` - Payment volume data
- `GET /api/analytics/receipts/stats` - Receipt generation stats

## Usage Examples

### Create Payment with Full Integration
```bash
curl -X POST http://localhost:8083/api/payments/create \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "0x742d35Cc...",
    "token": "0x0000000000000000000000000000000000000000",
    "amount": "1000000000000000000", 
    "sender_ens": "alice.eth",
    "recipient_ens": "bob.eth",
    "metadata_uri": "ipfs://QmTest123"
  }'
```

Response includes:
- Payment ID and status
- Current oracle price (locked)
- Auto-generated receipt CID
- ENS resolution confirmation
- Transaction hash

### Payment with Receipt Generation
The service automatically:
1. Resolves ENS names to addresses
2. Locks current oracle price  
3. Creates blockchain payment
4. Generates receipt and uploads to Filecoin
5. Returns CID for receipt retrieval

## Service Communication

### Microservices Architecture
```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Client    │───►│   Payment    │───►│  Storage    │
│ Application │    │  Processor   │    │  Worker     │
└─────────────┘    └──────┬───────┘    └─────────────┘
                          │
                    ┌─────┼─────┐
                    │     │     │
              ┌─────▼─┐ ┌─▼──┐ ┌▼──────┐
              │Oracle │ │ENS │ │ Cache │
              │Service│ │    │ │ Layer │
              └───────┘ └────┘ └───────┘
```

### Service Discovery
- **Health Checks**: Continuous service availability monitoring
- **Retry Logic**: Exponential backoff for failed requests
- **Timeout Handling**: Configurable timeouts per service
- **Fallback Routes**: Alternative paths when services unavailable

## Configuration

Environment variables:
- `STORAGE_SERVICE_URL`: Storage worker endpoint
- `ORACLE_SERVICE_URL`: Oracle service endpoint  
- `ENS_SERVICE_URL`: ENS resolver endpoint
- `DATABASE_URL`: PostgreSQL connection string

## Error Handling

### Service Failures
- **Graceful Degradation**: Continue with reduced functionality
- **Retry Mechanisms**: Automatic retry with backoff
- **Circuit Breakers**: Prevent cascade failures
- **Fallback Data**: Use cached/default values when possible

### User Experience
- **Meaningful Errors**: Clear error messages for users
- **Partial Success**: Show what succeeded vs failed
- **Retry Options**: Allow users to retry failed operations
- **Status Updates**: Real-time operation progress

## Performance

### Response Time Targets
- Payment creation: < 5 seconds
- Receipt generation: < 10 seconds
- Data retrieval: < 2 seconds
- Analytics queries: < 3 seconds

### Throughput
- 100 payments/minute sustained
- 1000 ENS resolutions/minute
- 50 receipt generations/minute
- 500 analytics queries/minute

## Database Schema

Key tables managed by payment processor:
- `payments` - Payment records with all metadata
- `receipts` - Receipt tracking and CID storage
- `oracle_requests` - Oracle operation logging
- `ens_cache` - ENS resolution cache
- `analytics_daily` - Aggregated daily metrics

## Development

```bash
# Install dependencies
go mod tidy

# Run locally with hot reload
go run .

# Build production image
docker build -t payment-processor .

# Run tests
go test ./...

# Database migrations
go run migrations/*.go
```

## Integration Testing

```bash
# Test full payment flow
curl -X POST http://localhost:8083/api/payments/create \
  -d '{"recipient":"bob.eth","amount":"1000000","token":"0x0"}'

# Verify receipt generation  
curl http://localhost:8083/api/receipts/generate/[payment_id]

# Check oracle integration
curl http://localhost:8083/api/oracle/price/ETH/USD

# Test ENS resolution
curl http://localhost:8083/api/ens/resolve/alice.eth
```