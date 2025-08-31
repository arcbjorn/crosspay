import { writable } from 'svelte/store';

export interface ChainConfig {
  id: number;
  name: string;
  nativeCurrency: {
    name: string;
    symbol: string;
    decimals: number;
  };
  rpcUrls: string[];
  blockExplorers: {
    name: string;
    url: string;
  }[];
  testnet: boolean;
  faucets?: string[];
}

export const SUPPORTED_CHAINS: Record<number, ChainConfig> = {
  4202: {
    id: 4202,
    name: 'Lisk Sepolia',
    nativeCurrency: {
      name: 'Sepolia Ether',
      symbol: 'ETH',
      decimals: 18,
    },
    rpcUrls: ['https://rpc.sepolia-api.lisk.com'],
    blockExplorers: [
      {
        name: 'Lisk Sepolia Explorer',
        url: 'https://sepolia-blockscout.lisk.com',
      },
    ],
    testnet: true,
    faucets: ['https://sepolia-faucet.lisk.com'],
  },
  84532: {
    id: 84532,
    name: 'Base Sepolia',
    nativeCurrency: {
      name: 'Sepolia Ether',
      symbol: 'ETH',
      decimals: 18,
    },
    rpcUrls: ['https://sepolia.base.org'],
    blockExplorers: [
      {
        name: 'Base Sepolia Explorer',
        url: 'https://sepolia-explorer.base.org',
      },
    ],
    testnet: true,
    faucets: ['https://www.coinbase.com/faucets/base-ethereum-sepolia-faucet'],
  },
};

export const DEFAULT_CHAIN_ID = 4202; // Lisk Sepolia

export const chainStore = writable<ChainConfig>(SUPPORTED_CHAINS[DEFAULT_CHAIN_ID]);

export const setChain = (chainId: number) => {
  const chain = SUPPORTED_CHAINS[chainId];
  if (chain) {
    chainStore.set(chain);
  }
};

export const getChain = (chainId: number): ChainConfig | undefined => {
  return SUPPORTED_CHAINS[chainId];
};

export const isChainSupported = (chainId: number): boolean => {
  return chainId in SUPPORTED_CHAINS;
};