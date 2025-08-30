## Project: **CrossPay Protocol**
*Tagline: "Verifiable, Private, Cross-Chain Payment Infrastructure"*

### Core Architecture Philosophy

```
THREE LAYERS OF ABSTRACTION:
1. Payment Core (shared logic - 60% of code)
2. Chain Adapters (network-specific - 30% of code)
3. Feature Modules (track-specific showcases - 10% of code)
```

### Master Project Structure

```
crosspay/
├── contracts/
│   ├── core/
│   │   ├── PaymentCore.sol          # Base escrow + streaming + splits
│   │   ├── ReceiptRegistry.sol      # Universal receipt storage
│   │   └── ComplianceBase.sol       # KYC/AML hooks
│   ├── privacy/
│   │   ├── ConfidentialPayments.sol # Zama FHE implementation
│   │   └── SelectiveDisclosure.sol  # Zama ACL management
│   ├── security/
│   │   ├── RelayValidator.sol       # Symbiotic relay verification
│   │   ├── TrancheVault.sol        # Symbiotic senior/junior
│   │   └── DVNAdapter.sol          # Symbiotic cross-chain
│   ├── oracles/
│   │   ├── FlareAdapter.sol        # FTSO + FDC + RNG
│   │   └── PriceFeeds.sol          # Multi-oracle abstraction
│   └── attestations/
│       ├── EASAdapter.sol          # ENS-linked attestations
│       └── SubnameRegistry.sol     # Auto-subnames for communities
│
├── services/
│   ├── payment-processor/          # Go - High performance core
│   │   ├── api/                   # REST/GraphQL endpoints
│   │   ├── chain/                 # Multi-chain coordinators
│   │   └── risk/                  # Fraud detection engine
│   ├── storage-worker/             # Go - Filecoin integration
│   │   └── synapse/               # SynapseSDK wrapper
│   ├── relay-network/              # Go - Symbiotic validators
│   │   ├── validator/             # Node implementation
│   │   └── aggregator/            # Signature aggregation
│   └── ai-copilot/                # TypeScript - ML services
│       ├── normalize/             # Invoice parsing
│       ├── risk/                  # Fraud scoring
│       └── bio/                   # Optional biotech module
│
├── app/                           # SvelteKit main interface
│   ├── src/
│   │   ├── routes/
│   │   │   ├── pay/              # Universal payment flow
│   │   │   ├── receipts/         # Receipt viewer/verifier
│   │   │   ├── grants/           # Zama private grants
│   │   │   └── analytics/        # Symbiotic dashboard
│   │   └── lib/
│   │       ├── components/
│   │       │   ├── ChainSelector.svelte
│   │       │   ├── PrivacyToggle.svelte
│   │       │   └── ENSResolver.svelte
│   │       └── stores/
│   │           ├── payment.ts
│   │           └── chain.ts
│
├── packages/                      # Shared libraries
│   ├── sdk/                      # TypeScript SDK
│   ├── types/                    # Shared type definitions
│   └── utils/                    # Common utilities
│
└── deployments/                   # Network configurations
    ├── production/
    └── testnet/
```

### Data Model (Universal Across All Chains)

```typescript
interface Payment {
  // Core identifiers
  id: string
  chainId: number
  txHash: string

  // Participants (ENS-resolved)
  sender: Address
  senderENS?: string
  recipient: Address
  recipientENS?: string

  // Payment details
  token: Address
  amount: bigint | EncryptedAmount  // Plain or Zama FHE

  // Privacy settings
  isPrivate: boolean
  disclosurePolicy?: ACLPolicy      // Zama selective disclosure

  // Verification
  attestationId?: string             // EAS attestation
  validatorProof?: AggregatedProof  // Symbiotic validation

  // Storage
  receiptCID?: string                // Filecoin IPFS hash
  metadata?: PaymentMetadata

  // Oracle data
  fxRate?: OraclePrice               // Flare FTSO
  randomSeed?: bytes32               // Flare RNG
  externalProof?: FDCProof          // Flare FDC

  // Status
  status: 'pending' | 'completed' | 'refunded'
  createdAt: number
  completedAt?: number
}

interface Receipt extends Payment {
  // Additional receipt-specific fields
  receiptId: string
  receiptType: 'payment' | 'grant' | 'remittance'
  attestation: Attestation
  complianceChecks?: ComplianceResult
}
```

### Service Architecture (Microservices Pattern)

```yaml
# docker-compose.yml structure
services:
  payment-processor:
    language: Go
    responsibilities:
      - Multi-chain transaction coordination
      - Risk assessment and fraud detection
      - Payment scheduling and streaming
      - WebSocket real-time updates

  storage-worker:
    language: Go
    responsibilities:
      - Filecoin/IPFS operations
      - Receipt generation (PDF/JSON)
      - CID management
      - Retrieval serving

  relay-network:
    language: Go
    responsibilities:
      - Symbiotic validator node
      - Signature aggregation
      - Proof generation
      - Slashing monitoring

  ai-copilot:
    language: TypeScript
    responsibilities:
      - Natural language processing
      - Risk scoring models
      - Compliance automation
      - Optional: Biotech features
```

