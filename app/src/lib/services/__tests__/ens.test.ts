import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ENSService } from '../ens';

// Mock fetch
global.fetch = vi.fn();

describe('ENSService', () => {
  let ensService: ENSService;

  beforeEach(() => {
    vi.clearAllMocks();
    ensService = new ENSService('http://localhost:3002');
  });

  describe('resolveENSName', () => {
    it('should resolve ENS name to address successfully', async () => {
      const mockResponse = {
        name: 'alice.eth',
        address: '0x1234567890123456789012345678901234567890',
        resolved: true,
      };

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      } as Response);

      const result = await ensService.resolveENSName('alice.eth');

      expect(fetch).toHaveBeenCalledWith('http://localhost:3002/resolve/alice.eth');
      expect(result).toEqual(mockResponse);
    });

    it('should handle unresolved ENS names', async () => {
      const mockResponse = {
        name: 'nonexistent.eth',
        address: null,
        resolved: false,
      };

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      } as Response);

      const result = await ensService.resolveENSName('nonexistent.eth');

      expect(result).toEqual(mockResponse);
      expect(result.resolved).toBe(false);
    });

    it('should handle network errors', async () => {
      vi.mocked(fetch).mockRejectedValueOnce(new Error('Network error'));

      await expect(ensService.resolveENSName('alice.eth')).rejects.toThrow('Network error');
    });

    it('should handle API errors', async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
      } as Response);

      await expect(ensService.resolveENSName('alice.eth')).rejects.toThrow('Failed to resolve ENS name: 500 Internal Server Error');
    });
  });

  describe('reverseResolveAddress', () => {
    it('should reverse resolve address to ENS name successfully', async () => {
      const mockResponse = {
        address: '0x1234567890123456789012345678901234567890',
        name: 'alice.eth',
        resolved: true,
      };

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      } as Response);

      const result = await ensService.reverseResolveAddress('0x1234567890123456789012345678901234567890');

      expect(fetch).toHaveBeenCalledWith('http://localhost:3002/reverse/0x1234567890123456789012345678901234567890');
      expect(result).toEqual(mockResponse);
    });

    it('should handle addresses without ENS names', async () => {
      const mockResponse = {
        address: '0x9999999999999999999999999999999999999999',
        name: null,
        resolved: false,
      };

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      } as Response);

      const result = await ensService.reverseResolveAddress('0x9999999999999999999999999999999999999999');

      expect(result).toEqual(mockResponse);
      expect(result.resolved).toBe(false);
    });
  });

  describe('constructor', () => {
    it('should use provided base URL', () => {
      const service = new ENSService('https://api.example.com');
      expect(service['baseUrl']).toBe('https://api.example.com');
    });

    it('should use environment variable when no URL provided', () => {
      vi.stubEnv('VITE_ENS_URL', 'https://env.example.com');
      const service = new ENSService();
      expect(service['baseUrl']).toBe('https://env.example.com');
    });

    it('should fallback to localhost when no URL or env var', () => {
      vi.unstubAllEnvs();
      const service = new ENSService();
      expect(service['baseUrl']).toBe('http://localhost:3002');
    });
  });
});