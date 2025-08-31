import { parseEther, formatEther, parseEventLogs, type Address } from 'viem';
import { getPublicClient, getWalletClient, getContractAddress, PaymentCoreABI } from '../contracts';
import type { Payment, PaymentStatus } from '../../../packages/types/contracts';

export class PaymentService {
  constructor(private chainId: number) {}

  async createPayment(
    recipient: Address,
    token: Address,
    amount: string,
    metadataURI: string,
    senderAddress: Address
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
      const { request } = await publicClient.simulateContract({
        address: contractAddress,
        abi: PaymentCoreABI,
        functionName: 'createPayment',
        args: [recipient, token, parsedAmount, metadataURI],
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

      return {
        hash,
        paymentId: (args.id ?? args.paymentId) as bigint,
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
      // Struct order: id, sender, recipient, token, amount, fee, status, createdAt, completedAt, metadataURI
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
}

// Export singleton instances for each supported chain
export const liskPaymentService = new PaymentService(4202);
export const basePaymentService = new PaymentService(84532);
