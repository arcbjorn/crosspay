# CrossPay Protocol - Technical Architecture

**Verifiable, Private, Cross-Chain Payment Infrastructure**

CrossPay is a unified payment system that integrates multiple blockchain networks to provide comprehensive payment capabilities including privacy, security validation, permanent storage, and oracle integration.

## Core Architecture Philosophy

CrossPay follows a **three-layer modular architecture**:

1. **Payment Core** (60% shared logic) - Universal payment logic and state management
2. **Chain Adapters** (30% network-specific) - Blockchain-specific implementations  
3. **Feature Modules** (10% specialized) - Network-specific capabilities and extensions

## Network Integrations

CrossPay integrates with 8+ blockchain networks, each providing specialized capabilities:

### Tier 1 Networks (Core Features)
- **Symbiotic**: Security layer with validator network and tranche vault system
- **Zama**: Privacy layer with FHE encryption and selective disclosure  
- **Lisk**: Primary deployment for low-cost payments with ENS and EAS integration
- **Base**: Consumer mini-app with viral sharing and USDC payments

### Tier 2 Networks (Enhanced Features)
- **Flare**: Oracle services (FTSO pricing, FDC proofs, secure RNG)
- **Filecoin**: Permanent receipt storage via SynapseSDK
- **ENS**: Universal name resolution and subname registry
- **Citrea**: Bitcoin-denominated escrow and settlement

## Component Architecture

### Smart Contracts
- **Payment Core**: Universal escrow and payment logic
- **Privacy Module**: Zama FHE confidential payments 
- **Security Module**: Symbiotic validator integration
- **Oracle Module**: Flare data feed adapters
- **Storage Module**: Filecoin receipt management
- **Attestation Module**: EAS and ENS integration

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

### Service Architecture

**Core Services:**
- **Payment Processor**: Multi-chain transaction coordination and risk assessment
- **Storage Worker**: Filecoin/IPFS operations and receipt management
- **Relay Network**: Symbiotic validator nodes and proof aggregation

**User Interfaces:**
- **Web Application**: SvelteKit frontend with universal payment flows
- **Base Mini App**: Mobile-optimized viral payment experience
- **Developer SDK**: TypeScript SDK for third-party integration


## Key Capabilities

### Privacy & Confidentiality
- Zama FHE encryption for confidential payment amounts
- Selective disclosure with role-based access control
- Privacy-preserving grant and payment systems

### Security & Validation  
- Symbiotic validator network for payment verification
- Tranche vault system with risk-stratified positions
- Cross-chain message validation via DVN adapters

### Permanent Storage
- Filecoin integration for durable receipt storage
- Content-addressed retrieval via IPFS CIDs
- Automated storage management and renewal

### Oracle Integration
- Flare FTSO for real-time price feeds
- FDC for external proof verification
- Secure randomness for fair selection processes

### User Experience
- ENS name resolution for human-readable addresses
- EAS attestations for verifiable payment receipts
- Mobile-first mini-app with viral sharing mechanics

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

## Universal Payment Data Model

### Core Payment Interface
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
  status: 'pending' | 'validated' | 'completed' | 'refunded'
  createdAt: number
  completedAt?: number
}
```

### Extended Receipt Interface
```typescript
interface Receipt extends Payment {
  receiptId: string
  receiptType: 'payment' | 'grant' | 'remittance'
  attestation: EASAttestation
  complianceChecks?: ComplianceResult
  permanentStorageProof: FilecoinProof
}
```

## Core System Flows

### 1. Standard Public Payment Flow
```
1. ENS Resolution     → Resolve recipient name to address
2. File Attachment    → Optional document upload → CID
3. Oracle Price Lock  → Flare FTSO snapshot for FX protection
4. Escrow Creation    → Funds locked in chain-specific contract
5. Receipt Generation → EAS attestation + Filecoin storage
6. Payment Release    → Direct or validator-gated release
7. Confirmation       → Explorer links, receipt sharing
```

### 2. Confidential Payment Flow (Zama FHE)
```
1. Privacy Toggle     → User enables confidential mode
2. Client Encryption  → Amount encrypted per Zama protocol
3. Confidential Escrow → Encrypted amount stored on-chain
4. Selective Disclosure → Role-based decryption permissions
5. Private Release    → Amount remains hidden unless disclosed
6. Compliance Access → Auditors can request controlled reveal
```

### 3. Validator-Secured Payment Flow (Symbiotic)
```
1. High-Value Detection → Payment exceeds security threshold
2. Validator Request   → Relay network receives proof request
3. Signature Collection → BFT consensus among validators
4. Proof Aggregation   → Signatures combined into single proof
5. Gated Release      → Smart contract verifies proof before release
6. Tranche Updates    → Vault positions reflect security activity
```

### 4. Oracle-Enhanced Payment Flow (Flare)
```
1. Price Feed Query   → FTSO provides current FX rate
2. Rate Snapshot      → Price locked at payment creation
3. External Verification → Optional FDC proof for confirmations
4. Slippage Protection → Payment adjusts based on rate changes
5. RNG Integration    → Random selection for grants/features
```

## Smart Contract Architecture

### Interface Hierarchy

**Base Interface - IPaymentCore:**
```solidity
interface IPaymentCore {
    // Core payment operations
    function createPayment(
        address recipient, 
        uint256 amount, 
        bytes32 metadata
    ) external returns (uint256 paymentId);
    
