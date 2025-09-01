<script lang="ts">
  import CyberCard from './CyberCard.svelte';
  import { onMount } from 'svelte';
  
  export let validators: Array<{
    id: string;
    address: string;
    status: 'active' | 'inactive' | 'syncing';
    stake: number;
    uptime: number;
  }> = [];
  
  export let consensusStatus: 'finalized' | 'pending' | 'error' = 'pending';
  export let securityLevel: 'low' | 'medium' | 'high' | 'maximum' = 'high';
  
  let matrixCells = Array(64).fill(0).map((_, i) => ({
    id: i,
    active: false,
    type: Math.random() > 0.7 ? 'validator' : 'data',
    strength: Math.random()
  }));
  
  let animationFrame = 0;
  
  onMount(() => {
    const interval = setInterval(() => {
      matrixCells = matrixCells.map(cell => ({
        ...cell,
        active: Math.random() > 0.8,
        strength: Math.random()
      }));
      animationFrame++;
    }, 1000);
    
    return () => clearInterval(interval);
  });
  
  $: activeValidators = validators.filter(v => v.status === 'active').length;
  $: totalStake = validators.reduce((sum, v) => sum + v.stake, 0);
  $: averageUptime = validators.length > 0 ? 
    validators.reduce((sum, v) => sum + v.uptime, 0) / validators.length : 0;
    
  $: securityColor = {
    'low': 'text-cyber-error',
    'medium': 'text-cyber-warning', 
    'high': 'text-cyber-mint',
    'maximum': 'text-cyber-success'
  }[securityLevel];
  
  $: consensusColor = {
    'finalized': 'text-cyber-success',
    'pending': 'text-cyber-warning',
    'error': 'text-cyber-error'
  }[consensusStatus];
</script>

<CyberCard variant="default" padding="lg">
  <div class="terminal-text text-cyber-mint text-xl mb-6">
    [SECURITY_MATRIX]
  </div>
  
  <!-- Matrix Visualization -->
  <div class="grid grid-cols-8 gap-1 mb-6 p-4 bg-cyber-bg-secondary/50 border border-cyber-border-mint/20">
    {#each matrixCells as cell}
      <div 
        class="aspect-square border transition-all duration-500 {
          cell.active 
            ? cell.type === 'validator' 
              ? 'bg-cyber-success border-cyber-success animate-pulse' 
              : 'bg-cyber-mint border-cyber-mint'
            : 'bg-cyber-surface-2 border-cyber-text-tertiary/20'
        }"
        style="opacity: {cell.strength * 0.8 + 0.2}"
      ></div>
    {/each}
  </div>
  
  <!-- Status Grid -->
  <div class="grid md:grid-cols-2 gap-6 mb-6">
    <!-- Consensus Status -->
    <div class="terminal-text">
      <div class="text-cyber-text-secondary text-sm mb-2">CONSENSUS_STATUS:</div>
      <div class="{consensusColor} text-lg font-mono">
        [{consensusStatus.toUpperCase()}]
      </div>
      {#if consensusStatus === 'pending'}
        <div class="text-cyber-text-tertiary text-xs mt-1 animate-pulse">
          > Waiting for validator signatures...
        </div>
      {:else if consensusStatus === 'finalized'}
        <div class="text-cyber-text-tertiary text-xs mt-1">
          > Block finalized and verified
        </div>
      {:else}
        <div class="text-cyber-text-tertiary text-xs mt-1">
          > Consensus failure detected
        </div>
      {/if}
    </div>
    
    <!-- Security Level -->
    <div class="terminal-text">
      <div class="text-cyber-text-secondary text-sm mb-2">SECURITY_LEVEL:</div>
      <div class="{securityColor} text-lg font-mono">
        [{securityLevel.toUpperCase()}]
      </div>
      <div class="text-cyber-text-tertiary text-xs mt-1">
        > {activeValidators}/{validators.length} validators active
      </div>
    </div>
  </div>
  
  <!-- Validator Stats -->
  <div class="border-t border-cyber-border-mint/20 pt-4">
    <div class="terminal-text text-cyber-text-secondary text-sm mb-3">
      VALIDATOR_METRICS:
    </div>
    
    <div class="grid grid-cols-3 gap-4 text-center">
      <div class="cyber-card bg-cyber-surface-2 p-3">
        <div class="text-cyber-mint text-lg font-mono">
          {activeValidators}
        </div>
        <div class="text-cyber-text-tertiary text-xs">
          ACTIVE_NODES
        </div>
      </div>
      
      <div class="cyber-card bg-cyber-surface-2 p-3">
        <div class="text-cyber-lavender text-lg font-mono">
          {totalStake.toFixed(2)}K
        </div>
        <div class="text-cyber-text-tertiary text-xs">
          TOTAL_STAKE
        </div>
      </div>
      
      <div class="cyber-card bg-cyber-surface-2 p-3">
        <div class="text-cyber-success text-lg font-mono">
          {averageUptime.toFixed(1)}%
        </div>
        <div class="text-cyber-text-tertiary text-xs">
          AVG_UPTIME
        </div>
      </div>
    </div>
  </div>
  
  <!-- Individual Validators -->
  {#if validators.length > 0}
    <div class="mt-6 border-t border-cyber-border-mint/20 pt-4">
      <div class="terminal-text text-cyber-text-secondary text-sm mb-3">
        VALIDATOR_LIST:
      </div>
      
      <div class="space-y-2 max-h-32 overflow-y-auto">
        {#each validators.slice(0, 5) as validator}
          <div class="flex items-center justify-between text-xs terminal-text">
            <div class="flex items-center gap-3">
              <div class="w-2 h-2 rounded-full {
                validator.status === 'active' ? 'bg-cyber-success animate-pulse' : 
                validator.status === 'syncing' ? 'bg-cyber-warning' : 'bg-cyber-error'
              }"></div>
              <span class="text-cyber-text-primary font-mono">
                {validator.address.slice(0, 8)}...{validator.address.slice(-6)}
              </span>
            </div>
            
            <div class="flex gap-4 text-cyber-text-tertiary">
              <span>{validator.stake.toFixed(1)}K</span>
              <span>{validator.uptime.toFixed(1)}%</span>
              <span class="uppercase">{validator.status}</span>
            </div>
          </div>
        {/each}
        
        {#if validators.length > 5}
          <div class="text-cyber-text-tertiary text-xs text-center pt-2">
            ... and {validators.length - 5} more validators
          </div>
        {/if}
      </div>
    </div>
  {/if}
</CyberCard>