import { writable } from 'svelte/store';
import {
	connect,
	disconnect,
	switchChain as wagmiSwitchChain,
	getAccount,
	getChainId,
	getBalance
} from 'wagmi/actions';
import { walletConnect, metaMask } from 'wagmi/connectors';
import { wagmiConfig } from '$lib/wagmi';
import type { Address } from 'viem';

export interface WalletState {
	isConnected: boolean;
	address?: Address;
	chainId?: number;
	balance?: bigint;
	isConnecting: boolean;
	error?: string;
}

const initialState: WalletState = {
	isConnected: false,
	isConnecting: false
};

export const walletStore = writable<WalletState>(initialState);

// Update store with current wagmi state
function updateStoreFromWagmi() {
	const account = getAccount(wagmiConfig);
	const chainId = getChainId(wagmiConfig);

	walletStore.update((state) => ({
		...state,
		isConnected: account.isConnected,
		address: account.address,
		chainId: chainId,
		isConnecting: false
	}));

	// Get balance if connected
	if (account.address) {
		getBalance(wagmiConfig, { address: account.address })
			.then((balance) => {
				walletStore.update((state) => ({ ...state, balance: balance.value }));
			})
			.catch(console.error);
	}
}

export const connectWallet = async (preferredConnector?: 'metamask' | 'walletconnect') => {
	walletStore.update((state) => ({ ...state, isConnecting: true, error: undefined }));

	try {
		let connector;

		if (preferredConnector === 'walletconnect') {
			connector = walletConnect({
				projectId: import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || 'demo-project-id',
				metadata: {
					name: 'CrossPay Protocol',
					description: 'Verifiable, private, cross-chain payments',
					url: 'https://crosspay.protocol',
					icons: ['https://crosspay.protocol/favicon.svg']
				}
			});
		} else {
			// Default to MetaMask
			connector = metaMask();
		}

		await connect(wagmiConfig, { connector });
		updateStoreFromWagmi();

		// Auto-switch to Lisk Sepolia (4202) after connection
		const account = getAccount(wagmiConfig);
		const currentChainId = getChainId(wagmiConfig);

		if (account.isConnected && ![4202, 84532, 5115].includes(currentChainId)) {
			try {
				console.log('Auto-switching to Lisk Sepolia...');
				await wagmiSwitchChain(wagmiConfig, { chainId: 4202 });
				updateStoreFromWagmi();
			} catch (error) {
				console.log('Auto chain-switch failed, user can manually switch:', error);
			}
		}
	} catch (error) {
		console.error('Wallet connection error:', error);
		walletStore.update((state) => ({
			...state,
			isConnecting: false,
			error: error instanceof Error ? error.message : 'Failed to connect wallet'
		}));
	}
};

export const disconnectWallet = async () => {
	try {
		await disconnect(wagmiConfig);
		walletStore.set(initialState);
	} catch (error) {
		console.error('Disconnect error:', error);
		walletStore.set(initialState);
	}
};

export const switchChain = async (chainId: number) => {
	try {
		await wagmiSwitchChain(wagmiConfig, { chainId: chainId as any });
		updateStoreFromWagmi();
	} catch (error: any) {
		console.error('Chain switch error:', error);
		throw error;
	}
};

// Initialize store on load
if (typeof window !== 'undefined') {
	updateStoreFromWagmi();
}

// Export function to manually update store (called by wagmi provider)
export { updateStoreFromWagmi };
