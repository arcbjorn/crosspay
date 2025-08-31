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
  blockTime: number; // Average block time in seconds
  gasPrice: {
    average: string; // In gwei
    fast: string;
    standard: string;
  };
  confirmations: number; // Required confirmations
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
    blockTime: 2, // 2 second blocks
    gasPrice: {
      average: '0.1',
      fast: '0.2', 
      standard: '0.1'
    },
    confirmations: 6,
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
    blockTime: 2, // 2 second blocks
    gasPrice: {
      average: '0.1',
      fast: '0.15',
      standard: '0.1'
    },
    confirmations: 6,
  },
  5115: {
    id: 5115,
    name: 'Citrea Testnet',
    nativeCurrency: {
      name: 'Citrea Bitcoin',
      symbol: 'cBTC',
      decimals: 8,
    },
    rpcUrls: ['https://rpc.testnet.citrea.xyz'],
    blockExplorers: [
      {
        name: 'Citrea Testnet Explorer',
        url: 'https://explorer.testnet.citrea.xyz',
      },
    ],
    testnet: true,
    faucets: ['https://citrea.xyz/faucet'],
    blockTime: 10, // 10 second blocks  
    gasPrice: {
      average: '0.01',
      fast: '0.02',
      standard: '0.01'
    },
    confirmations: 3,
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