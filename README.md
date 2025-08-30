# CrossPay Protocol

Verifiable, private, cross-chain payment infrastructure targeting 9 blockchain networks with unified architecture.

## Overview

CrossPay enables secure payments across multiple chains with privacy preservation (Zama FHE), validator security (Symbiotic), permanent storage (Filecoin), and oracle integration (Flare).

## Quick Start

```bash
pnpm install
cd contracts && forge deploy
cd ../app && pnpm dev
docker-compose up  # Optional services
```

## Architecture

- **Contracts**: Core payment logic, privacy (FHE), security (validators), oracles, attestations
- **Services**: Payment processor (Go), storage worker (Go), relay network (Go), AI copilot (TS)
- **Frontend**: SvelteKit app with chain-specific routes
- **Mini App**: Base MiniKit for mobile payments

## Network Deployments

| Network | Purpose | Features |
|---------|---------|----------|
| Symbiotic | Security layer | Relay validators, tranche vault, DVN |
| Lisk | Main deployment | Core payments, ENS, attestations |
| Base | Consumer app | Mini app, viral sharing |
| Zama | Privacy | FHE encryption, selective disclosure |
| Flare | Oracles | FTSO pricing, FDC proofs, RNG |
| Filecoin | Storage | Receipt storage via IPFS |
| Citrea | Bitcoin | BTC-denominated escrow |

## Core Interfaces

```solidity
interface IPaymentCore {
    function createPayment(address recipient, uint256 amount, bytes32 metadata) returns (uint256);
    function completePayment(uint256 id);
}

interface IConfidentialPayment extends IPaymentCore {
    function createPrivatePayment(address recipient, euint256 amount);
    function grantDisclosure(uint256 id, address viewer);
}
```

## API Endpoints

- `POST /api/pay` - Process payment
- `POST /api/validate` - Request validation
- `POST /api/storage/upload` - Store receipt
- `POST /api/ai/risk` - Risk scoring

## Documentation

- [Architecture Details](./docs/ARCHITECTURE.md)
- [Deployment Guide](./docs/DEPLOYMENT.md)
- [Track Implementations](./docs/DoD.md)

## License

MIT
