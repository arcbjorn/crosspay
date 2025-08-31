import { parseEther, formatEther, parseEventLogs, type Address } from 'viem';
import { getPublicClient, getWalletClient, getContractAddress, PaymentCoreABI } from '../contracts';
import { oracleService } from './oracle';
import type { Payment, PaymentStatus } from '@types/contracts';

export class PaymentService {
  constructor(private chainId: number) {}

  async createPayment(
    recipient: Address,
    token: Address,
    amount: string,
    metadataURI: string,
    senderAddress: Address,
    senderENS?: string,
    recipientENS?: string
  ): Promise<{ hash: string; paymentId: bigint }> {
    const walletClient = getWalletClient(this.chainId);
    const publicClient = getPublicClient(this.chainId);
    const contractAddress = getContractAddress(this.chainId, 'PaymentCore');

    // Parse amount (ETH has 18 decimals)
    const parsedAmount = parseEther(amount);
    const isETH = token === '0x0000000000000000000000000000000000000000';
    // Protocol fee: 10 bps (0.1%)
    const fee = (parsedAmount * 10n) / 10000n;
    const totalValue = isETH ? parsedAmount + fee : 0n;
    
    try {
      // Get price snapshot from oracle
      let oraclePrice = 0n;
      try {
        const tokenSymbol = isETH ? 'ETH' : 'USDC'; // Default mapping
        const priceData = await oracleService.getCurrentPrice(`${tokenSymbol}/USD`);
        oraclePrice = BigInt(Math.round(priceData.price * 1e8)); // Convert to 8 decimal format
      } catch (oracleError) {
        console.warn('Failed to get oracle price:', oracleError);
        // Continue without price snapshot
      }

      // Handle ERC20 token approval if needed
      if (!isETH) {
        await this.approveTokenIfNeeded(token, parsedAmount + fee, senderAddress);
      }
      const { request } = await publicClient.simulateContract({
        address: contractAddress,
        abi: PaymentCoreABI,
        functionName: 'createPayment',
        args: [recipient, token, parsedAmount, metadataURI, senderENS || '', recipientENS || ''],
        account: senderAddress,
        value: totalValue,
      });

      const hash = await walletClient.writeContract(request);

      // Wait for transaction receipt to get payment ID
      const receipt = await publicClient.waitForTransactionReceipt({ hash });

      // Parse PaymentCreated event to get payment ID
      const parsedLogs = parseEventLogs({
        abi: PaymentCoreABI,
        logs: receipt.logs,
        eventName: 'PaymentCreated'
      });

      if (!parsedLogs.length) throw new Error('PaymentCreated event not found');
      const { args } = parsedLogs[0] as any;

      const paymentId = (args.id ?? args.paymentId) as bigint;

      // Set oracle price if we have one (owner-only function for demo)
      if (oraclePrice > 0n) {
        try {
          await this.setOraclePrice(paymentId, oraclePrice, senderAddress);
        } catch (oraclePriceError) {
          console.warn('Failed to set oracle price:', oraclePriceError);
          // Continue without setting price
        }
      }

      return {
        hash,
        paymentId,
      };
    } catch (error) {
      console.error('Payment creation failed:', error);
      throw error;
    }
  }

  async getPayment(paymentId: bigint): Promise<Payment> {
    const publicClient = getPublicClient(this.chainId);
    const contractAddress = getContractAddress(this.chainId, 'PaymentCore');

    try {
      const result = (await publicClient.readContract({
        address: contractAddress,
        abi: PaymentCoreABI,
        functionName: 'getPayment',
        args: [paymentId],
      })) as any;

      // Map the contract result (tuple) to our Payment interface
      // Struct order: id, sender, recipient, token, amount, fee, status, createdAt, completedAt, metadataURI, receiptCID, senderENS, recipientENS, oraclePrice, randomSeed
      const arr = Array.isArray(result) ? (result as any[]) : (Object.values(result) as any[]);
      return {
        id: paymentId,
        sender: arr[1] as Address,
        recipient: arr[2] as Address,
        token: arr[3] as Address,
        amount: arr[4] as bigint,
        fee: arr[5] as bigint,
        status: this.mapStatus(Number(arr[6] ?? 0)),
        createdAt: arr[7] as bigint,
        completedAt: arr[8] as bigint,
        metadataURI: arr[9] as string,
        receiptCID: arr[10] as string,
        senderENS: arr[11] as string,
        recipientENS: arr[12] as string,
        oraclePrice: BigInt(arr[13] as string || '0'),
        randomSeed: arr[14] as string,
      };
    } catch (error) {
      console.error('Failed to get payment:', error);
      throw error;
    }
  }

