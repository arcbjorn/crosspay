import { createPublicClient, createWalletClient, custom, http, type Address } from 'viem';
import { liskSepolia, baseSepolia } from 'viem/chains';

// Define Citrea testnet chain
const citreaTestnet = {
  id: 5115,
  name: 'Citrea Testnet',
  network: 'citrea-testnet',
  nativeCurrency: {
    name: 'Citrea Bitcoin',
    symbol: 'cBTC',
    decimals: 18,
  },
  rpcUrls: {
    default: { http: ['https://rpc.testnet.citrea.xyz'] },
    public: { http: ['https://rpc.testnet.citrea.xyz'] },
  },
  blockExplorers: {
    default: {
      name: 'Citrea Explorer',
      url: 'https://explorer.testnet.citrea.xyz',
    },
  },
  testnet: true,
} as const;
import PaymentCoreABI from '../abis/PaymentCore.json';
import ReceiptRegistryABI from '../abis/ReceiptRegistry.json';
import FlareOracleABI from '../abis/FlareOracle.json';
import SubnameRegistryABI from '../abis/SubnameRegistry.json';
import type { ChainConfig } from '@stores/chain';

// Contract addresses - these will be updated after deployment
const CONTRACT_ADDRESSES: Record<number, {
  PaymentCore: Address;
  ReceiptRegistry: Address;
}> = {
  4202: { // Lisk Sepolia
    PaymentCore: '0x0000000000000000000000000000000000000000' as Address,
    ReceiptRegistry: '0x0000000000000000000000000000000000000000' as Address,
  },
  84532: { // Base Sepolia
    PaymentCore: '0x0000000000000000000000000000000000000000' as Address,
    ReceiptRegistry: '0x0000000000000000000000000000000000000000' as Address,
  },
  5115: { // Citrea Testnet
    PaymentCore: '0x0000000000000000000000000000000000000000' as Address,
    ReceiptRegistry: '0x0000000000000000000000000000000000000000' as Address,
  },
};

const CHAIN_CONFIG = {
  4202: liskSepolia,
  84532: baseSepolia,
  5115: citreaTestnet,
};

export function getPublicClient(chainId: number) {
  const chain = CHAIN_CONFIG[chainId as keyof typeof CHAIN_CONFIG];
  if (!chain) {
    throw new Error(`Unsupported chain ID: ${chainId}`);
  }
  
  return createPublicClient({
    chain,
    transport: http(),
  });
}

export function getWalletClient(chainId: number) {
  const chain = CHAIN_CONFIG[chainId as keyof typeof CHAIN_CONFIG];
  if (!chain) {
    throw new Error(`Unsupported chain ID: ${chainId}`);
  }
  
  if (!window.ethereum) {
    throw new Error('MetaMask not found');
  }
  
  return createWalletClient({
    chain,
    transport: custom(window.ethereum),
  });
}

export function getContractAddress(chainId: number, contract: 'PaymentCore' | 'ReceiptRegistry'): Address {
  const addresses = CONTRACT_ADDRESSES[chainId];
  if (!addresses) {
    throw new Error(`No contract addresses for chain ID: ${chainId}`);
  }
  return addresses[contract];
}

export { PaymentCoreABI, ReceiptRegistryABI, FlareOracleABI, SubnameRegistryABI };

export type PaymentCoreContract = {
  address: Address;
  abi: typeof PaymentCoreABI;
};

export type ReceiptRegistryContract = {
  address: Address;
  abi: typeof ReceiptRegistryABI;
};