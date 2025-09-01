export interface RiskFactor {
  name: string;
  impact: number; // -100 to 100
  description: string;
  confidence: number; // 0 to 1
}

export interface TransactionData {
  amount: number;
  sender: string;
  recipient: string;
  network: string;
  timestamp: number;
  gasPrice?: number;
  memo?: string;
}

export interface RiskAssessment {
  score: number; // 0-100
  level: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  factors: RiskFactor[];
  recommendation: string;
  processingTime: number;
}

class RiskScoringAI {
  private knownBadActors = new Set([
    '0x0000000000000000000000000000000000000000',
    // Add known malicious addresses
  ]);
  
  private highRiskPatterns = [
    /^0x[0]{38,}[1-9a-f]$/i, // Suspicious zero-padded addresses
    /^0x[1]{40}$/i,          // All ones address
  ];
  
  async analyzeTransaction(data: TransactionData): Promise<RiskAssessment> {
    const startTime = Date.now();
    const factors: RiskFactor[] = [];
    
    // Amount-based risk factors
    factors.push(...this.analyzeAmount(data.amount, data.network));
    
    // Address reputation analysis
    factors.push(...this.analyzeAddresses(data.sender, data.recipient));
    
    // Network-specific risks
    factors.push(...this.analyzeNetwork(data.network));
    
    // Time-based patterns
    factors.push(...this.analyzeTimingPatterns(data.timestamp));
    
    // Gas price analysis (if available)
    if (data.gasPrice) {
      factors.push(...this.analyzeGasPrice(data.gasPrice));
    }
    
    // Memo analysis
    if (data.memo) {
      factors.push(...this.analyzeMemo(data.memo));
    }
    
    // Calculate overall risk score
    const score = this.calculateRiskScore(factors);
    const level = this.getRiskLevel(score);
    const recommendation = this.getRecommendation(score, factors);
    
    return {
      score,
      level,
      factors: factors.sort((a, b) => Math.abs(b.impact) - Math.abs(a.impact)),
      recommendation,
      processingTime: Date.now() - startTime
    };
  }
  
  private analyzeAmount(amount: number, network: string): RiskFactor[] {
    const factors: RiskFactor[] = [];
    
    // Very high amount risk
    if (amount > 10000) {
      factors.push({
        name: 'HIGH_VALUE_TRANSACTION',
        impact: 30,
        description: 'Transaction amount exceeds high-value threshold',
        confidence: 0.9
      });
    }
    
    // Unusual round numbers
    if (amount % 1000 === 0 && amount > 1000) {
      factors.push({
        name: 'ROUND_NUMBER_AMOUNT',
        impact: 10,
        description: 'Perfect round number amounts can indicate scripted attacks',
        confidence: 0.6
      });
    }
    
    // Dust transaction
    if (amount < 0.001) {
      factors.push({
        name: 'DUST_TRANSACTION',
        impact: 5,
        description: 'Very small amounts often used for address probing',
        confidence: 0.7
      });
    }
    
    return factors;
  }
  
  private analyzeAddresses(sender: string, recipient: string): RiskFactor[] {
    const factors: RiskFactor[] = [];
    
    // Check against known bad actors
    if (this.knownBadActors.has(recipient.toLowerCase())) {
      factors.push({
        name: 'BLACKLISTED_RECIPIENT',
        impact: 80,
        description: 'Recipient address is on known malicious actors list',
        confidence: 0.95
      });
    }
    
    if (this.knownBadActors.has(sender.toLowerCase())) {
      factors.push({
        name: 'BLACKLISTED_SENDER',
        impact: 70,
        description: 'Sender address is on known malicious actors list',
        confidence: 0.95
      });
    }
    
    // Check for suspicious patterns
    for (const pattern of this.highRiskPatterns) {
      if (pattern.test(recipient)) {
        factors.push({
          name: 'SUSPICIOUS_RECIPIENT_PATTERN',
          impact: 40,
          description: 'Recipient address matches suspicious pattern',
          confidence: 0.8
        });
      }
    }
    
    // Self-transaction (same sender and recipient)
    if (sender.toLowerCase() === recipient.toLowerCase()) {
      factors.push({
        name: 'SELF_TRANSACTION',
        impact: 15,
        description: 'Transaction to same address (potential token manipulation)',
        confidence: 0.9
      });
    }
    
    // New address risk (simplified check)
    if (recipient.endsWith('000000')) {
      factors.push({
        name: 'POTENTIALLY_NEW_ADDRESS',
        impact: 20,
        description: 'Recipient may be a newly created address',
        confidence: 0.5
      });
    }
    
    return factors;
  }
  