  async completePayment(paymentId: bigint, senderAddress: Address): Promise<string> {
    const walletClient = getWalletClient(this.chainId);
    const publicClient = getPublicClient(this.chainId);
    const contractAddress = getContractAddress(this.chainId, 'PaymentCore');

    try {
      const { request } = await publicClient.simulateContract({
        address: contractAddress,
        abi: PaymentCoreABI,
        functionName: 'completePayment',
        args: [paymentId],
        account: senderAddress,
      });

      return await walletClient.writeContract(request);
    } catch (error) {
      console.error('Payment completion failed:', error);
      throw error;
    }
  }

  async refundPayment(paymentId: bigint, senderAddress: Address): Promise<string> {
    const walletClient = getWalletClient(this.chainId);
    const publicClient = getPublicClient(this.chainId);
    const contractAddress = getContractAddress(this.chainId, 'PaymentCore');

    try {
      const { request } = await publicClient.simulateContract({
        address: contractAddress,
        abi: PaymentCoreABI,
        functionName: 'refundPayment',
        args: [paymentId],
        account: senderAddress,
      });

      return await walletClient.writeContract(request);
    } catch (error) {
      console.error('Payment refund failed:', error);
      throw error;
    }
  }

  async getUserPayments(userAddress: Address, sent: boolean = true): Promise<bigint[]> {
    const publicClient = getPublicClient(this.chainId);
    const contractAddress = getContractAddress(this.chainId, 'PaymentCore');

    try {
      const result = await publicClient.readContract({
        address: contractAddress,
        abi: PaymentCoreABI,
        functionName: sent ? 'getSenderPayments' : 'getRecipientPayments',
        args: [userAddress],
      }) as bigint[];

      return result;
    } catch (error) {
      console.error('Failed to get user payments:', error);
      throw error;
    }
  }

  private mapStatus(status: number): PaymentStatus {
    switch (status) {
      case 0: return 'pending';
      case 1: return 'completed';
      case 2: return 'refunded';
      case 3: return 'cancelled';
      default: return 'pending';
    }
  }

  formatAmount(amount: bigint): string {
    return formatEther(amount);
  }

  getContractAddress(): Address {
    return getContractAddress(this.chainId, 'PaymentCore');
  }

  async setOraclePrice(paymentId: bigint, price: bigint, senderAddress: Address): Promise<string> {
    const walletClient = getWalletClient(this.chainId);
    const publicClient = getPublicClient(this.chainId);
    const contractAddress = getContractAddress(this.chainId, 'PaymentCore');

    try {
      const { request } = await publicClient.simulateContract({
        address: contractAddress,
        abi: PaymentCoreABI,
        functionName: 'setOraclePrice',
        args: [paymentId, price],
        account: senderAddress,
      });

      return await walletClient.writeContract(request);
    } catch (error) {
      console.error('Failed to set oracle price:', error);
      throw error;
    }
  }

  private async approveTokenIfNeeded(
    tokenAddress: Address,
    amount: bigint,
    ownerAddress: Address
  ): Promise<void> {
    const publicClient = getPublicClient(this.chainId);
    const walletClient = getWalletClient(this.chainId);
    const contractAddress = getContractAddress(this.chainId, 'PaymentCore');

    // ERC20 ABI for allowance and approve
    const erc20ABI = [
      {
        name: 'allowance',
        type: 'function',
        stateMutability: 'view',
        inputs: [
          { name: 'owner', type: 'address' },
          { name: 'spender', type: 'address' }
        ],
        outputs: [{ name: '', type: 'uint256' }]
      },
      {
        name: 'approve',
        type: 'function',
        stateMutability: 'nonpayable',
        inputs: [
          { name: 'spender', type: 'address' },
          { name: 'amount', type: 'uint256' }
        ],
        outputs: [{ name: '', type: 'bool' }]
      }
    ] as const;

    try {
      // Check current allowance
      const currentAllowance = await publicClient.readContract({
        address: tokenAddress,
        abi: erc20ABI,
        functionName: 'allowance',
        args: [ownerAddress, contractAddress],
      }) as bigint;

      // If allowance is sufficient, no approval needed
      if (currentAllowance >= amount) {
        console.log('Sufficient token allowance already exists');
        return;
      }

      console.log('Token approval required, requesting approval...');

      // Request approval for the required amount
      const { request } = await publicClient.simulateContract({
        address: tokenAddress,
        abi: erc20ABI,
        functionName: 'approve',
        args: [contractAddress, amount],
        account: ownerAddress,
      });

      const hash = await walletClient.writeContract(request);
      
      // Wait for approval transaction to be mined
      await publicClient.waitForTransactionReceipt({ hash });
      
      console.log('Token approval successful');

    } catch (error) {
      console.error('Token approval failed:', error);
      throw new Error('Token approval failed. Please try again.');
    }
  }
}

// Export singleton instances for each supported chain
export const liskPaymentService = new PaymentService(4202);
export const basePaymentService = new PaymentService(84532);
export const citreaPaymentService = new PaymentService(5115);
