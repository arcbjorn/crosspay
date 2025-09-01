<script lang="ts">
  import CyberCard from './CyberCard.svelte';
  import { onMount } from 'svelte';
  
  export let riskScore: number = 0; // 0-100
  export let factors: Array<{
    name: string;
    impact: number; // -100 to 100
    description: string;
  }> = [];
  
  export let loading = false;
  
  let meterBars = 20;
  let animatedScore = 0;
  
  onMount(() => {
    const interval = setInterval(() => {
      if (animatedScore < riskScore) {
        animatedScore = Math.min(animatedScore + 2, riskScore);
      } else if (animatedScore > riskScore) {
        animatedScore = Math.max(animatedScore - 2, riskScore);
      }
    }, 50);
    
    return () => clearInterval(interval);
  });
  
  $: riskLevel = 
    riskScore < 25 ? 'LOW' : 
    riskScore < 50 ? 'MEDIUM' : 
    riskScore < 75 ? 'HIGH' : 'CRITICAL';
    
  $: riskColor = 
    riskScore < 25 ? 'text-cyber-success' : 
    riskScore < 50 ? 'text-cyber-warning' : 
    riskScore < 75 ? 'text-cyber-error' : 'text-red-400';
    
  $: activeBars = Math.ceil((animatedScore / 100) * meterBars);
</script>

<CyberCard variant="danger" padding="lg">
  <div class="terminal-text text-cyber-error text-xl mb-6">
    [RISK_ASSESSMENT_AI]
  </div>
  
  {#if loading}
    <div class="flex flex-col items-center py-8">
      <div class="ascii-loader text-cyber-mint mb-4">
        <pre class="animate-pulse">
{`
┌─────────────────────────────┐
│ ▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░ │
│ ANALYZING_RISK_FACTORS...   │
│ ░░░░▓▓▓▓░░░░░░░░░░░░░░░░░░░ │
└─────────────────────────────┘
`}
        </pre>
      </div>
      <div class="terminal-text text-cyber-text-secondary text-sm">
        > Processing transaction patterns...
      </div>
    </div>
  {:else}
    <!-- ASCII Risk Meter -->
    <div class="terminal-text text-center mb-6">
      <div class="text-cyber-text-secondary text-sm mb-2">RISK_SCORE:</div>
      <div class="{riskColor} text-3xl font-mono mb-2">
        {Math.round(animatedScore)}%
      </div>
      <div class="{riskColor} text-lg">
        [{riskLevel}]
      </div>
      
      <!-- ASCII Meter Visualization -->
      <div class="mt-4 font-mono text-xs">
        <div class="border border-cyber-border-mint p-2 bg-cyber-bg-secondary">
          ┌{'─'.repeat(meterBars + 2)}┐<br>
          │ {Array.from({length: meterBars}, (_, i) => {
            if (i < activeBars) {
              return i < meterBars * 0.25 ? '█' : 
                     i < meterBars * 0.5 ? '▓' :
                     i < meterBars * 0.75 ? '▒' : '░';
            }
            return '·';
          }).join('')} │<br>
          └{'─'.repeat(meterBars + 2)}┘<br>
          <span class="text-cyber-success">LOW</span>{'─'.repeat(Math.floor(meterBars/4))}
          <span class="text-cyber-warning">MED</span>{'─'.repeat(Math.floor(meterBars/4))}
          <span class="text-cyber-error">HIGH</span>{'─'.repeat(Math.floor(meterBars/4))}
          <span class="text-red-400">CRIT</span>
        </div>
      </div>
    </div>
    
    <!-- Risk Factors Analysis -->
    {#if factors.length > 0}
      <div class="border-t border-cyber-border-mint/20 pt-4">
        <div class="terminal-text text-cyber-text-secondary text-sm mb-3">
          RISK_FACTORS_ANALYSIS:
        </div>
        
        <div class="space-y-3">
          {#each factors as factor}
            <div class="cyber-card bg-cyber-surface-2 p-3">
              <div class="flex items-center justify-between mb-2">
                <span class="terminal-text text-cyber-text-primary text-sm">
                  {factor.name.toUpperCase()}
                </span>
                <span class="terminal-text {
                  factor.impact > 20 ? 'text-cyber-error' : 
                  factor.impact > 0 ? 'text-cyber-warning' : 
                  factor.impact > -20 ? 'text-cyber-mint' : 'text-cyber-success'
                } text-sm">
                  {factor.impact > 0 ? '+' : ''}{factor.impact}
                </span>
              </div>
              
              <!-- Impact Bar -->
              <div class="relative h-2 bg-cyber-text-tertiary/20 mb-2">
                <div 
                  class="absolute left-1/2 h-full {
                    factor.impact > 0 ? 'bg-cyber-error' : 'bg-cyber-success'
                  }"
                  style="width: {Math.abs(factor.impact)}%; {
                    factor.impact > 0 
                      ? 'transform: translateX(0)' 
                      : 'transform: translateX(-100%)'
                  }"
                ></div>
              </div>
              
              <div class="terminal-text text-cyber-text-tertiary text-xs">
                {factor.description}
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
    
    <!-- Recommendations -->
    <div class="border-t border-cyber-border-mint/20 pt-4 mt-4">
      <div class="terminal-text text-cyber-text-secondary text-sm mb-2">
        AI_RECOMMENDATION:
      </div>
      
      <div class="terminal-text text-cyber-text-tertiary text-xs">
        {#if riskScore < 25}
          > PROCEED: Transaction appears safe
        {:else if riskScore < 50}
          > CAUTION: Review transaction details
        {:else if riskScore < 75}
          > WARNING: High risk detected, additional verification recommended
        {:else}
          > CRITICAL: Transaction blocked, manual review required
        {/if}
      </div>
    </div>
  {/if}
</CyberCard>