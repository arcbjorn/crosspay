import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ERC20Service } from '@services/erc20';
import type { Address } from 'viem';

// Mock the contracts module
vi.mock('@contracts', () => ({
  getPublicClient: vi.fn(() => ({
    readContract: vi.fn(),
    estimateContractGas: vi.fn(),
    simulateContract: vi.fn(),
    getBalance: vi.fn(),
  })),
  getWalletClient: vi.fn(() => ({
    writeContract: vi.fn(),
  })),
}));

describe('ERC20Service', () => {
  let erc20Service: ERC20Service;
  const mockTokenAddress = '0x1234567890123456789012345678901234567890' as Address;
  const mockUserAddress = '0x0987654321098765432109876543210987654321' as Address;
  const mockSpenderAddress = '0xabcdefabcdefabcdefabcdefabcdefabcdefabcd' as Address;

  beforeEach(() => {
    vi.clearAllMocks();
    erc20Service = new ERC20Service(4202);
  });

  describe('parseTokenAmount', () => {
    it('should parse token amount correctly', () => {
      const result = erc20Service.parseTokenAmount('1.5', 18);
      expect(result).toBe(BigInt('1500000000000000000'));
    });

    it('should handle zero amount', () => {
      const result = erc20Service.parseTokenAmount('0', 18);
      expect(result).toBe(0n);
    });

    it('should handle invalid amount', () => {
      const result = erc20Service.parseTokenAmount('invalid', 18);
      expect(result).toBe(0n);
    });
  });

  describe('formatTokenAmount', () => {
    it('should format token amount correctly', () => {
      const result = erc20Service.formatTokenAmount(BigInt('1500000000000000000'), 18, 4);
      expect(result).toBe('1.5');
    });

    it('should trim trailing zeros', () => {
      const result = erc20Service.formatTokenAmount(BigInt('1000000000000000000'), 18, 4);
      expect(result).toBe('1');
    });

    it('should handle zero amount', () => {
      const result = erc20Service.formatTokenAmount(0n, 18, 4);
      expect(result).toBe('0');
    });
  });

  describe('getApprovalStatus', () => {
    it('should return correct approval status when approved', async () => {
      // Mock checkAllowance method directly
      vi.spyOn(erc20Service, 'checkAllowance').mockResolvedValue(BigInt('2000000000000000000'));
      
      const result = await erc20Service.getApprovalStatus(
        mockTokenAddress,
        mockUserAddress,
        mockSpenderAddress,
        BigInt('1000000000000000000') // 1 ETH required
      );

      expect(result.hasApproval).toBe(true);
      expect(result.needsApproval).toBe(false);
      expect(result.currentAllowance).toBe(BigInt('2000000000000000000'));
      expect(result.requiredAmount).toBe(BigInt('1000000000000000000'));
    });

    it('should return correct approval status when not approved', async () => {
      // Mock checkAllowance method directly
      vi.spyOn(erc20Service, 'checkAllowance').mockResolvedValue(BigInt('500000000000000000'));
      
      const result = await erc20Service.getApprovalStatus(
        mockTokenAddress,
        mockUserAddress,
        mockSpenderAddress,
        BigInt('1000000000000000000') // 1 ETH required
      );

      expect(result.hasApproval).toBe(false);
      expect(result.needsApproval).toBe(true);
      expect(result.currentAllowance).toBe(BigInt('500000000000000000'));
      expect(result.requiredAmount).toBe(BigInt('1000000000000000000'));
    });
  });
});