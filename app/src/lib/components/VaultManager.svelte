<script lang="ts">
  import { writable } from 'svelte/store';
  
  export let userAddress: string = '';
  
  interface TrancheInfo {
    name: string;
    risk: 'low' | 'medium' | 'high';
    apy: number;
    tvl: bigint;
    userDeposit: bigint;
    withdrawDelay: number;
    color: string;
    description: string;
  }

  interface VaultMetrics {
    totalTVL: bigint;
    utilizationRate: number;
    insuranceFund: bigint;
    totalYieldDistributed: bigint;
    performanceFee: number;
  }

  let selectedTranche: keyof typeof tranches | null = null;
  let depositAmount = '';
  let withdrawAmount = '';
  let loading = false;
  let activeTab: 'deposit' | 'withdraw' | 'metrics' = 'deposit';
  
  const error = writable<string>('');

  // Mock data - would be fetched from contracts
  const tranches: Record<string, TrancheInfo> = {
    senior: {
      name: 'Senior Tranche',
      risk: 'low',
      apy: 8.5,
      tvl: BigInt('50000000000000000000000'), // 50k ETH
      userDeposit: BigInt('5000000000000000000000'), // 5k ETH
      withdrawDelay: 24,
      color: '#28a745',
      description: 'Lowest risk, first in line for yields and last to absorb losses'
    },
    mezzanine: {
      name: 'Mezzanine Tranche',
      risk: 'medium',
      apy: 12.8,
      tvl: BigInt('25000000000000000000000'), // 25k ETH
      userDeposit: BigInt('2000000000000000000000'), // 2k ETH
      withdrawDelay: 48,
      color: '#ffc107',
      description: 'Moderate risk, balanced yield and loss absorption'
    },
    junior: {
      name: 'Junior Tranche',
      risk: 'high',
      apy: 18.2,
      tvl: BigInt('10000000000000000000000'), // 10k ETH
      userDeposit: BigInt('500000000000000000000'), // 0.5k ETH
      withdrawDelay: 72,
      color: '#dc3545',
      description: 'Highest risk and yield, first to absorb losses'
    }
  };

  const vaultMetrics: VaultMetrics = {
    totalTVL: BigInt('85000000000000000000000'), // 85k ETH
    utilizationRate: 73.5,
    insuranceFund: BigInt('2500000000000000000000'), // 2.5k ETH
    totalYieldDistributed: BigInt('8750000000000000000000'), // 8.75k ETH
    performanceFee: 15 // 15%
  };

  function formatEth(wei: bigint): string {
    return (Number(wei) / 1e18).toLocaleString(undefined, { maximumFractionDigits: 2 });
  }

  function getRiskColor(risk: 'low' | 'medium' | 'high'): string {
    return risk === 'low' ? '#28a745' : risk === 'medium' ? '#ffc107' : '#dc3545';
  }

  async function deposit() {
    if (!selectedTranche || !depositAmount) {
      error.set('Please select a tranche and enter deposit amount');
      return;
    }

    loading = true;
    error.set('');

    try {
      const amountWei = BigInt(parseFloat(depositAmount) * 1e18);
      
      console.log('Depositing to vault:', {
        tranche: selectedTranche,
        amount: amountWei.toString(),
        userAddress
      });

      // Mock contract call
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      // Update local state for demo
      tranches[selectedTranche].userDeposit += amountWei;
      tranches[selectedTranche].tvl += amountWei;
      
      depositAmount = '';
      
    } catch (e) {
      error.set(`Deposit failed: ${e}`);
    } finally {
      loading = false;
    }
  }

  async function withdraw() {
    if (!selectedTranche || !withdrawAmount) {
      error.set('Please select a tranche and enter withdrawal amount');
      return;
    }

    loading = true;
    error.set('');

    try {
      const amountWei = BigInt(parseFloat(withdrawAmount) * 1e18);
      
      if (amountWei > tranches[selectedTranche].userDeposit) {
        throw new Error('Insufficient balance');
      }
      
      console.log('Withdrawing from vault:', {
        tranche: selectedTranche,
        amount: amountWei.toString(),
        userAddress
      });

      // Mock contract call
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      // Update local state for demo
      tranches[selectedTranche].userDeposit -= amountWei;
      tranches[selectedTranche].tvl -= amountWei;
      
      withdrawAmount = '';
      
    } catch (e) {
      error.set(`Withdrawal failed: ${e}`);
    } finally {
      loading = false;
    }
  }
</script>

