# ENS Resolver Service

Enterprise-grade ENS (Ethereum Name Service) resolution with subname registry and caching for CrossPay Protocol.

## Features

### Core Resolution
- **Forward Resolution**: ENS name → Ethereum address
- **Reverse Resolution**: Ethereum address → ENS name  
- **Batch Processing**: Resolve multiple names in single request
- **Avatar Support**: ENS avatar image retrieval
- **Text Records**: Custom text record resolution

### Subname Registry
- **Domain Delegation**: Authorize subname creation
- **Bulk Registration**: Register multiple subnames efficiently
- **Fee Management**: Configurable registration and renewal fees
- **Expiration Tracking**: TTL and renewal management
- **Community Namespaces**: Organizational subdomain management

### Performance Optimization
- **Intelligent Caching**: TTL-based cache with auto-eviction
- **Cache Analytics**: Hit/miss ratios and performance metrics
- **Batch Operations**: Minimize RPC calls
- **Response Time**: < 1 second with caching

## API Endpoints

### ENS Resolution
- `GET /api/ens/resolve/:name` - Resolve ENS name to address
- `GET /api/ens/reverse/:address` - Reverse resolve address
- `POST /api/ens/resolve/batch` - Batch resolve names
- `GET /api/ens/avatar/:name` - Get ENS avatar URL
- `GET /api/ens/text/:name/:key` - Get text record value
- `GET /api/ens/search` - Search ENS names

### Subname Management
- `POST /api/subnames/register` - Register new subname
- `GET /api/subnames/list/:domain` - List domain subnames
- `POST /api/subnames/bulk` - Bulk register subnames
- `DELETE /api/subnames/revoke/:subname` - Revoke subname

### Cache Management
- `GET /api/cache/stats` - Cache statistics
- `DELETE /api/cache/clear` - Clear entire cache
- `DELETE /api/cache/entry/:key` - Clear specific entry

## Usage Examples

### Resolve ENS Name
```bash
curl http://localhost:8082/api/ens/resolve/alice.eth
# Response: {"name": "alice.eth", "address": "0x1234...", "avatar": "..."}
```

### Batch Resolution
```bash
curl -X POST http://localhost:8082/api/ens/resolve/batch \
  -H "Content-Type: application/json" \
  -d '{"names": ["alice.eth", "bob.eth", "crosspay.eth"]}'
```

### Register Subname
```bash
curl -X POST http://localhost:8082/api/subnames/register \
  -H "Content-Type: application/json" \
  -d '{
    "subname": "pay",
    "domain": "crosspay.eth", 
    "owner": "0xabcd...",
    "address": "0x1111...",
    "ttl": 31536000
  }'
```

### Get Cache Statistics
```bash
curl http://localhost:8082/api/cache/stats
```

## Configuration

Environment variables:
- `ENS_RPC_URL`: Ethereum RPC for ENS queries
- `CACHE_TTL`: Default cache TTL in seconds (3600)
- `SERVICE_NAME`: Service identifier

## Cache System

### Cache Layers
1. **Forward Cache**: Name → Address mappings
2. **Reverse Cache**: Address → Name mappings  
3. **Subname Cache**: Subdomain registrations

### Cache Policies
- **TTL-based expiration**: Configurable per record type
- **Auto-eviction**: Background cleanup every 5 minutes
- **Size limits**: Prevents memory exhaustion
- **Hit rate monitoring**: Performance optimization

### Cache Statistics
```json
{
  "forward_entries": 150,
  "reverse_entries": 120,
  "subname_entries": 45,
  "hit_rate": 87.5,
  "cache_hits": 1750,
  "cache_misses": 250
}
```

## Subname Registry

### Domain Delegation
Domain owners can delegate subname management:
```solidity
function delegateDomain(
    string domain,
    address owner,
    uint256 registrationFee,
    uint256 maxSubnames
) external
```

### Bulk Registration
Efficient batch registration for organizations:
```bash
curl -X POST http://localhost:8082/api/subnames/bulk \
  -d '{
    "domain": "company.eth",
    "subnames": ["alice", "bob", "charlie"],
    "owner": "0xabcd..."
  }'
```

## Error Handling

Graceful handling of:
- Network failures and timeouts
- Invalid ENS names and addresses
- Rate limiting from ENS providers
- Cache eviction under memory pressure
- Malformed resolution requests

## Performance Targets

- **Resolution**: < 1 second (cached), < 3 seconds (uncached)
- **Batch operations**: < 5 seconds for 50 names
- **Cache hit rate**: > 80% under normal load
- **Availability**: 99.9% uptime

## Development

```bash
# Install dependencies
go mod tidy

# Run locally
go run .

# Build Docker image
docker build -t ens-resolver .

# Run tests
go test ./...
```

## Integration

### With Payment Processor
```go
// Resolve ENS names during payment creation
senderAddr := resolveENS(payment.SenderENS)
recipientAddr := resolveENS(payment.RecipientENS)
```

### With Frontend
```javascript
// Resolve ENS name in UI
const resolved = await fetch('/api/ens/resolve/alice.eth')
const { address } = await resolved.json()
```

### With Smart Contracts
```solidity
// Store resolved names in payment
payment.senderENS = senderENS;
payment.recipientENS = recipientENS;
```

## Monitoring

Metrics tracked:
- Resolution latency by type
- Cache hit/miss ratios
- Error rates by endpoint
- Subname registration volume
- Memory usage and cache size