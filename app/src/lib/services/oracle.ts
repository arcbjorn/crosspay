interface PriceData {
  symbol: string;
  price: number;
  timestamp: number;
  decimals: number;
  valid: boolean;
}

interface RandomNumberRequest {
  requestId: string;
  timestamp: number;
  fulfilled: boolean;
  seed?: string;
}

interface ProofData {
  proofId: string;
  merkleRoot: string;
  timestamp: number;
  verified: boolean;
  data: string;
}

interface OracleHealthStatus {
  healthy: boolean;
  lastHealthCheck: number;
  ftsoHealthy: boolean;
  randomHealthy: boolean;
  fdcHealthy: boolean;
}

export class OracleService {
  private baseUrl: string;
  private wsConnection: WebSocket | null = null;
  private priceSubscriptions: Map<string, Set<(price: PriceData) => void>> = new Map();

  constructor(baseUrl?: string) {
    this.baseUrl = baseUrl || import.meta.env.VITE_ORACLE_URL || 'http://localhost:3003';
  }

  /**
   * Get current price for a trading pair
   */
  async getCurrentPrice(symbol: string): Promise<PriceData | null> {
    try {
      const response = await fetch(`${this.baseUrl}/ftso/price/${symbol}`);
      
      if (!response.ok) {
        console.error('Price fetch failed:', response.statusText);
        return null;
      }

      return await response.json();
    } catch (error) {
      console.error('Price fetch request failed:', error);
      return null;
    }
  }

