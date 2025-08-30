# Deployment Guide

This guide summarizes how to deploy contracts and run the stack for demos. Commands assume Foundry, pnpm, and Go are installed.

## Prerequisites
- Foundry (`forge`, `cast`) installed and configured
- RPC URLs and private keys in `.env` (never commit secrets)
- pnpm installed; Go 1.21+

## Contracts
- Build and test:
  ```bash
  cd contracts
  forge build
  forge test -vvv
  ```
- Deploy (use the project’s deploy script if present):
  ```bash
  # If a convenience alias/script exists
  forge deploy

  # Otherwise, run a script explicitly (example names)
  forge script script/Deploy.s.sol:Deploy \
    --rpc-url $RPC_URL --broadcast --verify -vvvv
  ```
- Networks (examples):
  - Lisk: core escrow + attestations
  - Base: mini app integration
  - Zama/FHE: privacy contracts on supported FHE testnet

## Services (optional)
- Start any Go services:
  ```bash
  cd services/payment-processor && go build ./... && ./payment-processor
  # Repeat for storage-worker and relay-network as needed
  ```
- Compose (if file exists):
  ```bash
  docker-compose up -d
  ```

## Frontend
```bash
cd app
pnpm install
pnpm dev
```

## Verification Checklist
- Addresses and chain IDs recorded in README or `.env.example`
- Contract events include CIDs (Filecoin) where applicable
- Health checks pass for services; basic payment flow works end-to-end
