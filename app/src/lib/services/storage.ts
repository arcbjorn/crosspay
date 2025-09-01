interface UploadResult {
  cid: string;
  size: number;
  cost: string;
  timestamp: string;
  dealId?: string;
  status?: string;
}

interface RetrieveResult {
  data: ArrayBuffer;
  filename: string;
  contentType: string;
  metadata: Record<string, string>;
  size: number;
  timestamp: string;
}

interface FileInfo {
  cid: string;
  size: number;
  dealId: string;
  storageCost: string;
  status: string;
  metadata: Record<string, string>;
  createdAt: string;
  name: string;
  type: string;
  uploadedAt: string;
}

interface DealStatus {
  dealId: string;
  cid: string;
  status: string;
  storageCost: string;
  duration: number;
  createdAt: string;
  expiresAt: string;
}

interface CostEstimate {
  sizeBytes: number;
  estimatedFIL: string;
  usdEquivalent: string;
}

interface NetworkInfo {
  networkId: string;
  blockHeight: number;
  gasPrice: string;
  storageProviders: number;
}

interface VerificationResult {
  cid: string;
  valid: boolean;
  size: number;
  pinned: boolean;
  metadata: Record<string, string> | null;
}

export class StorageService {
  private baseUrl: string;

  constructor(baseUrl?: string) {
    this.baseUrl = (baseUrl || import.meta.env.VITE_STORAGE_URL || 'http://localhost:3001').replace(/\/$/, ''); // Remove trailing slash
  }

  /**
   * Upload a file to Filecoin via SynapseSDK
   */
  async uploadFile(file: File): Promise<UploadResult> {
    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await fetch(`${this.baseUrl}/api/storage/upload`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `Upload failed with status ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('File upload failed:', error);
      throw error;
    }
  }

  /**
   * Upload raw data to Filecoin
   */
  async uploadData(data: ArrayBuffer, filename: string, contentType: string = 'application/octet-stream'): Promise<UploadResult> {
    const file = new File([data], filename, { type: contentType });
    return this.uploadFile(file);
  }

  /**
   * Retrieve a file from Filecoin by CID
   */
  async retrieveFile(cid: string): Promise<RetrieveResult> {
    try {
      const response = await fetch(`${this.baseUrl}/api/storage/retrieve/${cid}`);

      if (!response.ok) {
        if (response.status === 404) {
          throw new Error(`File not found: ${cid}`);
        }
        const error = await response.json();
        throw new Error(error.error || `Retrieval failed with status ${response.status}`);
      }

      const result = await response.json();
      
      // Convert base64 data back to ArrayBuffer if needed
      if (typeof result.data === 'string') {
        const binaryString = atob(result.data);
        const data = new ArrayBuffer(binaryString.length);
        const view = new Uint8Array(data);
        for (let i = 0; i < binaryString.length; i++) {
          view[i] = binaryString.charCodeAt(i);
        }
        result.data = data;
      }

      return result;
    } catch (error) {
      console.error('File retrieval failed:', error);
      throw error;
    }
  }

  /**
   * Download a file and trigger browser download
   */
  async downloadFile(cid: string, filename?: string): Promise<void> {
    try {
      const result = await this.retrieveFile(cid);
      
      const blob = new Blob([result.data], { 
        type: result.contentType || 'application/octet-stream' 
      });
      
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename || result.filename || `file-${cid.slice(0, 8)}.bin`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('File download failed:', error);
      throw error;
    }
  }

  /**
   * Get cost estimate for storing data
   */
  async getCostEstimate(sizeBytes: number): Promise<CostEstimate> {
    try {
      const response = await fetch(`${this.baseUrl}/api/storage/cost/${sizeBytes}`);
      
      if (!response.ok) {
        throw new Error(`Cost estimate failed with status ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Cost estimate failed:', error);
      throw error;
    }
  }

  /**
   * List stored files
   */
  async listFiles(): Promise<FileInfo[]> {
    try {
      const response = await fetch(`${this.baseUrl}/api/storage/files`);
      
      if (!response.ok) {
        throw new Error(`List files failed with status ${response.status}`);
      }

      const result = await response.json();
      return result.files || [];
    } catch (error) {
      console.error('List files failed:', error);
      throw error;
    }
  }

  /**
   * Pin a file to IPFS
   */
  async pinToIPFS(cid: string): Promise<void> {
    try {
      const response = await fetch(`${this.baseUrl}/api/storage/pin/${cid}`, {
        method: 'POST',
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `Pin failed with status ${response.status}`);
      }
    } catch (error) {
      console.error('Pin to IPFS failed:', error);
      throw error;
    }
  }

  /**
   * Get storage deal status
   */
  async getDealStatus(dealId: string): Promise<DealStatus> {
    try {
      const response = await fetch(`${this.baseUrl}/api/storage/deal-status/${dealId}`);
      
      if (!response.ok) {
        if (response.status === 404) {
          throw new Error(`Deal not found: ${dealId}`);
        }
        const error = await response.json();
        throw new Error(error.error || `Deal status failed with status ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Deal status check failed:', error);
      throw error;
    }
  }

  /**
   * Get Filecoin network information
   */
  async getNetworkInfo(): Promise<NetworkInfo> {
    try {
      const response = await fetch(`${this.baseUrl}/api/storage/network/info`);
      
      if (!response.ok) {
        throw new Error(`Network info failed with status ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Network info failed:', error);
      throw error;
    }
  }

  /**
   * Verify file integrity and metadata
   */
  async verifyFile(cid: string): Promise<VerificationResult> {
    try {
      const response = await fetch(`${this.baseUrl}/verify/${cid}`);
      
      if (!response.ok) {
        throw new Error(`Verification failed: ${response.status} ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      console.error('File verification failed:', error);
      throw error;
    }
  }

  /**
   * Check if the storage service is available
   */
  async isServiceAvailable(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/health`);
      return response.ok;
    } catch (error) {
      console.error('Storage service health check failed:', error);
      return false;
    }
  }

  /**
   * Upload a payment receipt with metadata
   */
  async uploadReceipt(receiptData: any, paymentId: string): Promise<UploadResult> {
    const jsonData = JSON.stringify(receiptData);
    const encoder = new TextEncoder();
    const data = encoder.encode(jsonData);
    
    return this.uploadData(
      data.buffer, 
      `receipt-${paymentId}.json`,
      'application/json'
    );
  }

  /**
   * Format file size for display
   */
  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  /**
   * Check if a CID is valid IPFS format
   */
  isValidCID(cid: string): boolean {
    // Basic CID validation - starts with common IPFS prefixes
    return /^(Qm[1-9A-HJ-NP-Za-km-z]{44}|baf[a-z0-9]{50,})$/.test(cid);
  }

  /**
   * Generate a shareable IPFS gateway URL
   */
  getIPFSGatewayUrl(cid: string, gateway: string = 'https://gateway.ipfs.io'): string {
    return `${gateway}/ipfs/${cid}`;
  }
}

// Export singleton instance
export const storageService = new StorageService();

// Export types for use in components
export type { 
  UploadResult, 
  RetrieveResult, 
  FileInfo, 
  DealStatus, 
  CostEstimate, 
  NetworkInfo,
  VerificationResult 
};