  /**
   * Get multiple prices at once
   */
  async getPrices(symbols: string[]): Promise<Record<string, PriceData | null>> {
    try {
      const response = await fetch(`${this.baseUrl}/ftso/prices`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ symbols }),
      });

      if (!response.ok) {
        console.error('Batch price fetch failed:', response.statusText);
        return {};
      }

      const data = await response.json();
      return data.prices || {};
    } catch (error) {
      console.error('Batch price fetch request failed:', error);
      return {};
    }
  }

  /**
   * Get historical price at specific timestamp
   */
  async getHistoricalPrice(symbol: string, timestamp: number): Promise<PriceData | null> {
    try {
      const response = await fetch(`${this.baseUrl}/ftso/price/${symbol}/${timestamp}`);
      
      if (!response.ok) {
        console.error('Historical price fetch failed:', response.statusText);
        return null;
      }

      return await response.json();
    } catch (error) {
      console.error('Historical price fetch request failed:', error);
      return null;
    }
  }

  /**
   * Request a random number from Flare RNG
   */
  async requestRandomNumber(): Promise<RandomNumberRequest | null> {
    try {
      const response = await fetch(`${this.baseUrl}/rng/request`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        console.error('Random number request failed:', response.statusText);
        return null;
      }

      return await response.json();
    } catch (error) {
      console.error('Random number request failed:', error);
      return null;
    }
  }

  /**
   * Get status of a random number request
   */
  async getRandomNumberStatus(requestId: string): Promise<RandomNumberRequest | null> {
    try {
      const response = await fetch(`${this.baseUrl}/rng/status/${requestId}`);
      
      if (!response.ok) {
        console.error('Random number status fetch failed:', response.statusText);
        return null;
      }

      return await response.json();
    } catch (error) {
      console.error('Random number status fetch request failed:', error);
      return null;
    }
  }

  /**
   * Submit external proof for FDC verification
   */
  async submitProof(
    proofId: string,
    merkleRoot: string,
    proof: string[],
    data: string
  ): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/fdc/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          proofId,
          merkleRoot,
          proof,
          data,
        }),
      });

      if (!response.ok) {
        console.error('Proof submission failed:', response.statusText);
        return false;
      }

      const result = await response.json();
      return result.success || false;
    } catch (error) {
      console.error('Proof submission request failed:', error);
      return false;
    }
  }

  /**
   * Verify an external proof
   */
  async verifyProof(proofId: string): Promise<ProofData | null> {
    try {
      const response = await fetch(`${this.baseUrl}/fdc/verify/${proofId}`);
      
      if (!response.ok) {
        console.error('Proof verification failed:', response.statusText);
        return null;
      }

      return await response.json();
    } catch (error) {
      console.error('Proof verification request failed:', error);
      return null;
    }
  }

  /**
   * Get oracle health status
   */
  async getHealthStatus(): Promise<OracleHealthStatus | null> {
    try {
      const response = await fetch(`${this.baseUrl}/health`);
      
      if (!response.ok) {
        console.error('Health status fetch failed:', response.statusText);
        return null;
      }

      return await response.json();
    } catch (error) {
      console.error('Health status request failed:', error);
      return null;
    }
  }

  /**
   * Subscribe to real-time price updates via WebSocket
   */
  subscribeToPriceUpdates(symbol: string, callback: (price: PriceData) => void): () => void {
    // Add callback to subscription map
    if (!this.priceSubscriptions.has(symbol)) {
      this.priceSubscriptions.set(symbol, new Set());
    }
    this.priceSubscriptions.get(symbol)!.add(callback);

    // Initialize WebSocket connection if needed
    if (!this.wsConnection) {
      this.initializeWebSocket();
    }

    // Send subscription message
    if (this.wsConnection?.readyState === WebSocket.OPEN) {
      this.wsConnection.send(JSON.stringify({
        type: 'subscribe',
        symbol,
      }));
    }

    // Return unsubscribe function
    return () => {
      const callbacks = this.priceSubscriptions.get(symbol);
      if (callbacks) {
        callbacks.delete(callback);
        if (callbacks.size === 0) {
          this.priceSubscriptions.delete(symbol);
          // Send unsubscribe message
          if (this.wsConnection?.readyState === WebSocket.OPEN) {
            this.wsConnection.send(JSON.stringify({
              type: 'unsubscribe',
              symbol,
            }));
          }
        }
      }
    };
  }

  /**
   * Initialize WebSocket connection for real-time updates
   */
  private initializeWebSocket() {
    const wsUrl = this.baseUrl.replace('http', 'ws') + '/ws';
    
    try {
      this.wsConnection = new WebSocket(wsUrl);

      this.wsConnection.onopen = () => {
        console.log('Oracle WebSocket connected');
        
        // Resubscribe to all active symbols
        for (const symbol of this.priceSubscriptions.keys()) {
          this.wsConnection!.send(JSON.stringify({
            type: 'subscribe',
            symbol,
          }));
        }
      };

      this.wsConnection.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          
          if (data.type === 'priceUpdate' && data.price) {
            const callbacks = this.priceSubscriptions.get(data.price.symbol);
            if (callbacks) {
              callbacks.forEach(callback => callback(data.price));
            }
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.wsConnection.onclose = () => {
        console.log('Oracle WebSocket disconnected');
        this.wsConnection = null;
        
        // Attempt to reconnect after 5 seconds if we have active subscriptions
        if (this.priceSubscriptions.size > 0) {
          setTimeout(() => this.initializeWebSocket(), 5000);
        }
      };

      this.wsConnection.onerror = (error) => {
        console.error('Oracle WebSocket error:', error);
      };

    } catch (error) {
      console.error('Failed to initialize WebSocket:', error);
    }
  }

  /**
   * Close WebSocket connection
   */
  disconnect() {
    if (this.wsConnection) {
      this.wsConnection.close();
      this.wsConnection = null;
    }
    this.priceSubscriptions.clear();
  }

  /**
   * Format price for display
   */
  formatPrice(price: number, decimals: number = 2): string {
    return price.toFixed(decimals);
  }

  /**
   * Check if price data is stale (older than 10 minutes)
   */
  isPriceStale(timestamp: number): boolean {
    const tenMinutes = 10 * 60 * 1000; // 10 minutes in milliseconds
    return Date.now() - timestamp > tenMinutes;
  }

  /**
   * Get supported trading pairs
   */
  getSupportedPairs(): string[] {
    return ['ETH/USD', 'BTC/USD', 'FLR/USD', 'USDC/USD'];
  }
}

// Export singleton instance
export const oracleService = new OracleService();

// Export types for use in components
export type { PriceData, RandomNumberRequest, ProofData, OracleHealthStatus };