  private analyzeNetwork(network: string): RiskFactor[] {
    const factors: RiskFactor[] = [];
    
    const networkRisk = {
      'ethereum': -10,
      'base': -5,
      'polygon': 5,
      'bsc': 10,
      'testnet': 20
    }[network.toLowerCase()] || 0;
    
    if (networkRisk !== 0) {
      factors.push({
        name: 'NETWORK_RISK_PROFILE',
        impact: networkRisk,
        description: `${network} network has ${networkRisk > 0 ? 'higher' : 'lower'} risk profile`,
        confidence: 0.7
      });
    }
    
    return factors;
  }
  
  private analyzeTimingPatterns(timestamp: number): RiskFactor[] {
    const factors: RiskFactor[] = [];
    const now = Date.now();
    const hour = new Date(timestamp).getHours();
    
    // Off-hours transaction (potential automated attack)
    if (hour < 6 || hour > 22) {
      factors.push({
        name: 'OFF_HOURS_TRANSACTION',
        impact: 5,
        description: 'Transaction during unusual hours (automated activity)',
        confidence: 0.4
      });
    }
    
    // Future-dated transaction
    if (timestamp > now + 300000) { // 5 minutes in future
      factors.push({
        name: 'FUTURE_TIMESTAMP',
        impact: 25,
        description: 'Transaction timestamp is in the future',
        confidence: 0.9
      });
    }
    
    return factors;
  }
  
  private analyzeGasPrice(gasPrice: number): RiskFactor[] {
    const factors: RiskFactor[] = [];
    
    // Extremely high gas price (potential front-running)
    if (gasPrice > 100) {
      factors.push({
        name: 'EXCESSIVE_GAS_PRICE',
        impact: 20,
        description: 'Unusually high gas price may indicate MEV or front-running',
        confidence: 0.8
      });
    }
    
    // Extremely low gas price (potential stuck transaction)
    if (gasPrice < 1) {
      factors.push({
        name: 'VERY_LOW_GAS_PRICE',
        impact: 10,
        description: 'Very low gas price may result in failed transaction',
        confidence: 0.7
      });
    }
    
    return factors;
  }
  
  private analyzeMemo(memo: string): RiskFactor[] {
    const factors: RiskFactor[] = [];
    
    // Suspicious keywords in memo
    const suspiciousKeywords = ['hack', 'exploit', 'drain', 'rugpull', 'scam'];
    const lowerMemo = memo.toLowerCase();
    
    for (const keyword of suspiciousKeywords) {
      if (lowerMemo.includes(keyword)) {
        factors.push({
          name: 'SUSPICIOUS_MEMO_CONTENT',
          impact: 30,
          description: `Memo contains suspicious keyword: ${keyword}`,
          confidence: 0.8
        });
        break; // Only add once
      }
    }
    
    // Very long memo (potential data exfiltration)
    if (memo.length > 500) {
      factors.push({
        name: 'EXCESSIVE_MEMO_LENGTH',
        impact: 15,
        description: 'Unusually long memo may contain hidden data',
        confidence: 0.6
      });
    }
    
    return factors;
  }
  
  private calculateRiskScore(factors: RiskFactor[]): number {
    const baseScore = 10; // Base risk for any transaction
    let totalImpact = 0;
    let totalConfidence = 0;
    
    for (const factor of factors) {
      const weightedImpact = factor.impact * factor.confidence;
      totalImpact += weightedImpact;
      totalConfidence += Math.abs(factor.confidence);
    }
    
    // Normalize confidence
    const avgConfidence = totalConfidence > 0 ? totalConfidence / factors.length : 0.5;
    
    // Calculate final score
    let score = baseScore + totalImpact;
    
    // Apply confidence weighting
    score = score * (0.5 + avgConfidence * 0.5);
    
    // Ensure score is within 0-100 range
    return Math.max(0, Math.min(100, Math.round(score)));
  }
  
  private getRiskLevel(score: number): 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL' {
    if (score < 25) return 'LOW';
    if (score < 50) return 'MEDIUM';
    if (score < 75) return 'HIGH';
    return 'CRITICAL';
  }
  
  private getRecommendation(score: number, factors: RiskFactor[]): string {
    const highestRiskFactor = factors.reduce((prev, current) => 
      Math.abs(current.impact) > Math.abs(prev.impact) ? current : prev, 
      { impact: 0, name: '', description: '', confidence: 0 }
    );
    
    if (score < 25) {
      return 'PROCEED: Transaction appears safe with minimal risk factors.';
    } else if (score < 50) {
      return `CAUTION: Moderate risk detected. Primary concern: ${highestRiskFactor.name}. Review transaction details before proceeding.`;
    } else if (score < 75) {
      return `WARNING: High risk transaction. Major concern: ${highestRiskFactor.name}. Additional verification strongly recommended.`;
    } else {
      return `CRITICAL: Extremely high risk detected. Primary threat: ${highestRiskFactor.name}. Transaction should be blocked pending manual review.`;
    }
  }
}

export const riskScoringAI = new RiskScoringAI();