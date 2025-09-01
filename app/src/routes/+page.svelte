<script lang="ts">
  import '../app.css';
  // import { walletStore } from '$lib/stores/wallet';
  import { onMount } from 'svelte';
  
  let typewriterText = '';
  const fullText = 'CrossPay Protocol v2.1.0';
  
  onMount(() => {
    let i = 0;
    const timer = setInterval(() => {
      typewriterText = fullText.slice(0, i);
      i++;
      if (i > fullText.length) {
        clearInterval(timer);
      }
    }, 100);
    
    return () => clearInterval(timer);
  });
  
  // $: wallet = $walletStore;
  const wallet = { isConnected: false }; // Placeholder
</script>

<!-- Terminal Header -->
<div class="relative min-h-screen flex flex-col">
  <!-- ASCII Art Header -->
  <header class="terminal-text p-8 text-center relative z-10">
    <pre class="text-cyber-mint text-sm md:text-base mb-4">
{`
╔═══════════════════════════════════════════════════════════════╗
║  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  ║
║  ░░ ██████╗██████╗  ██████╗ ███████╗███████╗██████╗  █████╗  ░░  ║
║  ░░██╔════╝██╔══██╗██╔═══██╗██╔════╝██╔════╝██╔══██╗██╔══██╗ ░░  ║
║  ░░██║     ██████╔╝██║   ██║███████╗███████╗██████╔╝███████║ ░░  ║
║  ░░██║     ██╔══██╗██║   ██║╚════██║╚════██║██╔═══╝ ██╔══██║ ░░  ║
║  ░░╚██████╗██║  ██║╚██████╔╝███████║███████║██║     ██║  ██║ ░░  ║
║  ░░ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚══════╝╚═╝     ╚═╝  ╚═╝ ░░  ║
║  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  ║
╚═══════════════════════════════════════════════════════════════╝
`}
    </pre>
    
    <div class="font-mono text-cyber-text-primary mb-2">
      > Initializing {typewriterText}<span class="animate-cursor-blink">|</span>
    </div>
    
    <div class="terminal-text text-cyber-text-secondary text-sm">
      [SYSTEM STATUS: ONLINE] [VALIDATORS: ACTIVE] [PRIVACY: ENABLED]
    </div>
  </header>

  <!-- Main Hero Section -->
  <main class="flex-1 flex items-center justify-center relative z-10 px-4">
    <div class="max-w-4xl text-center">
      <h1 class="font-mono text-2xl md:text-4xl text-cyber-text-primary mb-6 glitch-text" data-text="VERIFIABLE CROSS-CHAIN PAYMENTS">
        VERIFIABLE CROSS-CHAIN PAYMENTS
      </h1>
      
      <p class="text-cyber-text-secondary text-lg mb-8 max-w-2xl mx-auto leading-relaxed">
        Secure payment infrastructure across 9 blockchain networks. 
        Private transactions with selective disclosure. 
        Cryptographic proofs for high-value transfers.
      </p>
      
      {#if wallet.isConnected}
        <div class="flex gap-4 justify-center flex-col sm:flex-row">
          <a href="/pay" class="cyber-btn-primary text-lg px-8 py-3 glitch-hover">
            [INITIATE_PAYMENT]
          </a>
          <a href="/receipts" class="cyber-btn-secondary text-lg px-8 py-3 glitch-hover">
            [VIEW_RECEIPTS]
          </a>
        </div>
      {:else}
        <div class="cyber-card max-w-md mx-auto mb-8">
          <div class="flex items-center gap-3">
            <div class="terminal-text text-cyber-warning">⚠</div>
            <span class="terminal-text">WALLET_CONNECTION_REQUIRED</span>
          </div>
          <div class="terminal-text text-sm text-cyber-text-tertiary mt-2">
            > Execute wallet.connect() to proceed
          </div>
        </div>
        
        <button class="cyber-btn-primary text-lg px-8 py-3 glitch-hover">
          [CONNECT_WALLET]
        </button>
      {/if}
    </div>
  </main>

  <!-- Feature Matrix -->
  <section class="relative z-10 py-16">
    <div class="container mx-auto px-4">
      <div class="terminal-text text-center mb-12">
        <div class="text-cyber-mint text-lg mb-2">SYSTEM_MODULES:</div>
      </div>
      
      <div class="grid md:grid-cols-3 gap-6 max-w-6xl mx-auto">
        <!-- Privacy Module -->
        <div class="cyber-card group hover:border-cyber-mint-hover transition-all duration-300">
          <div class="security-matrix mb-4">
            {#each Array(16) as _, i}
              <div class="matrix-cell {i % 7 === 0 ? 'active' : ''}"></div>
            {/each}
          </div>
          
          <div class="terminal-text text-cyber-mint text-lg mb-2">
            [PRIVACY_SHIELD]
          </div>
          <div class="text-cyber-text-secondary text-sm mb-4">
            MODULE_STATUS: ACTIVE<br>
            ENCRYPTION: ZAMA_FHE<br>
            DISCLOSURE: SELECTIVE
          </div>
          <div class="text-cyber-text-tertiary text-xs">
            Payment amounts encrypted with homomorphic encryption. 
            Role-based decryption for compliance and auditing.
          </div>
        </div>

        <!-- Cross-Chain Module -->
        <div class="cyber-card group hover:border-cyber-lavender-hover transition-all duration-300">
          <div class="data-stream h-16 mb-4 flex items-center justify-center">
            <div class="terminal-text text-cyber-lavender">
              ╔═══╗ ═══════════ ╔═══╗<br>
              ║ A ║ ═══════════ ║ B ║<br>
              ╚═══╝ ═══════════ ╚═══╝
            </div>
          </div>
          
          <div class="terminal-text text-cyber-lavender text-lg mb-2">
            [CROSS_CHAIN]
          </div>
          <div class="text-cyber-text-secondary text-sm mb-4">
            NETWORKS: 9_ACTIVE<br>
            ADAPTERS: DEPLOYED<br>
            INTEROP: UNIFIED
          </div>
          <div class="text-cyber-text-tertiary text-xs">
            Seamless payments across Ethereum, Base, Lisk, Polygon, 
            and 5 additional networks with unified infrastructure.
          </div>
        </div>

        <!-- Security Module -->
        <div class="cyber-card group hover:border-cyber-success transition-all duration-300">
          <div class="flex justify-center mb-4">
            <div class="terminal-text text-cyber-success text-center">
              ┌─────────────────┐<br>
              │ VALIDATOR_POOL  │<br>
              │ ███████████████ │<br>
              │ STATUS: SECURED │<br>
              └─────────────────┘
            </div>
          </div>
          
          <div class="terminal-text text-cyber-success text-lg mb-2">
            [VALIDATOR_NETWORK]
          </div>
          <div class="text-cyber-text-secondary text-sm mb-4">
            CONSENSUS: BFT<br>
            PROOFS: CRYPTOGRAPHIC<br>
            SECURITY: SYMBIOTIC
          </div>
          <div class="text-cyber-text-tertiary text-xs">
            High-value transactions secured by distributed validator 
            network with cryptographic proof aggregation.
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- Terminal Footer -->
  <footer class="terminal-text text-center py-8 text-cyber-text-tertiary text-xs relative z-10">
    <div>CROSSPAY_PROTOCOL © 2024 | SECURITY_LEVEL: MAXIMUM | UPTIME: 99.99%</div>
    <div class="mt-1">VERSION: 2.1.0 | BUILD: STABLE | LAST_DEPLOY: {new Date().toISOString().split('T')[0]}</div>
  </footer>
</div>

<!-- Background Effects -->
<div class="fixed inset-0 pointer-events-none">
  <!-- Matrix Rain Effect (subtle) -->
  <div class="absolute inset-0 opacity-5">
    {#each Array(20) as _, i}
      <div 
        class="absolute w-px bg-cyber-mint animate-matrix-rain" 
        style="left: {Math.random() * 100}%; animation-delay: {Math.random() * 10}s; height: 100vh;"
      ></div>
    {/each}
  </div>
</div>
