<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let type: 'info' | 'success' | 'warning' | 'error' = 'info';
  export let title = '';
  export let message = '';
  export let duration = 5000;
  export let dismissible = true;
  export let showProgress = true;

  const dispatch = createEventDispatcher();

  let visible = true;
  let progressElement: HTMLElement;

  $: typeConfig = {
    info: {
      icon: 'ℹ',
      borderColor: 'border-cyber-border-mint',
      bgColor: 'bg-cyber-surface-1',
      iconColor: 'text-cyber-mint',
      progressColor: 'bg-cyber-mint'
    },
    success: {
      icon: '✓',
      borderColor: 'border-cyber-success',
      bgColor: 'bg-cyber-surface-1',
      iconColor: 'text-cyber-success',
      progressColor: 'bg-cyber-success'
    },
    warning: {
      icon: '⚠',
      borderColor: 'border-cyber-warning',
      bgColor: 'bg-cyber-surface-1',
      iconColor: 'text-cyber-warning',
      progressColor: 'bg-cyber-warning'
    },
    error: {
      icon: '✕',
      borderColor: 'border-cyber-error',
      bgColor: 'bg-cyber-surface-1',
      iconColor: 'text-cyber-error',
      progressColor: 'bg-cyber-error'
    }
  };

  function dismiss() {
    visible = false;
    setTimeout(() => dispatch('dismiss'), 300);
  }

  function startTimer() {
    if (duration > 0) {
      setTimeout(dismiss, duration);
      
      if (showProgress && progressElement) {
        progressElement.style.transition = `width ${duration}ms linear`;
        setTimeout(() => {
          if (progressElement) {
            progressElement.style.width = '0%';
          }
        }, 50);
      }
    }
  }

  $: if (visible) {
    startTimer();
  }
</script>

<div 
  class="cyber-toast font-mono transition-all duration-300 {visible ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'}"
  role="alert"
  aria-live="polite"
>
  <div class="relative {typeConfig[type].bgColor} {typeConfig[type].borderColor} border overflow-hidden">
    <!-- Scan lines effect -->
    <div class="absolute inset-0 bg-scan-lines opacity-30 pointer-events-none" />
    
    <!-- Content -->
    <div class="relative z-10 p-4">
      <div class="flex items-start gap-3">
        <!-- Icon -->
        <div class="{typeConfig[type].iconColor} text-lg font-bold flex-shrink-0">
          {typeConfig[type].icon}
        </div>
        
        <!-- Message content -->
        <div class="flex-1 min-w-0">
          {#if title}
            <div class="text-cyber-text-primary font-medium mb-1 terminal-text">
              > {title}
            </div>
          {/if}
          
          <div class="text-cyber-text-secondary text-sm">
            {message}
          </div>
        </div>
        
        <!-- Dismiss button -->
        {#if dismissible}
          <button
            on:click={dismiss}
            class="flex-shrink-0 text-cyber-text-tertiary hover:text-cyber-text-primary transition-colors duration-200 text-lg leading-none"
            aria-label="Dismiss notification"
          >
            ×
          </button>
        {/if}
      </div>
    </div>
    
    <!-- Progress bar -->
    {#if showProgress && duration > 0}
      <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-cyber-text-tertiary/20">
        <div 
          bind:this={progressElement}
          class="{typeConfig[type].progressColor} h-full w-full origin-left"
          style="transition: none;"
        />
      </div>
    {/if}
    
    <!-- Terminal corners -->
    <div class="absolute top-0 left-0 w-2 h-2 border-t border-l {typeConfig[type].borderColor}" />
    <div class="absolute top-0 right-0 w-2 h-2 border-t border-r {typeConfig[type].borderColor}" />
    <div class="absolute bottom-0 left-0 w-2 h-2 border-b border-l {typeConfig[type].borderColor}" />
    <div class="absolute bottom-0 right-0 w-2 h-2 border-b border-r {typeConfig[type].borderColor}" />
  </div>
</div>

<style>
  .cyber-toast {
    @apply max-w-md w-full;
  }
  
  .cyber-toast .terminal-text {
    @apply font-mono;
  }
</style>