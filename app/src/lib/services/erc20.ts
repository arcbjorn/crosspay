import type { Address } from 'viem';
import { parseUnits, formatUnits } from 'viem';
import { getPublicClient, getWalletClient } from '../contracts';
import type { TokenInfo } from '../config/tokens';

// Standard ERC20 ABI (minimal required functions)
const ERC20_ABI = [
  {
    name: 'balanceOf',
    type: 'function',
    stateMutability: 'view',
    inputs: [{ name: 'account', type: 'address' }],
    outputs: [{ name: '', type: 'uint256' }],
  },
  {
    name: 'allowance',
    type: 'function',
    stateMutability: 'view',
    inputs: [
      { name: 'owner', type: 'address' },
      { name: 'spender', type: 'address' }
    ],
    outputs: [{ name: '', type: 'uint256' }],
  },
  {
    name: 'approve',
    type: 'function',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'spender', type: 'address' },
      { name: 'amount', type: 'uint256' }
    ],
    outputs: [{ name: '', type: 'bool' }],
  },
  {
    name: 'transfer',
    type: 'function',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'to', type: 'address' },
      { name: 'amount', type: 'uint256' }
    ],
    outputs: [{ name: '', type: 'bool' }],
  },
  {
    name: 'transferFrom',
    type: 'function',
    stateMutability: 'nonpayable',
    inputs: [
      { name: 'from', type: 'address' },
      { name: 'to', type: 'address' },
      { name: 'amount', type: 'uint256' }
    ],
    outputs: [{ name: '', type: 'bool' }],
  },
  {
    name: 'decimals',
    type: 'function',
    stateMutability: 'view',
    inputs: [],
    outputs: [{ name: '', type: 'uint8' }],
  },
  {
    name: 'symbol',
    type: 'function',
    stateMutability: 'view',
    inputs: [],
    outputs: [{ name: '', type: 'string' }],
  },
  {
    name: 'name',
    type: 'function',
    stateMutability: 'view',
    inputs: [],
    outputs: [{ name: '', type: 'string' }],
  },
  // Events
  {
    name: 'Transfer',
    type: 'event',
    inputs: [
      { name: 'from', type: 'address', indexed: true },
      { name: 'to', type: 'address', indexed: true },
      { name: 'value', type: 'uint256', indexed: false }
    ],
    anonymous: false,
  },
  {
    name: 'Approval',
    type: 'event',
    inputs: [
      { name: 'owner', type: 'address', indexed: true },
      { name: 'spender', type: 'address', indexed: true },
      { name: 'value', type: 'uint256', indexed: false }
    ],
    anonymous: false,
  },
] as const;

export interface ERC20TokenBalance {
  token: TokenInfo;
  balance: bigint;
  formattedBalance: string;
  usdValue?: number;
}

export interface ERC20ApprovalStatus {
  hasApproval: boolean;
  currentAllowance: bigint;
  requiredAmount: bigint;
  needsApproval: boolean;
}

export class ERC20Service {
  constructor(private chainId: number) {}

  /**
   * Get token balance for an address
   */
  async getTokenBalance(tokenAddress: Address, userAddress: Address): Promise<bigint> {
    const publicClient = getPublicClient(this.chainId);

    try {
      const balance = await publicClient.readContract({
        address: tokenAddress,
        abi: ERC20_ABI,
        functionName: 'balanceOf',
        args: [userAddress],
      }) as bigint;

      return balance;
    } catch (error) {
      console.error('Failed to get token balance:', error);
      return 0n;
    }
  }

  /**
   * Get multiple token balances at once
   */
  async getTokenBalances(tokens: TokenInfo[], userAddress: Address): Promise<ERC20TokenBalance[]> {
    const balances: ERC20TokenBalance[] = [];

    for (const token of tokens) {
      try {
        const balance = token.isNative 
          ? await this.getNativeBalance(userAddress)
          : await this.getTokenBalance(token.address, userAddress);

        balances.push({
          token,
          balance,
          formattedBalance: formatUnits(balance, token.decimals),
        });
      } catch (error) {
        console.error(`Failed to get balance for ${token.symbol}:`, error);
        balances.push({
          token,
          balance: 0n,
          formattedBalance: '0',
        });
      }
    }

    return balances;
  }

  /**
   * Get native token balance (ETH, cBTC, etc.)
   */
  async getNativeBalance(userAddress: Address): Promise<bigint> {
    const publicClient = getPublicClient(this.chainId);
    
    try {
      return await publicClient.getBalance({ address: userAddress });
    } catch (error) {
      console.error('Failed to get native balance:', error);
      return 0n;
    }
  }

  /**
   * Check token allowance
   */
  async checkAllowance(
    tokenAddress: Address, 
    ownerAddress: Address, 
    spenderAddress: Address
  ): Promise<bigint> {
    const publicClient = getPublicClient(this.chainId);

    try {
      const allowance = await publicClient.readContract({
        address: tokenAddress,
        abi: ERC20_ABI,
        functionName: 'allowance',
        args: [ownerAddress, spenderAddress],
      }) as bigint;

      return allowance;
    } catch (error) {
      console.error('Failed to check allowance:', error);
      return 0n;
    }
  }

