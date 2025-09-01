import { describe, it, expect, vi, beforeEach } from 'vitest';
import { PaymentService } from '@services/payment';
import type { Address } from 'viem';

// Mock the contracts module
const mockSimulateContract = vi.fn();
const mockWriteContract = vi.fn();
const mockWaitForTransactionReceipt = vi.fn();

vi.mock('@contracts', () => ({
	getPublicClient: vi.fn(() => ({
		simulateContract: mockSimulateContract,
		waitForTransactionReceipt: mockWaitForTransactionReceipt,
		readContract: vi.fn()
	})),
	getWalletClient: vi.fn(() => ({
		writeContract: mockWriteContract
	})),
	getContractAddress: vi.fn(() => '0xcontract123'),
	PaymentCoreABI: []
}));

// Mock the oracle service
vi.mock('@services/oracle', () => ({
	oracleService: {
		getCurrentPrice: vi.fn()
	}
}));

// Mock viem functions
vi.mock('viem', async () => {
	const actual = await vi.importActual('viem');
	return {
		...actual,
		parseEventLogs: vi.fn()
	};
});

describe('PaymentService', () => {
	let paymentService: PaymentService;
	const mockSender = '0x1234567890123456789012345678901234567890' as Address;
	const mockRecipient = '0x0987654321098765432109876543210987654321' as Address;
	const mockToken = '0x0000000000000000000000000000000000000000' as Address; // ETH

	beforeEach(() => {
		vi.clearAllMocks();
		paymentService = new PaymentService(4202);
	});

	describe('formatAmount', () => {
		it('should format ETH amount correctly', () => {
			const result = paymentService.formatAmount(BigInt('1000000000000000000'));
			expect(result).toBe('1');
		});

		it('should handle zero amount', () => {
			const result = paymentService.formatAmount(0n);
			expect(result).toBe('0');
		});
	});

	describe('getContractAddress', () => {
		it('should return contract address', () => {
			const address = paymentService.getContractAddress();
			expect(address).toBe('0xcontract123');
		});
	});

	describe('createPayment', () => {
		it('should create payment with oracle price snapshot', async () => {
			const mockRequest = {
				address: '0xcontract123' as Address,
				abi: [],
				functionName: 'createPayment',
				args: [mockRecipient, mockToken, BigInt('1000000000000000000'), '', 'alice.eth', 'bob.eth'],
				account: mockSender,
				value: BigInt('1001000000000000000')
			};

			// Setup mocks using the direct mock functions
			const { oracleService } = await import('@services/oracle');

			// Mock oracle price
			vi.mocked(oracleService.getCurrentPrice).mockResolvedValue({
				symbol: 'ETH/USD',
				price: 2500.0,
				timestamp: Date.now(),
				decimals: 8,
				valid: true
			});

			// Mock contract simulation
			mockSimulateContract.mockResolvedValue({
				request: mockRequest,
				result: undefined
			});

			// Mock transaction
			mockWriteContract.mockResolvedValue('0xtxhash123' as Address);

			// Mock receipt with logs
			mockWaitForTransactionReceipt.mockResolvedValue({
				logs: [
					{
						address: '0xcontract123',
						topics: ['0xevent123'],
						data: '0x123'
					}
				]
			} as any);

			// Mock parseEventLogs globally
			const { parseEventLogs } = await import('viem');
			vi.mocked(parseEventLogs).mockReturnValue([
				{
					args: { paymentId: 123n },
					eventName: 'PaymentCreated'
				} as any
			]);

			// Mock the private approveTokenIfNeeded method
			vi.spyOn(paymentService as any, 'approveTokenIfNeeded').mockResolvedValue(undefined);

			const result = await paymentService.createPayment(
				mockRecipient,
				mockToken,
				'1.0',
				'',
				mockSender,
				'alice.eth',
				'bob.eth'
			);

			expect(result.hash).toBe('0xtxhash123');
			expect(oracleService.getCurrentPrice).toHaveBeenCalledWith('ETH/USD');
		});
	});
});
