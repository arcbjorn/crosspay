import { describe, it, expect, vi, beforeEach } from 'vitest';
import { StorageService } from '../storage';

// Mock fetch
global.fetch = vi.fn();

describe('StorageService', () => {
  let storageService: StorageService;

  beforeEach(() => {
    vi.clearAllMocks();
    storageService = new StorageService('http://localhost:3001');
  });

  describe('uploadFile', () => {
    it('should upload file successfully', async () => {
      const mockResponse = {
        cid: 'bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi',
        size: 1024,
        filename: 'test.json',
        url: 'https://gateway.lighthouse.storage/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi',
      };

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      } as Response);

      const file = new File(['test content'], 'test.json', { type: 'application/json' });
      const result = await storageService.uploadFile(file);

      expect(fetch).toHaveBeenCalledWith('http://localhost:3001/upload', {
        method: 'POST',
        body: expect.any(FormData),
      });
      expect(result).toEqual(mockResponse);
    });

    it('should handle upload errors', async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.resolve({ error: 'Upload failed: 500 Internal Server Error' }),
      } as Response);

      const file = new File(['test content'], 'test.json', { type: 'application/json' });

      await expect(storageService.uploadFile(file)).rejects.toThrow('Upload failed: 500 Internal Server Error');
    });

    it('should handle network errors', async () => {
      vi.mocked(fetch).mockRejectedValueOnce(new Error('Network error'));

      const file = new File(['test content'], 'test.json', { type: 'application/json' });

      await expect(storageService.uploadFile(file)).rejects.toThrow('Network error');
    });
  });

  describe('downloadFile', () => {
    beforeEach(() => {
      // Mock DOM methods
      Object.defineProperty(document, 'createElement', {
        value: vi.fn().mockReturnValue({
          href: '',
          download: '',
          click: vi.fn(),
        }),
        writable: true,
      });
      Object.defineProperty(document.body, 'appendChild', {
        value: vi.fn(),
        writable: true,
      });
      Object.defineProperty(document.body, 'removeChild', {
        value: vi.fn(),
        writable: true,
      });
      Object.defineProperty(URL, 'createObjectURL', {
        value: vi.fn().mockReturnValue('blob:test'),
        writable: true,
      });
      Object.defineProperty(URL, 'revokeObjectURL', {
        value: vi.fn(),
        writable: true,
      });
    });

    it('should download file successfully', async () => {
      const mockRetrieveResult = {
        data: new ArrayBuffer(12),
        filename: 'test.json',
        contentType: 'application/json',
        metadata: {},
        size: 12,
        timestamp: '2024-01-01T00:00:00Z',
      };

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockRetrieveResult),
      } as Response);

      await storageService.downloadFile('bafybeigtest');

      expect(fetch).toHaveBeenCalledWith('http://localhost:3001/api/storage/retrieve/bafybeigtest');
      expect(document.createElement).toHaveBeenCalledWith('a');
    });

    it('should handle download errors', async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: () => Promise.resolve({ error: 'File not found: invalid-cid' }),
      } as Response);

      await expect(storageService.downloadFile('invalid-cid')).rejects.toThrow('File not found: invalid-cid');
    });
  });

  describe('verifyFile', () => {
    it('should verify file successfully', async () => {
      const mockResponse = {
        cid: 'bafybeigtest',
        valid: true,
        size: 1024,
        pinned: true,
        metadata: {
          filename: 'test.json',
          contentType: 'application/json',
          uploadedAt: '2024-01-01T00:00:00Z',
        },
      };

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      } as Response);

      const result = await storageService.verifyFile('bafybeigtest');

      expect(fetch).toHaveBeenCalledWith('http://localhost:3001/verify/bafybeigtest');
      expect(result).toEqual(mockResponse);
    });

    it('should handle verification errors', async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      } as Response);

      await expect(storageService.verifyFile('invalid-cid')).rejects.toThrow('Verification failed: 404 Not Found');
    });

    it('should return invalid status for corrupted files', async () => {
      const mockResponse = {
        cid: 'bafybeigtest',
        valid: false,
        size: 0,
        pinned: false,
        metadata: null,
      };

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      } as Response);

      const result = await storageService.verifyFile('bafybeigtest');

      expect(result.valid).toBe(false);
    });
  });

  describe('constructor', () => {
    it('should use provided base URL', () => {
      const service = new StorageService('https://api.example.com');
      expect(service['baseUrl']).toBe('https://api.example.com');
    });

    it('should use environment variable when no URL provided', () => {
      vi.stubEnv('VITE_STORAGE_URL', 'https://env.example.com');
      const service = new StorageService();
      expect(service['baseUrl']).toBe('https://env.example.com');
    });

    it('should fallback to localhost when no URL or env var', () => {
      vi.unstubAllEnvs();
      const service = new StorageService();
      expect(service['baseUrl']).toBe('http://localhost:3001');
    });

    it('should remove trailing slash from base URL', () => {
      const service = new StorageService('https://api.example.com/');
      expect(service['baseUrl']).toBe('https://api.example.com');
    });
  });
});