  /**
   * Get approval status for a token and amount
   */
  async getApprovalStatus(
    tokenAddress: Address,
    ownerAddress: Address,
    spenderAddress: Address,
    requiredAmount: bigint
  ): Promise<ERC20ApprovalStatus> {
    const currentAllowance = await this.checkAllowance(tokenAddress, ownerAddress, spenderAddress);
    const hasApproval = currentAllowance >= requiredAmount;

    return {
      hasApproval,
      currentAllowance,
      requiredAmount,
      needsApproval: !hasApproval,
    };
  }

  /**
   * Approve token spending
   */
  async approveToken(
    tokenAddress: Address, 
    spenderAddress: Address, 
    amount: bigint,
    userAddress: Address
  ): Promise<string> {
    const walletClient = getWalletClient(this.chainId);
    const publicClient = getPublicClient(this.chainId);

    try {
      // First simulate the transaction
      const { request } = await publicClient.simulateContract({
        address: tokenAddress,
        abi: ERC20_ABI,
        functionName: 'approve',
        args: [spenderAddress, amount],
        account: userAddress,
      });

      // Execute the transaction
      const hash = await walletClient.writeContract(request);

      // Wait for confirmation
      await publicClient.waitForTransactionReceipt({ hash });

      return hash;
    } catch (error) {
      console.error('Token approval failed:', error);
      throw error;
    }
  }

  /**
   * Approve maximum amount (useful for better UX)
   */
  async approveMax(
    tokenAddress: Address, 
    spenderAddress: Address, 
    userAddress: Address
  ): Promise<string> {
    const maxAmount = 2n ** 256n - 1n; // Maximum uint256 value
    return this.approveToken(tokenAddress, spenderAddress, maxAmount, userAddress);
  }

  /**
   * Revoke token approval (set to 0)
   */
  async revokeApproval(
    tokenAddress: Address, 
    spenderAddress: Address, 
    userAddress: Address
  ): Promise<string> {
    return this.approveToken(tokenAddress, spenderAddress, 0n, userAddress);
  }

  /**
   * Get token info from contract
   */
  async getTokenInfo(tokenAddress: Address): Promise<Partial<TokenInfo>> {
    const publicClient = getPublicClient(this.chainId);

    try {
      const [name, symbol, decimals] = await Promise.all([
        publicClient.readContract({
          address: tokenAddress,
          abi: ERC20_ABI,
          functionName: 'name',
        }) as Promise<string>,
        publicClient.readContract({
          address: tokenAddress,
          abi: ERC20_ABI,
          functionName: 'symbol',
        }) as Promise<string>,
        publicClient.readContract({
          address: tokenAddress,
          abi: ERC20_ABI,
          functionName: 'decimals',
        }) as Promise<number>,
      ]);

      return {
        address: tokenAddress,
        name,
        symbol,
        decimals,
      };
    } catch (error) {
      console.error('Failed to get token info:', error);
      return { address: tokenAddress };
    }
  }

  /**
   * Parse token amount string to bigint
   */
  parseTokenAmount(amount: string, decimals: number): bigint {
    try {
      return parseUnits(amount, decimals);
    } catch (error) {
      console.error('Failed to parse token amount:', error);
      return 0n;
    }
  }

  /**
   * Format token amount bigint to string
   */
  formatTokenAmount(amount: bigint, decimals: number, displayDecimals: number = 4): string {
    try {
      const formatted = formatUnits(amount, decimals);
      const num = parseFloat(formatted);
      return num.toFixed(displayDecimals).replace(/\.?0+$/, '');
    } catch (error) {
      console.error('Failed to format token amount:', error);
      return '0';
    }
  }

  /**
   * Estimate gas for token approval
   */
  async estimateApprovalGas(
    tokenAddress: Address,
    spenderAddress: Address,
    amount: bigint,
    userAddress: Address
  ): Promise<bigint> {
    const publicClient = getPublicClient(this.chainId);

    try {
      const gas = await publicClient.estimateContractGas({
        address: tokenAddress,
        abi: ERC20_ABI,
        functionName: 'approve',
        args: [spenderAddress, amount],
        account: userAddress,
      });

      return gas;
    } catch (error) {
      console.error('Failed to estimate approval gas:', error);
      return 100000n; // Fallback gas estimate
    }
  }

  /**
   * Check if address has sufficient token balance
   */
  async hasSufficientBalance(
    tokenAddress: Address,
    userAddress: Address,
    requiredAmount: bigint
  ): Promise<boolean> {
    const balance = await this.getTokenBalance(tokenAddress, userAddress);
    return balance >= requiredAmount;
  }

  /**
   * Safe approve - handles tokens that require 0 approval first
   */
  async safeApprove(
    tokenAddress: Address,
    spenderAddress: Address,
    amount: bigint,
    userAddress: Address
  ): Promise<string> {
    // Check current allowance
    const currentAllowance = await this.checkAllowance(tokenAddress, userAddress, spenderAddress);
    
    // If current allowance is not 0 and we're trying to set non-zero, reset to 0 first
    if (currentAllowance > 0n && amount > 0n) {
      await this.revokeApproval(tokenAddress, spenderAddress, userAddress);
    }
    
    // Now set the desired allowance
    return this.approveToken(tokenAddress, spenderAddress, amount, userAddress);
  }
}

// Export ERC20 ABI for use in other parts of the app
export { ERC20_ABI };

// Export types
export type { ERC20TokenBalance as TokenBalance, ERC20ApprovalStatus as ApprovalStatus };