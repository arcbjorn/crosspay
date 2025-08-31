# Storage Worker Service

Filecoin storage integration service for CrossPay Protocol receipts and documents using SynapseSDK.

## Features

- **SynapseSDK Integration**: Direct uploads to Filecoin Calibration testnet
- **Receipt Generation**: JSON and PDF receipt creation with QR codes
- **Queue System**: Async processing with retry logic and exponential backoff
- **Cost Estimation**: Real-time storage cost calculation
- **CID Management**: Content addressing and retrieval system

## API Endpoints

### Storage Operations
- `POST /api/storage/upload` - Upload file to Filecoin
- `GET /api/storage/retrieve/:cid` - Retrieve file by CID
- `GET /api/storage/cost/:size` - Estimate storage cost

### Receipt Operations  
- `POST /api/receipts/generate` - Generate payment receipt
- `GET /api/receipts/download/:id` - Download receipt file
- `GET /api/receipts/verify/:cid` - Verify receipt authenticity

### Health & Monitoring
- `GET /health` - Service health check

## Usage

### Upload File
```bash
curl -X POST http://localhost:8080/api/storage/upload \
  -F "file=@receipt.json" \
  -F "metadata={\"type\":\"receipt\"}"
```

### Generate Receipt
```bash
curl -X POST http://localhost:8080/api/receipts/generate \
  -H "Content-Type: application/json" \
  -d '{
    "payment_id": 123,
    "format": "pdf", 
    "language": "en"
  }'
```

### Retrieve by CID
```bash
curl http://localhost:8080/api/storage/retrieve/bafybeig...
```

## Configuration

Environment variables:
- `FILECOIN_RPC_URL`: Filecoin node RPC endpoint
- `STORAGE_API_KEY`: SynapseSDK API key
- `SERVICE_NAME`: Service identifier for logging

## Queue System

The service implements an async job queue with:
- 3 concurrent workers
- Exponential backoff retry (max 3 attempts)
- Dead letter queue for failed jobs
- Job status tracking and monitoring

## Error Handling

All endpoints handle:
- Network failures gracefully
- Storage service downtime
- Invalid file formats
- Timeout scenarios
- Rate limiting

## Performance

- Upload target: < 10 seconds
- Retrieval target: < 3 seconds  
- Receipt generation: < 5 seconds
- Queue processing: < 30 seconds per job

## Development

```bash
# Install dependencies
go mod tidy

# Run locally
go run .

# Build Docker image
docker build -t storage-worker .

# Run tests
go test ./...
```