    function completePayment(uint256 paymentId) external;
    function refundPayment(uint256 paymentId) external;
    function getPaymentStatus(uint256 paymentId) external view returns (PaymentStatus);
    function getReceipt(uint256 paymentId) external view returns (Receipt memory);
    
    // Events
    event PaymentCreated(uint256 indexed paymentId, address indexed sender, address indexed recipient);
    event PaymentCompleted(uint256 indexed paymentId);
    event PaymentRefunded(uint256 indexed paymentId);
}
```

**Privacy Extension - IConfidentialPayment:**
```solidity
interface IConfidentialPayment is IPaymentCore {
    function createPrivatePayment(
        address recipient,
        euint256 encryptedAmount,
        bytes calldata aclPolicy
    ) external returns (uint256 paymentId);
    
    function grantDisclosure(
        uint256 paymentId, 
        address viewer, 
        bytes calldata proof
    ) external;
    
    function revokeDisclosure(uint256 paymentId, address viewer) external;
    
    event DisclosureGranted(uint256 indexed paymentId, address indexed viewer);
    event DisclosureRevoked(uint256 indexed paymentId, address indexed viewer);
}
```

**Validation Extension - IValidatedPayment:**
```solidity
interface IValidatedPayment is IPaymentCore {
    function completeWithProof(
        uint256 paymentId,
        bytes calldata aggregatedSignature,
        uint256 validatorBitmap
    ) external;
    
    function setValidatorSet(
        address[] calldata validators,
        uint256 threshold
    ) external;
    
    event ValidationRequired(uint256 indexed paymentId);
    event ProofVerified(uint256 indexed paymentId, bytes32 proofHash);
}
```

**Oracle Extension - IOraclePayment:**
```solidity
interface IOraclePayment is IPaymentCore {
    function createWithOracle(
        address recipient,
        uint256 amount,
        bytes32 feedId,
        uint256 maxSlippage
    ) external returns (uint256 paymentId);
    
    function attachExternalProof(
        uint256 paymentId,
        bytes32 merkleRoot,
        bytes32 leaf,
        bytes32[] calldata proof
    ) external;
    
    event OraclePriceUsed(uint256 indexed paymentId, bytes32 feedId, uint256 price);
    event ExternalProofAttached(uint256 indexed paymentId, bytes32 root, bytes32 leaf);
}
```

## Network-Specific Integrations

### Symbiotic Security Layer

**Relay Validator Network:**
- Minimum 3 validator nodes for BFT consensus
- Domain-separated signature scheme prevents replay attacks
- Aggregated proof verification reduces gas costs
- Slashing conditions for validator misbehavior

**Tranche Vault System:**
- Senior tranche: Low risk, low yield, last to be slashed
- Junior tranche: High risk, high yield, first to absorb losses
- Automatic rebalancing based on risk assessment
- Real-time TVL and position tracking

**Cross-Chain Validation (DVN):**
- Validator signatures verify cross-chain messages
- Same validator set secures both payments and messages
- Merkle proof verification for message authenticity

### Zama Privacy Layer

**FHE Encryption Architecture:**
- Client-side encryption before blockchain submission
- Homomorphic operations preserve privacy on-chain
- Selective decryption through Access Control Lists (ACL)
- Role-based disclosure for compliance and auditing

**Privacy Modes:**
1. **Public**: Standard transparent payments
2. **Private**: Encrypted amounts, visible participants
3. **Shielded**: Encrypted amounts and participants
4. **Compliant**: Private with regulatory disclosure capabilities

**Encryption Operations:**
```solidity
// Encrypted arithmetic operations
euint256 encryptedAmount = TFHE.asEuint256(amount);
euint256 fee = TFHE.mul(encryptedAmount, feeRate);
euint256 netAmount = TFHE.sub(encryptedAmount, fee);

