<script lang="ts">
  import { onMount } from 'svelte';
  import CyberCard from './CyberCard.svelte';
  import CyberButton from './CyberButton.svelte';
  
  export let fallbackTitle = 'SYSTEM_ERROR_DETECTED';
  export let showStack = false;
  export let recoverable = true;
  
  let hasError = false;
  let errorInfo: {
    message: string;
    stack?: string;
    timestamp: number;
    userAgent: string;
    url: string;
  } | null = null;
  
  onMount(() => {
    const handleError = (event: ErrorEvent) => {
      hasError = true;
      errorInfo = {
        message: event.message,
        stack: event.error?.stack,
        timestamp: Date.now(),
        userAgent: navigator.userAgent,
        url: window.location.href
      };
    };
    
    const handleUnhandledRejection = (event: PromiseRejectionEvent) => {
      hasError = true;
      errorInfo = {
        message: event.reason?.message || 'Unhandled promise rejection',
        stack: event.reason?.stack,
        timestamp: Date.now(),
        userAgent: navigator.userAgent,
        url: window.location.href
      };
    };
    
    window.addEventListener('error', handleError);
    window.addEventListener('unhandledrejection', handleUnhandledRejection);
    
    return () => {
      window.removeEventListener('error', handleError);
      window.removeEventListener('unhandledrejection', handleUnhandledRejection);
    };
  });
  
  function retry() {
    hasError = false;
    errorInfo = null;
    window.location.reload();
  }
  
  function reportError() {
    if (!errorInfo) return;
    
    const report = {
      ...errorInfo,
      userAction: 'error_report',
      sessionId: Date.now().toString()
    };
    
    console.log('Error report:', report);
    
    // In a real app, send to error tracking service
    navigator.clipboard?.writeText(JSON.stringify(report, null, 2));
    alert('Error report copied to clipboard');
  }
  
  function generateErrorCode(): string {
    if (!errorInfo) return 'ERR_UNKNOWN';
    
    const hash = errorInfo.message
      .split('')
      .reduce((acc, char) => ((acc << 5) - acc + char.charCodeAt(0)) & 0xffffff, 0);
    
    return `ERR_${hash.toString(16).toUpperCase().padStart(6, '0')}`;
  }
  
  $: errorCode = generateErrorCode();
</script>

