# Relay Network

Symbiotic validator network service for CrossPay Protocol consensus and validation.

## Features

- Validator registration and stake management
- BFT consensus with 67% threshold
- Signature aggregation and proof construction
- P2P mesh networking
- Automated slashing for misbehavior
- High-value payment validation

## Quick Start

```bash
# Install dependencies
go mod tidy

# Generate or load validator key
export KEY_PATH=./validator.key

# Set configuration
export PORT=8080
export P2P_PORT=9090
export CONTRACT_ADDRESS=0x742d35...
export RPC_ENDPOINT=http://localhost:8545

# Run validator node
go run .
```

## Configuration

Environment variables:

```bash
PORT=8080                           # HTTP API port
KEY_PATH=./validator.key            # Validator private key file
CONTRACT_ADDRESS=0x742d35...        # RelayValidator contract address
RPC_ENDPOINT=http://localhost:8545  # Blockchain RPC endpoint
CHAIN_ID=1337                       # Network chain ID

# P2P Networking
P2P_PORT=9090                       # P2P listen port
BOOTSTRAP_PEERS=peer1:9090,peer2:9090 # Initial peer connections
MAX_PEERS=50                        # Maximum peer connections

# Validation Settings
VALIDATION_TIMEOUT=300              # Validation timeout (seconds)
MAX_CONCURRENT_VALIDATIONS=10       # Concurrent validation limit
SIGNATURE_REQUIRED=true             # Require signature validation
```

## API Endpoints

### Health & Status
- `GET /health` - Node health check
- `GET /status` - Detailed node status with peer info
- `GET /peers` - Connected peer information

### Validation
- `POST /validate` - Request network validation
- `POST /sign` - Submit signature for validation request
- `POST /register` - Register validator on network

## Validation Flow

```
1. Payment Request
   │
   ▼
2. Validation Request ──────┐
   │                       │
   ▼                       ▼
3. Broadcast to Peers ──▶ Sign Message
   │                       │
   ▼                       ▼
4. Collect Signatures ──▶ Aggregate Proof
   │
   ▼
5. Submit to Contract
```

## P2P Network

The validator network uses a custom P2P protocol for:
- Validation request broadcasting
- Signature sharing
- Network topology maintenance
- Peer discovery and health checks

### Message Types
```json
{
  "type": "validation_request",
  "request_id": 12345,
  "payment_id": 67890,
  "message_hash": "0xa1b2c3...",
  "timestamp": "2025-08-31T12:00:00Z"
}

{
  "type": "signature_share", 
  "request_id": 12345,
  "signature": "0x1a2b3c...",
  "signer": "0x742d35...",
  "timestamp": "2025-08-31T12:00:00Z"
}
```

## Security

### Validator Security
- Private keys stored securely (hardware HSM recommended)
- Signature verification before acceptance
- Rate limiting on validation requests
- Peer authentication and encryption

### Network Security
- BFT consensus requires 67% honest validators
- Economic security through stake requirements
- Slashing penalties for malicious behavior
- Timeout mechanisms prevent stalling

### Operational Security
- Regular key rotation procedures
- Monitoring and alerting systems
- Backup validator infrastructure
- Incident response procedures

## Key Management

```bash
# Generate new validator key
go run . --generate-key --key-path ./validator.key

# Load existing key
export KEY_PATH=./existing-validator.key

# Hardware HSM integration (production)
export HSM_PROVIDER=pkcs11
export HSM_CONFIG=/path/to/hsm.conf
```

## Monitoring

### Metrics Exposed
- Validation request counts and latency
- Peer connection health
- Signature aggregation success rates
- Stake amount and validator status

### Logging
- Structured JSON logging
- Validation event tracking
- P2P network activity
- Error and warning conditions

## Testing

```bash
# Run unit tests
go test ./...

# Integration tests with mock network
go test -tags=integration ./...

# Load testing
go test -bench=BenchmarkValidation ./...
```

## Deployment

### Single Node
```bash
# Build binary
go build -o relay-validator

# Run with systemd
sudo systemctl start relay-validator
```

### Docker
```bash
docker build -t relay-validator .
docker run -p 8080:8080 -p 9090:9090 relay-validator
```

### Kubernetes
```bash
kubectl apply -f k8s/relay-validator.yaml
```

## Network Participation

### Registration
1. Stake minimum 10 ETH in RelayValidator contract
2. Start validator node with proper configuration
3. Connect to bootstrap peers
4. Begin participating in validations

### Best Practices
- Maintain 99%+ uptime
- Respond to validation requests within 30 seconds
- Keep software updated
- Monitor network health and peer connections
- Backup private keys securely

## Troubleshooting

### Common Issues
- **Connection refused**: Check P2P port accessibility
- **Validation timeout**: Verify network connectivity to peers
- **Key not found**: Ensure KEY_PATH points to valid private key
- **Insufficient stake**: Verify validator registration on contract

### Debug Mode
```bash
export LOG_LEVEL=debug
go run . --debug
```