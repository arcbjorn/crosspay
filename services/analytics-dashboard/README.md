# Analytics Dashboard

Real-time monitoring and metrics service for CrossPay Protocol.

## Features

- Real-time validator performance tracking
- Vault health and TVL monitoring  
- Payment flow analytics
- Privacy usage metrics
- WebSocket live updates
- RESTful metrics API

## Quick Start

```bash
# Install dependencies
go mod tidy

# Set environment variables
export PORT=8090
export RPC_ENDPOINT=http://localhost:8545

# Run service
go run .

# Access dashboard
open http://localhost:8090
```

## Configuration

Environment variables:

```bash
PORT=8090                    # HTTP server port
RPC_ENDPOINT=http://...      # Blockchain RPC endpoint
METRICS_INTERVAL=30s         # Collection interval
DB_CONNECTION=postgres://... # Database connection (optional)
```

## API Endpoints

### Metrics
- `GET /health` - Service health check
- `GET /metrics` - Comprehensive system metrics
- `GET /metrics/validators` - Validator performance data
- `GET /metrics/vault` - Vault health and TVL data
- `GET /metrics/payments` - Payment processing metrics
- `GET /metrics/privacy` - Privacy feature usage

### Real-time Updates
- `GET /ws` - WebSocket endpoint for live updates

## WebSocket Events

```json
{
  "type": "validator_update",
  "data": {
    "address": "0x742d35...",
    "status": "active", 
    "performance_score": 98.5
  },
  "timestamp": "2025-08-31T12:00:00Z"
}
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   Metrics       │    │   WebSocket     │
│   Collector     │───▶│   Hub           │
│                 │    │                 │
│ • Blockchain    │    │ • Live Updates  │
│ • Validators    │    │ • Client Mgmt   │
│ • Vault Data    │    │ • Broadcasting  │
└─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│   HTTP API      │    │   Dashboard     │
│                 │    │   Frontend      │
│ • REST Endpoints│    │ • Real-time UI  │
│ • JSON Response │    │ • Charts/Graphs │
│ • Rate Limiting │    │ • Alerts        │
└─────────────────┘    └─────────────────┘
```

## Metrics Collected

### Validator Metrics
- Stake amounts and status
- Uptime and performance scores
- Validation counts and latency
- Slashing events and penalties

### Vault Metrics  
- Total TVL across all tranches
- APY rates and utilization
- Slashing event history
- Insurance fund balance

### Payment Metrics
- Total payment volume
- Success rates and latency
- Private vs public payments
- Cross-chain statistics

### Privacy Metrics
- Encrypted payment counts
- Disclosure request patterns
- Sealed bid grant activity
- Compliance interactions

## Testing

```bash
# Run unit tests
go test ./...

# Test with race detection
go test -race ./...

# Benchmark tests
go test -bench=. ./...
```

## Deployment

### Docker
```bash
docker build -t analytics-dashboard .
docker run -p 8090:8090 analytics-dashboard
```

### Kubernetes
```bash
kubectl apply -f k8s/analytics-dashboard.yaml
```

## Monitoring

The service exposes Prometheus metrics at `/metrics` and includes:
- HTTP request duration and counts
- WebSocket connection counts
- Data collection success rates
- Memory and CPU usage

## Security

- Rate limiting on all endpoints
- CORS configuration for web access
- Input validation on all parameters
- No sensitive data logged