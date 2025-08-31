import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import WalletConnect from '../WalletConnect.svelte';
import { walletStore } from '$lib/stores/wallet';
import type { WalletState } from '$lib/stores/wallet';

// Mock the wallet store
vi.mock('$lib/stores/wallet', () => ({
  walletStore: {
    subscribe: vi.fn(),
  },
  connectWallet: vi.fn(),
  disconnectWallet: vi.fn(),
}));

vi.mock('$lib/stores/chain', () => ({
  chainStore: {
    subscribe: vi.fn(() => ({
      id: 4202,
      name: 'Lisk Sepolia',
    })),
  },
  isChainSupported: vi.fn(() => true),
  switchChain: vi.fn(),
}));

describe('WalletConnect', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render connect button when wallet is not connected', () => {
    const mockWalletState: WalletState = {
      isConnected: false,
      isConnecting: false,
    };

    vi.mocked(walletStore.subscribe).mockImplementation((callback) => {
      callback(mockWalletState);
      return () => {};
    });

    render(WalletConnect);
    
    expect(screen.getByText('Connect Wallet')).toBeInTheDocument();
  });

  it('should render connecting state', () => {
    const mockWalletState: WalletState = {
      isConnected: false,
      isConnecting: true,
    };

    vi.mocked(walletStore.subscribe).mockImplementation((callback) => {
      callback(mockWalletState);
      return () => {};
    });

    render(WalletConnect);
    
    expect(screen.getByText('Connecting...')).toBeInTheDocument();
  });

  it('should render wallet info when connected', () => {
    const mockWalletState: WalletState = {
      isConnected: true,
      isConnecting: false,
      address: '0x1234567890123456789012345678901234567890' as any,
      balance: BigInt('1000000000000000000'), // 1 ETH
      chainId: 4202,
    };

    vi.mocked(walletStore.subscribe).mockImplementation((callback) => {
      callback(mockWalletState);
      return () => {};
    });

    render(WalletConnect);
    
    expect(screen.getByText('0x1234...7890')).toBeInTheDocument();
  });

  it('should show error toast when there is an error', () => {
    const mockWalletState: WalletState = {
      isConnected: false,
      isConnecting: false,
      error: 'Connection failed',
    };

    vi.mocked(walletStore.subscribe).mockImplementation((callback) => {
      callback(mockWalletState);
      return () => {};
    });

    render(WalletConnect);
    
    expect(screen.getByText('Connection failed')).toBeInTheDocument();
  });
});