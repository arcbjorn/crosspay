# CrossPay Analytics System

Real-time monitoring, metrics collection, and analytics for the CrossPay Protocol ecosystem.

## Overview

The analytics system provides comprehensive monitoring across all CrossPay components:
- Validator network performance and health
- Vault deposits, yields, and slashing events
- Payment processing metrics and success rates
- Privacy feature usage and compliance
- Real-time alerts and notifications

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Smart         │    │   Analytics     │    │   Dashboard     │
│   Contracts     │───▶│   Collector     │───▶│   Frontend      │
│                 │    │                 │    │                 │
│ • Events        │    │ • Aggregation   │    │ • Real-time UI  │
│ • State Queries │    │ • Time Series   │    │ • Charts        │
│ • Error Logs    │    │ • Caching       │    │ • Alerts        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       ▼                       │
         │              ┌─────────────────┐              │
         │              │   WebSocket     │              │
         └──────────────│   Hub           │◀─────────────┘
                        │                 │
                        │ • Live Updates  │
                        │ • Broadcasting  │
                        │ • Client Mgmt   │
                        └─────────────────┘
```

## Key Metrics

### Validator Performance
```json
{
  "address": "0x742d35Cc6634C0532925a3b8D34300e8",
  "stake": "15000000000000000000",
  "uptime": 99.5,
  "validation_count": 1250,
  "slash_count": 0,
  "performance_score": 98.5,
  "last_activity": "2025-08-31T11:58:00Z",
  "status": "active"
}
```

### Vault Health
```json
{
  "total_tvl": "1000000000000000000000",
  "junior_tvl": "200000000000000000000",
  "mezzanine_tvl": "300000000000000000000",
  "senior_tvl": "500000000000000000000",
  "junior_apy": 12.0,
  "mezzanine_apy": 8.0,
  "senior_apy": 5.0,
  "insurance_fund": "50000000000000000000",
  "utilization_rates": {
    "junior": 20.0,
    "mezzanine": 30.0,
    "senior": 50.0
  }
}
```

### Payment Analytics
```json
{
  "total_payments": 15420,
  "private_payments": 3845,
  "validated_payments": 12675,
  "average_amount": "500000000000000000",
  "total_volume": "7710000000000000000000",
  "validation_latency_ms": 2850.0,
  "success_rate": 98.7
}
```

### Privacy Usage
```json
{
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
}
```

## Real-time Updates

### WebSocket Events
The analytics system broadcasts real-time updates via WebSocket:

```javascript
const ws = new WebSocket('ws://localhost:8090/ws');

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  
  switch(update.type) {
    case 'validator_update':
      // Validator status changed
      console.log('Validator update:', update.data);
      break;
      
    case 'vault_update':
      // Vault event (deposit/withdrawal/slashing)
      console.log('Vault update:', update.data);
      break;
      
    case 'payment_update':
      // New payment processed
      console.log('Payment update:', update.data);
      break;
      
    case 'heartbeat':
      // Regular health check
      console.log('System healthy:', update.data);
      break;
  }
};
```

## API Reference

### GET /metrics
Returns comprehensive system metrics including all validator, vault, payment, and privacy data.

### GET /metrics/validators
Returns detailed validator performance metrics including uptime, validation counts, and performance scores.

### GET /metrics/vault
Returns vault health metrics including TVL, APY rates, utilization, and slashing event history.

### GET /metrics/payments
Returns payment processing statistics including volume, success rates, and validation latency.

### GET /metrics/privacy
Returns privacy feature usage including encrypted payments, disclosure patterns, and sealed bid activity.

### WebSocket /ws
Real-time event stream for live dashboard updates.

## Alerting System

### Alert Types
- **Validator Offline**: Validator hasn't responded in 5 minutes
- **Consensus Failure**: Unable to reach 67% threshold
- **Vault Imbalance**: Tranche ratios exceed target ranges
- **High Slashing**: Multiple slashing events in short period
- **Privacy Breach**: Unauthorized disclosure attempts

### Alert Channels
- WebSocket broadcast to connected clients
- HTTP webhook notifications
- Email alerts for critical issues
- Slack/Discord integration

## Data Storage

### Time Series Data
- Validator performance over time
- Vault TVL and yield trends
- Payment volume patterns
- Network health metrics

### Retention Policy
- Real-time data: 24 hours
- Hourly aggregates: 30 days  
- Daily aggregates: 1 year
- Monthly aggregates: Permanent

## Performance

### Collection Performance
- 30-second collection intervals
- <1 second update latency
- 1000+ concurrent WebSocket connections
- 10MB/hour data generation

### Query Performance
- Metrics API: <100ms response time
- Historical data: <500ms response time
- Real-time updates: <50ms latency
- Dashboard load: <2 seconds

## Integration

### SDK Usage
```typescript
import { CrossPayAnalytics } from '@crosspay/analytics-sdk';

const analytics = new CrossPayAnalytics({
  baseUrl: 'http://localhost:8090'
});

// Get current metrics
const metrics = await analytics.getMetrics();

// Subscribe to updates
analytics.subscribe('validator_update', (data) => {
  console.log('Validator updated:', data);
});
```

### Custom Dashboards
```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8090/ws');

// Build custom charts
const chart = new Chart(ctx, {
  type: 'line',
  data: {
    datasets: [{
      label: 'Validator Uptime',
      data: validatorUptimeData
    }]
  }
});

// Update chart on new data
ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  if (update.type === 'validator_update') {
    chart.data.datasets[0].data.push(update.data.uptime);
    chart.update();
  }
};
```

## Security Considerations

### Data Privacy
- No sensitive user data collected
- Validator addresses anonymized in public views
- Payment amounts aggregated only
- Compliance data access restricted

### Access Control
- Public metrics available without authentication
- Detailed data requires appropriate permissions
- Admin functions protected by role-based access
- Rate limiting prevents abuse

### Data Integrity
- Cryptographic verification of blockchain data
- Checksums on all data transfers
- Immutable audit logs
- Regular data validation checks