### Chain Deployment Strategy

```typescript
const deploymentStrategy = {
  // Tier 1 - Maximum effort
  tier1: {
    symbiotic: {
      features: ['relay', 'vault', 'dvn', 'analytics'],
      effort: 'DEEP',
      priority: 'HIGH'
    },
    lisk: {
      features: ['core', 'attestations', 'ens'],
      effort: 'COMPLETE',
      priority: 'HIGH'
    },
    base: {
      features: ['mini-app', 'viral-mechanics'],
      effort: 'POLISHED',
      priority: 'HIGH'
    },
    zama: {
      features: ['full-privacy', 'selective-disclosure'],
      effort: 'DEEP',
      priority: 'HIGH'
    }
  },

  // Tier 2 - Solid implementation
  tier2: {
    flare: {
      features: ['ftso', 'fdc', 'rng'],
      effort: 'COMPLETE',
      priority: 'MEDIUM'
    },
    filecoin: {
      features: ['storage', 'retrieval'],
      effort: 'STANDARD',
      priority: 'MEDIUM'
    },
    ai: {
      features: ['copilot', 'risk', 'bio-optional'],
      effort: 'STANDARD',
      priority: 'MEDIUM'
    }
  },

  // Tier 3 - Quick deploys
  tier3: {
    ens: {
      features: ['resolution', 'subnames'],
      effort: 'BASIC',
      priority: 'LOW'
    },
    citrea: {
      features: ['btc-escrow'],
      effort: 'MINIMAL',
      priority: 'LOW'
    }
  }
}
```

### Critical Success Metrics

```typescript
const successMetrics = {
  technical: {
    symbiotic: 'All 3 components implemented',
    zama: 'Full encryption with ACL',
    base: 'Virality score > 15/20',
    lisk: 'Low-cost transactions'
  },

  business: {
    totalAddressableMarket: 'Global remittances market',
    uniqueValueProp: 'Private + Verifiable + Cross-chain',
    competitiveAdvantage: 'Fintech expertise + Full stack'
  },

  demo: {
    coreFlow: '< 30 seconds end-to-end',
    perTrack: '60-90 second specific demo',
    errorHandling: 'Graceful failures',
    liveData: 'Real testnet transactions'
  }
}
```

### Risk Mitigation Architecture

```typescript
const riskMitigation = {
  timeRisk: {
    mitigation: 'Core first, features second',
    fallback: 'Ship Tier 1 only if needed'
  },

  technicalRisk: {
    zama: 'Pre-build encryption tests',
    symbiotic: 'Static validator set for demo',
    flare: 'Cache oracle responses'
  },

  integrationRisk: {
    mitigation: 'Adapter pattern for all chains',
    testing: 'Unit tests per adapter',
    monitoring: 'Health checks per service'
  }
}
```

### Documentation Structure

```markdown
docs/
├── README.md                 # Master documentation
├── TRACKS.md                # Track-specific guides
├── ARCHITECTURE.md          # This document
├── DEPLOYMENT.md            # How to deploy
├── DEMO_SCRIPTS.md          # 90-second demos
└── tracks/
    ├── symbiotic.md         # Deep dive
    ├── zama.md              # Privacy architecture
    ├── base.md              # Mini app guide
    └── [other tracks].md
```

### Smart Contract Interface Hierarchy

```solidity
// Base interface all chains implement
interface IPaymentCore {
    function createPayment(...) returns (uint256);
    function completePayment(uint256 id);
    function getReceipt(uint256 id) returns (Receipt);
}

// Track-specific extensions
interface IConfidentialPayment extends IPaymentCore {
    function createPrivatePayment(euint256 amount...);
    function grantDisclosure(uint256 id, address viewer);
}

interface IValidatedPayment extends IPaymentCore {
    function completeWithProof(uint256 id, bytes proof);
}

interface IOraclePayment extends IPaymentCore {
    function createWithOracle(uint256 amount, bytes32 feed);
}
```

### State Management Strategy

```typescript
// Unified state across all UIs
const PaymentState = {
  // Global state
  activeChain: ChainId,
  activeAccount: Address,

  // Feature flags
  features: {
    privacy: boolean,      // Zama on/off
    validation: boolean,   // Symbiotic on/off
    oracles: boolean,      // Flare on/off
    storage: boolean,      // Filecoin on/off
  },

  // Payment flow
  currentPayment: Payment,
  receipts: Receipt[],

  // Track-specific
  symbiotic: {
    validators: Validator[],
    vaultPositions: Position[]
  },
  zama: {
    encryptedBalance: EncryptedValue,
    disclosures: Disclosure[]
  }
}
```

This architecture ensures:
1. **Maximum code reuse** across tracks
2. **Clear separation** of concerns
3. **Easy testing** per component
4. **Quick pivots** if time runs short
5. **Professional presentation** to judges

The key is that every track sees a complete, polished implementation of their requirements, while you're actually building one coherent system.
