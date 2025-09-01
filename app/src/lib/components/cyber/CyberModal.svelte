<script lang="ts">
  export let open = false;
  export let title: string | undefined = undefined;
  export let closable = true;
  
  import { createEventDispatcher } from 'svelte';
  import CyberButton from './CyberButton.svelte';
  
  const dispatch = createEventDispatcher();
  
  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget && closable) {
      close();
    }
  }
  
  function close() {
    open = false;
    dispatch('close');
  }
  
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && closable) {
      close();
    }
  }
</script>

{#if open}
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div 
    class="cyber-modal"
    on:click={handleBackdropClick}
    on:keydown={handleKeydown}
    tabindex="-1"
    role="dialog"
    aria-modal="true"
  >
    <div class="cyber-modal-content" on:click|stopPropagation>
      {#if closable}
        <button 
          class="absolute top-4 right-4 text-cyber-text-tertiary hover:text-cyber-mint transition-colors"
          on:click={close}
          aria-label="Close modal"
        >
          <span class="terminal-text text-lg">×</span>
        </button>
      {/if}
      
      {#if title}
        <header class="mb-6">
          <h2 class="terminal-text text-cyber-mint text-xl">
            [{title.toUpperCase()}]
          </h2>
        </header>
      {/if}
      
      <main class="relative z-10">
        <slot />
      </main>
      
      {#if $$slots.footer}
        <footer class="mt-6 flex gap-3 justify-end">
          <slot name="footer" />
        </footer>
      {/if}
    </div>
  </div>
{/if}