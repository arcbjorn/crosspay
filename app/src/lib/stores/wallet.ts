import { writable } from 'svelte/store';
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
  isConnecting: false,
};

export const walletStore = writable<WalletState>(initialState);

export const connectWallet = async () => {
  walletStore.update(state => ({ ...state, isConnecting: true, error: undefined }));
  
  try {
    if (!window.ethereum) {
      throw new Error('No wallet detected. Please install MetaMask or another Ethereum wallet.');
    }

    const accounts = await window.ethereum.request({
      method: 'eth_requestAccounts'
    });

    if (accounts.length === 0) {
      throw new Error('No accounts found');
    }

    const chainId = await window.ethereum.request({
      method: 'eth_chainId'
    });

    const balance = await window.ethereum.request({
      method: 'eth_getBalance',
      params: [accounts[0], 'latest']
    });

    walletStore.update(state => ({
      ...state,
      isConnected: true,
      address: accounts[0] as Address,
      chainId: parseInt(chainId, 16),
      balance: BigInt(balance),
      isConnecting: false,
      error: undefined
    }));

  } catch (error) {
    console.error('Wallet connection error:', error);
    walletStore.update(state => ({
      ...state,
      isConnecting: false,
      error: error instanceof Error ? error.message : 'Failed to connect wallet'
    }));
  }
};

export const disconnectWallet = () => {
  walletStore.set(initialState);
};

export const switchChain = async (chainId: number) => {
  try {
    await window.ethereum.request({
      method: 'wallet_switchEthereumChain',
      params: [{ chainId: `0x${chainId.toString(16)}` }],
    });
  } catch (error: any) {
    if (error.code === 4902) {
      // Chain not added, need to add it first
      throw new Error(`Chain ${chainId} not found in wallet. Please add it manually.`);
    }
    throw error;
  }
};

// Listen for account changes
if (typeof window !== 'undefined' && window.ethereum) {
  window.ethereum.on('accountsChanged', (accounts: string[]) => {
    if (accounts.length === 0) {
      disconnectWallet();
    } else {
      walletStore.update(state => ({
        ...state,
        address: accounts[0] as Address
      }));
    }
  });

  window.ethereum.on('chainChanged', (chainId: string) => {
    walletStore.update(state => ({
      ...state,
      chainId: parseInt(chainId, 16)
    }));
  });
}

// Type augmentation for window.ethereum
declare global {
  interface Window {
    ethereum?: {
      request: (args: { method: string; params?: any[] }) => Promise<any>;
      on: (event: string, callback: (...args: any[]) => void) => void;
      removeListener: (event: string, callback: (...args: any[]) => void) => void;
    };
  }
}