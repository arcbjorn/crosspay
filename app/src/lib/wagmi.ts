import { createConfig, http } from 'wagmi';
import { liskSepolia, baseSepolia } from 'wagmi/chains';
import { defineChain } from 'viem';
import { walletConnect, metaMask, safe } from 'wagmi/connectors';

// Get WalletConnect project ID from environment
const projectId = import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || 'demo-project-id';

// Define Citrea Testnet chain
export const citreaTestnet = defineChain({
	id: 5115,
	name: 'Citrea Testnet',
	nativeCurrency: {
		name: 'Citrea Bitcoin',
		symbol: 'cBTC',
		decimals: 8
	},
	rpcUrls: {
		default: { http: ['https://rpc.testnet.citrea.xyz'] }
	},
	blockExplorers: {
		default: {
			name: 'Citrea Testnet Explorer',
			url: 'https://explorer.testnet.citrea.xyz'
		}
	},
	testnet: true
});

export const wagmiConfig = createConfig({
	chains: [liskSepolia, baseSepolia, citreaTestnet],
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
		safe()
	],
	transports: {
		[liskSepolia.id]: http('https://rpc.sepolia-api.lisk.com'),
		[baseSepolia.id]: http('https://sepolia.base.org'),
		[citreaTestnet.id]: http('https://rpc.testnet.citrea.xyz')
	}
});
