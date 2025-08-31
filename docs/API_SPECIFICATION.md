# CrossPay Protocol API Specification

## Overview

CrossPay exposes REST APIs for privacy-preserving payments, validator network management, vault operations, and real-time analytics.

## Base URLs

```
Analytics Dashboard: http://localhost:8090
Relay Validator: http://localhost:8080  
Payment Processor: http://localhost:3000 (from Module 1/2)
```

## Authentication

All APIs use role-based access control:
- **Public**: No authentication required
- **User**: Valid Ethereum signature required
- **Validator**: Registered validator address required
- **Admin**: Multi-sig or owner permission required

## Analytics Dashboard API

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-08-31T12:00:00Z",
  "service": "analytics-dashboard"
}
```

### GET /metrics
Get comprehensive system metrics.

**Response:**
```json
{
  "timestamp": "2025-08-31T12:00:00Z",
  "validator_metrics": {
    "0x742d35...": {
      "address": "0x742d35Cc6634C0532925a3b8D34300e8",
      "stake": "10000000000000000000",
      "uptime": 99.5,
      "validation_count": 1250,
      "slash_count": 0,
      "last_activity": "2025-08-31T11:58:00Z",
      "status": "active",
      "performance_score": 98.5
    }
  },
  "vault_metrics": {
    "total_tvl": "1000000000000000000000",
    "junior_tvl": "200000000000000000000",
    "mezzanine_tvl": "300000000000000000000", 
    "senior_tvl": "500000000000000000000",
    "junior_apy": 12.0,
    "mezzanine_apy": 8.0,
    "senior_apy": 5.0,
    "slashing_events": [],
    "insurance_fund": "50000000000000000000",
    "utilization_rates": {
      "junior": 20.0,
      "mezzanine": 30.0,
      "senior": 50.0
    }
  },
  "payment_metrics": {
    "total_payments": 15420,
    "private_payments": 3845,
    "validated_payments": 12675,
    "average_amount": "500000000000000000",
    "total_volume": "7710000000000000000000",
    "payments_by_status": {
      "pending": 45,
      "completed": 15200,
      "refunded": 125,
      "cancelled": 50
    },
    "validation_latency_ms": 2850.0,
    "success_rate": 98.7
  },
  "privacy_metrics": {
    "encrypted_payments": 3845,
    "disclosure_requests": 127,
    "approved_disclosures": 89,
    "sealed_bid_grants": 23,
    "privacy_usage_rate": 24.9,
    "disclosures_by_type": {
      "compliance": 45,
      "audit": 32,
      "participant": 12
    }
  },
  "system_status": "healthy"
}
```

### GET /metrics/validators
Get validator-specific metrics.

### GET /metrics/vault  
Get vault performance metrics.

### GET /metrics/payments
Get payment processing metrics.

### GET /metrics/privacy
Get privacy system metrics.

### GET /ws
WebSocket endpoint for real-time updates.

**WebSocket Messages:**
```json
{
  "type": "heartbeat",
  "data": {
    "connected_clients": 5,
    "server_time": "2025-08-31T12:00:00Z"
  },
  "timestamp": "2025-08-31T12:00:00Z"
}

{
  "type": "validator_update", 
  "data": {
    "address": "0x742d35...",
    "status": "slashed",
    "reason": "Failed validation"
  },
  "timestamp": "2025-08-31T12:00:00Z"
}

{
  "type": "vault_update",
  "data": {
    "event": "slashing_executed",
    "amount": "1000000000000000000",
    "tranches_affected": ["junior"]
  },
  "timestamp": "2025-08-31T12:00:00Z"
}
```

## Validator Network API

### GET /health
Validator node health check.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-08-31T12:00:00Z",
  "validator_address": "0x742d35Cc6634C0532925a3b8D34300e8",
  "is_registered": true,
  "stake": "15000000000000000000",
  "peer_count": 12,
  "pending_validations": 3,
  "network_running": true
}
```

### GET /status
Detailed validator status.

**Response:**
```json
{
  "validator_address": "0x742d35Cc6634C0532925a3b8D34300e8",
  "status": "active",
  "is_registered": true,
  "stake": "15000000000000000000",
  "peer_count": 12,
  "pending_validations": 3,
  "network_running": true,
  "peers": [
    {
      "address": "192.168.1.100:9090",
      "last_seen": "2025-08-31T11:59:00Z",
      "is_active": true
    }
  ]
}
```

### POST /validate
Request validation from network.

**Request:**
```json
{
  "payment_id": 12345,
  "message_hash": "0xa1b2c3...",
  "required_signatures": 3,
  "is_high_value": true
}
```

