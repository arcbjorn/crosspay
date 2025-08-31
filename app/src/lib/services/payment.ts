import { parseEther, parseUnits, formatEther, type Address } from 'viem';
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

    // Parse amount based on token (ETH has 18 decimals)
    const parsedAmount = parseEther(amount);

    // For ETH payments, send value along with the transaction
    const isETH = token === '0x0000000000000000000000000000000000000000';
    
    try {
      const { request } = await publicClient.simulateContract({
        address: contractAddress,
        abi: PaymentCoreABI,
        functionName: 'createPayment',
        args: [recipient, token, parsedAmount, metadataURI],
        account: senderAddress,
        value: isETH ? parsedAmount : 0n,
      });

      const hash = await walletClient.writeContract(request);

      // Wait for transaction receipt to get payment ID
      const receipt = await publicClient.waitForTransactionReceipt({ hash });
      
      // Find PaymentCreated event to get payment ID
      const paymentCreatedLog = receipt.logs.find(log => 
        log.address.toLowerCase() === contractAddress.toLowerCase()
      );

      if (!paymentCreatedLog) {
        throw new Error('PaymentCreated event not found');
      }

      // Decode the event log to get payment ID
      const { args } = await publicClient.parseEventLogs({
        abi: PaymentCoreABI,
        logs: [paymentCreatedLog],
        eventName: 'PaymentCreated'
      })[0];

      return {
        hash,
        paymentId: args.paymentId as bigint,
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
      const result = await publicClient.readContract({
        address: contractAddress,
        abi: PaymentCoreABI,
        functionName: 'getPayment',
        args: [paymentId],
      }) as any[];

      // Map the contract result to our Payment interface
      return {
        id: paymentId,
        sender: result[0] as Address,
        recipient: result[1] as Address,
        token: result[2] as Address,
        amount: result[3] as bigint,
        fee: result[4] as bigint,
        status: this.mapStatus(result[5] as number),
        createdAt: result[6] as bigint,
        completedAt: result[7] as bigint,
        metadataURI: result[8] as string,
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
        functionName: sent ? 'getSentPayments' : 'getReceivedPayments',
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