{#if hasError && errorInfo}
  <div class="error-boundary">
    <CyberCard variant="danger" padding="lg">
      <!-- ASCII Error Display -->
      <div class="terminal-text text-cyber-error text-center mb-6">
        <pre class="text-xs">
{`
╔═══════════════════════════════════════╗
║  ⚠  CRITICAL_SYSTEM_ERROR_DETECTED ⚠  ║
║                                       ║
║  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  ║
║  ░░██████╗██████╗██████╗ ░░██████╗░░  ║
║  ░░██╔═══╝██╔══██╗██╔══██╗ ██╔══██╗░░  ║
║  ░░██████╗██████╔╝██████╔╝ ██████╔╝░░  ║
║  ░░██╔═══╝██╔══██╗██╔══██╗ ██╔══██╗░░  ║
║  ░░███████╗██║  ██║██║  ██║ ██║  ██║░░  ║
║  ░░╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═╝  ╚═╝░░  ║
║  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  ║
╚═══════════════════════════════════════╝
`}
        </pre>
      </div>
      
      <!-- Error Details -->
      <div class="error-details mb-6">
        <div class="terminal-text text-cyber-error text-lg mb-4">
          [{fallbackTitle.toUpperCase()}]
        </div>
        
        <div class="cyber-card bg-cyber-bg-secondary p-4 mb-4">
          <div class="terminal-text text-cyber-text-secondary text-sm mb-2">
            ERROR_CODE: {errorCode}
          </div>
          <div class="terminal-text text-cyber-text-secondary text-sm mb-2">
            TIMESTAMP: {new Date(errorInfo.timestamp).toISOString()}
          </div>
          <div class="terminal-text text-cyber-text-primary text-sm">
            MESSAGE: {errorInfo.message}
          </div>
        </div>
        
        {#if showStack && errorInfo.stack}
          <details class="cyber-card bg-cyber-bg-secondary p-4 mb-4">
            <summary class="terminal-text text-cyber-text-secondary text-sm cursor-pointer">
              [EXPAND_STACK_TRACE]
            </summary>
            <pre class="terminal-text text-cyber-text-tertiary text-xs mt-2 overflow-x-auto">
              {errorInfo.stack}
            </pre>
          </details>
        {/if}
        
        <!-- System Diagnostics -->
        <div class="cyber-card bg-cyber-bg-secondary p-4">
          <div class="terminal-text text-cyber-text-secondary text-sm mb-2">
            SYSTEM_DIAGNOSTICS:
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-xs">
            <div class="flex justify-between">
              <span class="text-cyber-text-tertiary">URL:</span>
              <span class="text-cyber-text-primary font-mono">{errorInfo.url.split('/').pop()}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-cyber-text-tertiary">BROWSER:</span>
              <span class="text-cyber-text-primary font-mono">
                {errorInfo.userAgent.includes('Chrome') ? 'CHROME' : 
                 errorInfo.userAgent.includes('Firefox') ? 'FIREFOX' :
                 errorInfo.userAgent.includes('Safari') ? 'SAFARI' : 'UNKNOWN'}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-cyber-text-tertiary">VIEWPORT:</span>
              <span class="text-cyber-text-primary font-mono">
                {window.innerWidth}x{window.innerHeight}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-cyber-text-tertiary">MEMORY:</span>
              <span class="text-cyber-text-primary font-mono">
                {(performance as any).memory ? 
                  Math.round((performance as any).memory.usedJSHeapSize / 1024 / 1024) + 'MB' : 
                  'N/A'}
              </span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Recovery Actions -->
      <div class="flex gap-4 flex-col sm:flex-row">
        {#if recoverable}
          <CyberButton variant="primary" on:click={retry} class="flex-1">
            [RETRY_OPERATION]
          </CyberButton>
        {/if}
        
        <CyberButton variant="secondary" on:click={reportError} class="flex-1">
          [COPY_ERROR_REPORT]
        </CyberButton>
        
        <CyberButton variant="danger" on:click={() => window.location.href = '/'} class="flex-1">
          [RETURN_TO_HOME]
        </CyberButton>
      </div>
      
      <!-- Recovery Suggestions -->
      <div class="mt-6 terminal-text text-cyber-text-tertiary text-xs">
        <div class="mb-2">RECOVERY_SUGGESTIONS:</div>
        <div class="space-y-1">
          <div>> Try refreshing the page (⌘R / Ctrl+R)</div>
          <div>> Clear browser cache and cookies</div>
          <div>> Disable browser extensions temporarily</div>
          <div>> Check network connection status</div>
          <div>> Contact support with error code: {errorCode}</div>
        </div>
      </div>
    </CyberCard>
  </div>
{:else}
  <slot />
{/if}

<style>
  .error-boundary {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    background: var(--cyber-bg-primary);
  }
  
  .error-details {
    position: relative;
  }
  
  /* Glitch effect for error state */
  .error-boundary::before,
  .error-boundary::after {
    content: '';
    position: absolute;
    inset: 0;
    background: var(--cyber-bg-primary);
    animation: errorGlitch 0.5s ease-in-out infinite alternate;
    pointer-events: none;
  }
  
  .error-boundary::before {
    animation-delay: 0.1s;
    background: linear-gradient(90deg, transparent 0%, rgba(255, 179, 186, 0.1) 50%, transparent 100%);
  }
  
  .error-boundary::after {
    animation-delay: 0.2s;
    background: linear-gradient(-90deg, transparent 0%, rgba(255, 179, 186, 0.05) 50%, transparent 100%);
  }
  
  @keyframes errorGlitch {
    0% { transform: translateX(0); }
    25% { transform: translateX(-2px); }
    50% { transform: translateX(2px); }
    75% { transform: translateX(-1px); }
    100% { transform: translateX(1px); }
  }
  
  /* Reduce motion for accessibility */
  @media (prefers-reduced-motion: reduce) {
    .error-boundary::before,
    .error-boundary::after {
      animation: none;
    }
  }
</style>