// Conditional transfers based on encrypted conditions
TFHE.cmux(condition, recipient1, recipient2);
```

### Flare Oracle Integration

**FTSO (Flare Time Series Oracle):**
- Real-time price feeds updated every 3 minutes
- Median aggregation from multiple data providers
- Price feed verification and staleness protection
- Support for 100+ cryptocurrency and FX pairs

**FDC (Flare Data Connector):**
- Merkle proof verification of external data
- Web2 API attestation with cryptographic proofs
- Payment confirmation verification from external systems
- Verifiable randomness for fair selection processes

**Implementation Example:**
```solidity
// FTSO price query
(uint256 price, uint256 timestamp, uint256 decimals) = ftsoRegistry.getCurrentPriceWithDecimals(feedId);
require(block.timestamp - timestamp < MAX_PRICE_AGE, "Price too stale");

// FDC proof verification
bool isValid = merkleTree.verify(leaf, proof, attestedRoot);
require(isValid, "Invalid external proof");
```

### Filecoin Storage Layer

**SynapseSDK Integration:**
- Automated receipt upload to Filecoin network
- Content addressing through IPFS CID generation
- Deal status monitoring and renewal management
- Retrieval optimization with multiple storage providers

**Storage Architecture:**
```typescript
interface StorageService {
  uploadReceipt(receipt: Receipt): Promise<string>; // Returns CID
  retrieveReceipt(cid: string): Promise<Receipt>;
  verifyStorage(cid: string): Promise<StorageProof>;
  renewStorage(cid: string): Promise<DealStatus>;
}
```

## Service Coordination

### Payment Processor Service

**Multi-Chain Coordination:**
- Unified nonce management across chains
- Gas optimization and transaction batching
- Automatic retry logic with exponential backoff
- Real-time transaction status monitoring

**Risk Assessment Engine:**
- Transaction pattern analysis for fraud detection
- Compliance screening against sanctioned addresses
- Amount-based risk scoring and validation requirements
- Integration with external compliance services

### Storage Worker Service

**Receipt Management:**
- Automated PDF and JSON receipt generation
- Template-based receipt formatting for different use cases
- CID management and metadata indexing
- Backup redundancy across multiple storage providers

**Retrieval Optimization:**
- Content caching for frequently accessed receipts
- Range request support for large documents
- CDN integration for global distribution
- Fallback mechanisms for storage provider outages

### Relay Network Service

**Validator Node Implementation:**
- BLS signature scheme for efficient aggregation
- Threshold signature generation with configurable quorum
- Peer-to-peer communication for consensus coordination
- Automatic failover and recovery mechanisms

**Signature Aggregation:**
- Batch processing of multiple payment proofs
- Optimized verification to reduce gas costs
- Proof caching to prevent duplicate work
- Rate limiting and DDoS protection

## Cross-Chain Architecture

### Chain Abstraction Layer

**Unified Interface:**
```typescript
interface ChainAdapter {
  chainId: number;
  rpcUrl: string;
  explorerUrl: string;
  
  // Core operations
  createPayment(params: PaymentParams): Promise<TransactionResult>;
  getPaymentStatus(txHash: string): Promise<PaymentStatus>;
  estimateGas(operation: Operation): Promise<GasEstimate>;
  
