import type { Address } from 'viem';

interface ENSResolutionResponse {
  address: string;
  name?: string;
  avatar?: string;
  email?: string;
  url?: string;
  github?: string;
  twitter?: string;
  error?: string;
}

interface ENSReverseResponse {
  name: string;
  address: string;
  error?: string;
}

interface SubnameResponse {
  subname: string;
  domain: string;
  owner: string;
  registrationFee: string;
  success: boolean;
  error?: string;
}

export class ENSService {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:3002') {
    this.baseUrl = baseUrl;
  }

  /**
   * Resolve ENS name to Ethereum address
   */
  async resolveName(ensName: string): Promise<Address | null> {
    if (!ensName.endsWith('.eth')) {
      return null;
    }

    try {
      const response = await fetch(`${this.baseUrl}/resolve/${encodeURIComponent(ensName)}`);
      
      if (!response.ok) {
        console.error('ENS resolution failed:', response.statusText);
        return null;
      }

      const data: ENSResolutionResponse = await response.json();
      
      if (data.error) {
        console.error('ENS resolution error:', data.error);
        return null;
      }

      return data.address as Address;
    } catch (error) {
      console.error('ENS resolution request failed:', error);
      return null;
    }
  }

  /**
   * Reverse resolve Ethereum address to ENS name
   */
  async lookupAddress(address: Address): Promise<string | null> {
    if (!address || !address.startsWith('0x')) {
      return null;
    }

    try {
      const response = await fetch(`${this.baseUrl}/reverse/${address}`);
      
      if (!response.ok) {
        console.error('ENS reverse lookup failed:', response.statusText);
        return null;
      }

      const data: ENSReverseResponse = await response.json();
      
      if (data.error) {
        console.error('ENS reverse lookup error:', data.error);
        return null;
      }

      return data.name;
    } catch (error) {
      console.error('ENS reverse lookup request failed:', error);
      return null;
    }
  }

  /**
   * Get full ENS profile information
   */
  async getProfile(ensName: string): Promise<ENSResolutionResponse | null> {
    if (!ensName.endsWith('.eth')) {
      return null;
    }

    try {
      const response = await fetch(`${this.baseUrl}/resolve/${encodeURIComponent(ensName)}`);
      
      if (!response.ok) {
        console.error('ENS profile lookup failed:', response.statusText);
        return null;
      }

      const data: ENSResolutionResponse = await response.json();
      
      if (data.error) {
        console.error('ENS profile error:', data.error);
        return null;
      }

      return data;
    } catch (error) {
      console.error('ENS profile request failed:', error);
      return null;
    }
  }

  /**
   * Register a subname under a domain
   */
  async registerSubname(
    subname: string, 
    domain: string, 
    owner: Address,
    signerAddress: Address
  ): Promise<SubnameResponse | null> {
    try {
      const response = await fetch(`${this.baseUrl}/register-subname`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          subname,
          domain,
          owner,
          signer: signerAddress,
        }),
      });

      if (!response.ok) {
        console.error('Subname registration failed:', response.statusText);
        return null;
      }

      const data: SubnameResponse = await response.json();
      return data;
    } catch (error) {
      console.error('Subname registration request failed:', error);
      return null;
    }
  }

  /**
   * Batch resolve multiple ENS names
   */
  async batchResolve(ensNames: string[]): Promise<Record<string, Address | null>> {
    const validNames = ensNames.filter(name => name.endsWith('.eth'));
    
    if (validNames.length === 0) {
      return {};
    }

    try {
      const response = await fetch(`${this.baseUrl}/batch-resolve`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ names: validNames }),
      });

      if (!response.ok) {
        console.error('Batch ENS resolution failed:', response.statusText);
        return {};
      }

      const data = await response.json();
      return data.results || {};
    } catch (error) {
      console.error('Batch ENS resolution request failed:', error);
      return {};
    }
  }

  /**
   * Check if ENS service is available
   */
  async isServiceAvailable(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/health`, {
        method: 'GET',
        headers: { 'Accept': 'application/json' },
      });
      return response.ok;
    } catch (error) {
      console.error('ENS service health check failed:', error);
      return false;
    }
  }

  /**
   * Validate ENS name format
   */
  isValidENSName(name: string): boolean {
    return /^[a-zA-Z0-9-]+\.eth$/.test(name);
  }

  /**
   * Format address for display (truncated)
   */
  formatAddress(address: Address): string {
    if (!address || address.length < 10) return address;
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  }
}

// Export singleton instance
export const ensService = new ENSService();

// Export types for use in components
export type { ENSResolutionResponse, ENSReverseResponse, SubnameResponse };