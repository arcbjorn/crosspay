import { createConfig, http } from 'wagmi';
import { liskSepolia, baseSepolia } from 'wagmi/chains';
import { walletConnect, metaMask, safe } from 'wagmi/connectors';

// Get WalletConnect project ID from environment
const projectId = import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || 'demo-project-id';

export const wagmiConfig = createConfig({
  chains: [liskSepolia, baseSepolia],
  connectors: [
    walletConnect({ 
      projectId,
      metadata: {
        name: 'CrossPay Protocol',
        description: 'Verifiable, private, cross-chain payments',
        url: 'https://crosspay.protocol',
        icons: ['https://crosspay.protocol/favicon.svg']
      }
    }),
    metaMask(),
    safe(),
  ],
  transports: {
    [liskSepolia.id]: http('https://rpc.sepolia-api.lisk.com'),
    [baseSepolia.id]: http('https://sepolia.base.org'),
  },
});