**Response:**
```json
{
  "request_id": 67890,
  "status": "requested",
  "deadline": "2025-08-31T12:05:00Z"
}
```

### POST /sign
Submit signature for validation request.

**Request:**
```json
{
  "request_id": 67890,
  "message_hash": "0xa1b2c3..."
}
```

**Response:**
```json
{
  "request_id": 67890,
  "payment_id": 12345,
  "signatures_count": 2,
  "required_signatures": 3,
  "signatures": {
    "0x742d35...": "0x1a2b3c...",
    "0x845f21...": "0x4d5e6f..."
  },
  "deadline": "2025-08-31T12:05:00Z"
}
```

### GET /peers
Get connected peer information.

**Response:**
```json
{
  "peer_count": 12,
  "peers": [
    {
      "address": "192.168.1.100:9090", 
      "last_seen": "2025-08-31T11:59:00Z",
      "is_active": true
    }
  ]
}
```

### POST /register
Register validator on network.

**Query Parameters:**
- `stake`: Stake amount in ETH

**Response:**
```json
{
  "status": "registration_requested",
  "address": "0x742d35Cc6634C0532925a3b8D34300e8",
  "stake": "15.0",
  "message": "Registration transaction should be submitted to the RelayValidator contract"
}
```

## Smart Contract Events

### ConfidentialPayments Events

```solidity
event ConfidentialPaymentCreated(
    uint256 indexed id,
    address indexed sender, 
    address indexed recipient,
    address token,
    string metadataURI,
    bool isPrivate
);

event DisclosureRequested(
    uint256 indexed paymentId,
    address indexed requester,
    string reason
);

event DisclosureRevealed(
    uint256 indexed paymentId,
    address indexed viewer,
    uint256 amount,
    uint256 fee
);
```

### RelayValidator Events

```solidity
event ValidatorRegistered(
    address indexed validator,
    uint256 stake
);

event ValidationRequested(
    uint256 indexed requestId,
    uint256 indexed paymentId,
    bytes32 messageHash,
    uint256 requiredSignatures,
    uint256 deadline,
    bool isHighValue
);

event ValidationCompleted(
    uint256 indexed requestId,
    bytes aggregatedSignature,
    uint256 signerCount
);

event ValidatorSlashed(
    address indexed validator,
    uint256 slashedAmount,
    string reason
);
```

### TrancheVault Events

```solidity
event Deposited(
    address indexed user,
    TrancheType indexed tranche,
    uint256 amount,
    uint256 shares
);

event Slashed(
    uint256 indexed eventId,
    uint256 totalAmount,
    uint256 juniorLoss,
    uint256 mezzanineLoss,
    uint256 seniorLoss,
    address validator,
    string reason
);

event YieldDistributed(
    TrancheType indexed tranche,
    uint256 totalYield,
    uint256 perTokenYield
);
```

## Error Codes

### HTTP Status Codes
- `200 OK`: Request successful
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service temporarily unavailable

### Contract Error Codes
- `InvalidPaymentId()`: Payment does not exist
- `UnauthorizedAction()`: Caller lacks permission
- `InsufficientStake()`: Below minimum stake requirement
- `ValidationExpired()`: Validation deadline passed
- `InsufficientBalance()`: Insufficient funds for operation
- `WithdrawalDelayNotMet()`: Must wait for withdrawal delay

## Rate Limits

### Analytics API
- Public endpoints: 100 requests/minute
- Authenticated endpoints: 1000 requests/minute
- WebSocket connections: 10 concurrent per IP

### Validator API  
- Registration: 1 request/hour per IP
- Validation requests: 100 requests/minute
- Status queries: 1000 requests/minute

## SDK Integration

### JavaScript/TypeScript
```typescript
import { CrossPayAnalytics } from '@crosspay/analytics-sdk';

const analytics = new CrossPayAnalytics({
  baseUrl: 'http://localhost:8090',
  apiKey: 'your-api-key' // Optional
});

// Get real-time metrics
const metrics = await analytics.getMetrics();

// Subscribe to real-time updates
analytics.subscribe('validator_update', (data) => {
  console.log('Validator updated:', data);
});
```

### Go
```go
import "github.com/crosspay/go-sdk/analytics"

client := analytics.NewClient("http://localhost:8090")
metrics, err := client.GetMetrics()
```

### Python
```python
from crosspay import AnalyticsClient

client = AnalyticsClient("http://localhost:8090")
metrics = client.get_metrics()
```

---

*API Version: 3.0*  
*Last Updated: August 31, 2025*  
*Breaking Changes: See CHANGELOG.md*