<div class="vault-manager">
  <h2>Tranche Vault Management</h2>
  
  <!-- Vault Overview -->
  <div class="vault-overview">
    <div class="metric-card">
      <h4>Total Value Locked</h4>
      <div class="metric-value">{formatEth(vaultMetrics.totalTVL)} ETH</div>
    </div>
    
    <div class="metric-card">
      <h4>Utilization Rate</h4>
      <div class="metric-value">{vaultMetrics.utilizationRate}%</div>
    </div>
    
    <div class="metric-card">
      <h4>Insurance Fund</h4>
      <div class="metric-value">{formatEth(vaultMetrics.insuranceFund)} ETH</div>
    </div>
    
    <div class="metric-card">
      <h4>Total Yield Distributed</h4>
      <div class="metric-value">{formatEth(vaultMetrics.totalYieldDistributed)} ETH</div>
    </div>
  </div>

  <!-- Tab Navigation -->
  <div class="tab-nav">
    <button 
      class="tab" 
      class:active={activeTab === 'deposit'}
      on:click={() => activeTab = 'deposit'}
    >
      💰 Deposit
    </button>
    <button 
      class="tab"
      class:active={activeTab === 'withdraw'}
      on:click={() => activeTab = 'withdraw'}
    >
      💸 Withdraw
    </button>
    <button 
      class="tab"
      class:active={activeTab === 'metrics'}
      on:click={() => activeTab = 'metrics'}
    >
      📊 Analytics
    </button>
  </div>

  <!-- Tranche Selection -->
  <div class="tranche-selection">
    <h3>Select Tranche</h3>
    <div class="tranche-grid">
      {#each Object.entries(tranches) as [key, tranche]}
        <div 
          class="tranche-card" 
          class:selected={selectedTranche === key}
          style="border-color: {tranche.color}"
          on:click={() => selectedTranche = key}
        >
          <div class="tranche-header">
            <h4>{tranche.name}</h4>
            <span 
              class="risk-badge" 
              style="background: {getRiskColor(tranche.risk)}; color: white;"
            >
              {tranche.risk.toUpperCase()} RISK
            </span>
          </div>
          
          <div class="tranche-metrics">
            <div class="metric">
              <span>APY:</span>
              <strong style="color: {tranche.color}">{tranche.apy}%</strong>
            </div>
            <div class="metric">
              <span>TVL:</span>
              <strong>{formatEth(tranche.tvl)} ETH</strong>
            </div>
            <div class="metric">
              <span>Your Position:</span>
              <strong>{formatEth(tranche.userDeposit)} ETH</strong>
            </div>
            <div class="metric">
              <span>Withdraw Delay:</span>
              <strong>{tranche.withdrawDelay}h</strong>
            </div>
          </div>
          
          <div class="tranche-description">
            {tranche.description}
          </div>
        </div>
      {/each}
    </div>
  </div>

  <!-- Action Panel -->
  {#if activeTab === 'deposit'}
    <div class="action-panel">
      <h3>Deposit to Vault</h3>
      
      {#if selectedTranche}
        <div class="selected-tranche">
          Selected: <strong>{tranches[selectedTranche].name}</strong>
          <span style="color: {tranches[selectedTranche].color}">
            ({tranches[selectedTranche].apy}% APY)
          </span>
        </div>
        
        <div class="input-group">
          <label for="deposit-amount">Deposit Amount (ETH)</label>
          <input
            id="deposit-amount"
            type="number"
            bind:value={depositAmount}
            placeholder="1.0"
            step="0.01"
            min="0"
          />
        </div>
        
        <button
          on:click={deposit}
          disabled={loading || !depositAmount || !selectedTranche}
          class="action-button deposit"
        >
          {loading ? 'Processing...' : `Deposit ${depositAmount || '0'} ETH`}
        </button>
      {:else}
        <p>Please select a tranche above to deposit</p>
      {/if}
    </div>

  {:else if activeTab === 'withdraw'}
    <div class="action-panel">
      <h3>Withdraw from Vault</h3>
      
      {#if selectedTranche && tranches[selectedTranche].userDeposit > 0n}
        <div class="selected-tranche">
          Selected: <strong>{tranches[selectedTranche].name}</strong>
          <span>
            Available: {formatEth(tranches[selectedTranche].userDeposit)} ETH
          </span>
        </div>
        
        <div class="withdraw-warning">
          ⚠️ Withdrawals require a {tranches[selectedTranche].withdrawDelay}h delay period
        </div>
        
        <div class="input-group">
          <label for="withdraw-amount">Withdrawal Amount (ETH)</label>
          <input
            id="withdraw-amount"
            type="number"
            bind:value={withdrawAmount}
            placeholder="1.0"
            step="0.01"
            min="0"
            max={Number(tranches[selectedTranche].userDeposit) / 1e18}
          />
        </div>
        
        <button
          on:click={withdraw}
          disabled={loading || !withdrawAmount || !selectedTranche}
          class="action-button withdraw"
        >
          {loading ? 'Processing...' : `Withdraw ${withdrawAmount || '0'} ETH`}
        </button>
      {:else}
        <p>No deposits to withdraw from selected tranche</p>
      {/if}
    </div>

  {:else if activeTab === 'metrics'}
    <div class="analytics-panel">
      <h3>Vault Analytics</h3>
      
      <div class="analytics-grid">
        <div class="analytics-card">
          <h4>Risk Distribution</h4>
          <div class="risk-chart">
            {#each Object.entries(tranches) as [key, tranche]}
              <div class="risk-bar">
                <span class="risk-label">{tranche.name}</span>
                <div class="risk-progress">
                  <div 
                    class="risk-fill"
                    style="width: {Number(tranche.tvl) / Number(vaultMetrics.totalTVL) * 100}%; background: {tranche.color};"
                  ></div>
                </div>
                <span class="risk-percentage">
                  {((Number(tranche.tvl) / Number(vaultMetrics.totalTVL)) * 100).toFixed(1)}%
                </span>
              </div>
            {/each}
          </div>
        </div>
        
        <div class="analytics-card">
          <h4>Performance Metrics</h4>
          <div class="performance-list">
            <div class="performance-item">
              <span>Performance Fee:</span>
              <strong>{vaultMetrics.performanceFee}%</strong>
            </div>
            <div class="performance-item">
              <span>Insurance Coverage:</span>
              <strong>{((Number(vaultMetrics.insuranceFund) / Number(vaultMetrics.totalTVL)) * 100).toFixed(1)}%</strong>
            </div>
            <div class="performance-item">
              <span>Yield Generated (24h):</span>
              <strong>+{formatEth(BigInt('125000000000000000000'))} ETH</strong>
            </div>
            <div class="performance-item">
              <span>Last Rebalance:</span>
              <strong>2 hours ago</strong>
            </div>
          </div>
        </div>
      </div>
    </div>
  {/if}

  {#if $error}
    <div class="error">{$error}</div>
  {/if}
</div>

<style>
  .vault-manager {
    max-width: 1000px;
    margin: 0 auto;
    padding: 2rem;
  }

  h2 {
    text-align: center;
    margin-bottom: 2rem;
    color: #333;
  }

  .vault-overview {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .metric-card {
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    padding: 1.5rem;
    text-align: center;
  }

  .metric-card h4 {
    margin: 0 0 0.5rem 0;
    color: #666;
    font-size: 0.9rem;
    text-transform: uppercase;
  }

  .metric-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: #333;
  }

  .tab-nav {
    display: flex;
    gap: 1rem;
    margin-bottom: 2rem;
    border-bottom: 2px solid #eee;
  }

  .tab {
    padding: 0.75rem 1.5rem;
    border: none;
    background: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.2s;
  }

  .tab.active {
    border-bottom-color: #007bff;
    color: #007bff;
  }

  .tab:hover {
    background: #f8f9fa;
  }

  .tranche-selection {
    margin-bottom: 2rem;
  }

  .tranche-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1rem;
  }

  .tranche-card {
    border: 2px solid #ddd;
    border-radius: 12px;
    padding: 1.5rem;
    cursor: pointer;
    transition: all 0.2s;
    background: white;
  }

  .tranche-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  }

  .tranche-card.selected {
    border-width: 3px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  }

  .tranche-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .risk-badge {
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .tranche-metrics {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .metric {
    display: flex;
    justify-content: space-between;
    font-size: 0.9rem;
  }

  .tranche-description {
    font-size: 0.85rem;
    color: #666;
    line-height: 1.4;
  }

  .action-panel {
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 12px;
    padding: 2rem;
  }

  .selected-tranche {
    background: #f8f9fa;
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1rem;
    text-align: center;
  }

  .withdraw-warning {
    background: #fff3cd;
    border: 1px solid #ffeaa7;
    padding: 0.75rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    color: #856404;
  }

  .input-group {
    margin-bottom: 1.5rem;
  }

  .input-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 600;
  }

  .input-group input {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ddd;
    border-radius: 8px;
    font-size: 1rem;
  }

  .action-button {
    width: 100%;
    padding: 1rem;
    border: none;
    border-radius: 8px;
    font-size: 1.1rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .action-button.deposit {
    background: #28a745;
    color: white;
  }

  .action-button.withdraw {
    background: #dc3545;
    color: white;
  }

  .action-button:hover:not(:disabled) {
    opacity: 0.9;
    transform: translateY(-1px);
  }

  .action-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
  }

  .analytics-panel {
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 12px;
    padding: 2rem;
  }

  .analytics-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
  }

  .analytics-card {
    border: 1px solid #eee;
    border-radius: 8px;
    padding: 1.5rem;
  }

  .risk-chart {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .risk-bar {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .risk-label {
    width: 100px;
    font-size: 0.85rem;
    font-weight: 500;
  }

  .risk-progress {
    flex: 1;
    height: 20px;
    background: #f0f0f0;
    border-radius: 10px;
    overflow: hidden;
  }

  .risk-fill {
    height: 100%;
    transition: width 0.3s ease;
  }

  .risk-percentage {
    width: 40px;
    text-align: right;
    font-size: 0.85rem;
    font-weight: 600;
  }

  .performance-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .performance-item {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem 0;
    border-bottom: 1px solid #f0f0f0;
  }

  .error {
    background: #f8d7da;
    border: 1px solid #f5c6cb;
    color: #721c24;
    padding: 1rem;
    border-radius: 6px;
    margin-top: 1rem;
  }

  @media (max-width: 768px) {
    .analytics-grid {
      grid-template-columns: 1fr;
    }
    
    .tranche-metrics {
      grid-template-columns: 1fr;
    }
  }
</style>