# CrossPay Protocol Deployment Guide

## Prerequisites

1. **Testnet ETH**: Get Lisk Sepolia ETH from the faucet: https://sepolia-faucet.lisk.com
2. **Private Key**: Have a wallet private key with testnet ETH

## Environment Setup

1. Copy the environment template:
```bash
cd contracts
cp .env.example .env
```

2. Edit `.env` and add your private key (without 0x prefix):
```bash
PRIVATE_KEY=your_64_character_private_key_here
```

## Deploy to Lisk Sepolia

```bash
cd contracts
forge script script/Deploy.s.sol --rpc-url lisk-sepolia --broadcast --verify
```

## Deploy to Base Sepolia

```bash
cd contracts  
forge script script/Deploy.s.sol --rpc-url base-sepolia --broadcast --verify
```

## Verify Deployment

After deployment, contract addresses will be saved to:
- `deployments/4202.json` (Lisk Sepolia)
- `deployments/84532.json` (Base Sepolia)

Update the frontend contract addresses in `packages/types/contracts.ts`.

## Post-Deployment

1. Update contract addresses in frontend
2. Test payment creation flow
3. Verify on block explorers:
   - Lisk: https://sepolia-blockscout.lisk.com
   - Base: https://sepolia.basescan.org