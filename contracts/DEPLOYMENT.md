# Contract Deployment Guide

## Prerequisites

1. **Environment Setup**
   ```bash
   cp ../.env.example ../.env
   # Edit .env with your private key and API keys
   ```

2. **Required Funds**
   - Lisk Sepolia ETH: Get from https://sepolia-faucet.lisk.com
   - Base Sepolia ETH: Get from https://www.coinbase.com/faucets/base-ethereum-sepolia-faucet

## Deployment Commands

### Deploy to Lisk Sepolia (Chain ID: 4202)
```bash
pnpm deploy:lisk
# or
forge script script/Deploy.s.sol:Deploy --rpc-url lisk-sepolia --broadcast --verify
```

### Deploy to Base Sepolia (Chain ID: 84532)  
```bash
pnpm deploy:base
# or
forge script script/Deploy.s.sol:Deploy --rpc-url base-sepolia --broadcast --verify
```

## Post-Deployment Steps

1. **Verify Deployment Files**
   - Check `../deployments/4202.json` (Lisk Sepolia)
   - Check `../deployments/84532.json` (Base Sepolia)

2. **Update Frontend Configuration**
   ```bash
   # Copy contract addresses to app/.env
   VITE_LISK_PAYMENT_CORE=0x...
   VITE_LISK_RECEIPT_REGISTRY=0x...
   VITE_BASE_PAYMENT_CORE=0x...
   VITE_BASE_RECEIPT_REGISTRY=0x...
   ```

3. **Update Contract Integration**
   - Edit `app/src/lib/contracts/index.ts`
   - Replace placeholder addresses with deployed addresses

## Verification

After deployment, verify contracts are working:

```bash
# Check contract on Lisk Sepolia
forge verify-contract <address> src/PaymentCore.sol:PaymentCore --chain 4202

# Check contract on Base Sepolia  
forge verify-contract <address> src/PaymentCore.sol:PaymentCore --chain 84532
```

## Troubleshooting

- **Insufficient funds**: Get testnet ETH from faucets
- **Gas estimation failed**: Check network connectivity
- **Verification failed**: Ensure API keys are correct
- **Nonce too high**: Reset nonce in MetaMask

## Security Notes

- Never commit private keys to git
- Use a dedicated testnet wallet
- Verify contract addresses before frontend integration