  // Chain-specific features
  supportsFeature(feature: Feature): boolean;
  getFeatureConfig(feature: Feature): FeatureConfig;
}
```

**Network Configuration:**
- Chain-specific RPC endpoints and fallbacks
- Explorer integration for transaction deep-linking
- Faucet information for testnet operations
- Contract addresses and deployment verification

### State Synchronization

**Event Processing:**
- Real-time blockchain event monitoring
- Cross-chain state reconciliation
- Optimistic updates with rollback capabilities
- Event deduplication and ordering guarantees

**Data Consistency:**
- Eventually consistent cross-chain state
- Conflict resolution for concurrent updates
- Audit trails for all state changes
- Backup and recovery mechanisms

## Security Model

### Transaction Security

**Smart Contract Protections:**
- Reentrancy guards on all external calls
- Integer overflow/underflow protection
- Access control modifiers for privileged functions
- Emergency pause functionality for critical issues

**Validation Mechanisms:**
- Multi-signature requirements for high-value payments
- Time-lock delays for sensitive operations
- Address whitelist/blacklist capabilities
- Slippage protection for oracle-based pricing

### Privacy Protections

**FHE Security Model:**
- Client-side encryption with user-controlled keys
- Zero-knowledge proofs for balance verification
- Selective disclosure with audit trails
- Side-channel attack mitigation

**Metadata Protection:**
- Minimal on-chain metadata exposure
- Off-chain storage of sensitive information
- Encrypted communication channels
- Regular security audits and assessments

## API Interface

### Core Endpoints
- `POST /api/pay` - Process cross-chain payment
- `POST /api/validate` - Request validation proof  
- `POST /api/storage/upload` - Store receipt to permanent storage
- `GET /api/receipt/{id}` - Retrieve payment receipt
- `GET /api/status/{txHash}` - Get transaction status

### Feature-Specific Endpoints
- `POST /api/privacy/disclose` - Grant selective disclosure access
- `GET /api/oracle/price/{feedId}` - Get current price feeds
- `POST /api/attestation/create` - Generate payment attestation
- `POST /api/validator/proof` - Request validator proof
- `GET /api/storage/cid/{cid}` - Retrieve by content identifier

## Development & Integration

### SDK Integration

**TypeScript SDK:**
```typescript
import { CrossPaySDK } from '@crosspay/sdk';

const crossPay = new CrossPaySDK({
  apiKey: 'your-api-key',
  network: 'testnet' // or 'mainnet'
});

// Create a payment
const payment = await crossPay.createPayment({
  recipient: 'recipient.eth',
  amount: '100',
  token: 'USDC',
  chainId: 1135, // Lisk
  features: {
    privacy: false,
    validation: true,
    storage: true
  }
});

// Monitor payment status
const status = await crossPay.getPaymentStatus(payment.id);
```

### Testing Framework

**Smart Contract Testing:**
- Foundry test suites for each contract module
- Property-based testing for invariant verification
- Integration tests across multiple chains
- Gas optimization benchmarks

**Service Testing:**
- Unit tests for individual service components
- Integration tests for service coordination
- Load testing for high-throughput scenarios
- End-to-end testing for complete payment flows

### Deployment Architecture

**Contract Deployment:**
- Deterministic deployment across all supported chains
- Upgrade proxy pattern for iterative improvements
- Multi-signature deployment for production networks
- Automated verification and contract registration

**Service Deployment:**
- Containerized microservices with Docker
- Kubernetes orchestration for scalability
- Load balancing and automatic failover
- Monitoring and alerting integration

## Monitoring & Observability

### Metrics & Analytics

**System Metrics:**
- Payment volume and transaction counts
- Cross-chain success rates and latency
- Validator network performance
- Storage utilization and costs

**Business Metrics:**
- User adoption and retention rates
- Feature utilization across networks
- Revenue attribution by integration
- Compliance audit trail completeness

### Logging & Tracing

**Structured Logging:**
- Payment lifecycle event tracking
- Cross-service correlation identifiers
- Error classification and resolution tracking
- Performance profiling and optimization

**Distributed Tracing:**
- End-to-end payment flow visibility
- Cross-chain operation coordination
- Service dependency mapping
- Bottleneck identification and resolution

This comprehensive architecture enables CrossPay to serve as a unified payment infrastructure while showcasing the unique capabilities of each integrated blockchain network. The modular design ensures maximum code reuse, clear separation of concerns, and the flexibility to extend support to additional networks and features as the ecosystem evolves.
