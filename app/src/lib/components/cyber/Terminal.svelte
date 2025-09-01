<script lang="ts">
  export let prompt = '> ';
  export let lines: string[] = [];
  export let maxLines = 20;
  export let typewriter = false;
  export let allowInput = true;
  export let theme: 'mint' | 'lavender' | 'amber' = 'mint';

  let currentInput = '';
  let inputElement: HTMLInputElement;
  let terminalElement: HTMLElement;

  $: themeColors = {
    mint: {
      text: 'text-cyber-mint',
      border: 'border-cyber-border-mint',
      bg: 'bg-cyber-surface-1'
    },
    lavender: {
      text: 'text-cyber-lavender',
      border: 'border-cyber-border-lavender', 
      bg: 'bg-cyber-surface-1'
    },
    amber: {
      text: 'text-cyber-warning',
      border: 'border-cyber-warning',
      bg: 'bg-cyber-surface-1'
    }
  };

  function handleCommand(command: string) {
    const newLines = [...lines, `${prompt}${command}`];
    
    // Process command (basic examples)
    if (command.toLowerCase() === 'help') {
      newLines.push('Available commands: help, clear, status, version');
    } else if (command.toLowerCase() === 'clear') {
      lines = [];
      currentInput = '';
      return;
    } else if (command.toLowerCase() === 'status') {
      newLines.push('System status: ONLINE');
      newLines.push('Connected networks: 9');
      newLines.push('Active validators: 247');
    } else if (command.toLowerCase() === 'version') {
      newLines.push('CrossPay Protocol v2.1.0');
    } else if (command.trim() === '') {
      // Empty command, just add prompt
    } else {
      newLines.push(`Command not found: ${command}`);
    }
    
    // Limit lines
    lines = newLines.slice(-maxLines);
    currentInput = '';
    
    // Auto-scroll to bottom
    setTimeout(() => {
      if (terminalElement) {
        terminalElement.scrollTop = terminalElement.scrollHeight;
      }
    }, 50);
    
    // Focus input
    if (inputElement) {
      inputElement.focus();
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      handleCommand(currentInput);
    }
  }

  // Auto-focus on mount
  import { onMount } from 'svelte';
  onMount(() => {
    if (inputElement && allowInput) {
      inputElement.focus();
    }
  });
</script>

<div class="cyber-terminal {themeColors[theme].bg} {themeColors[theme].border} border font-mono text-sm relative overflow-hidden">
  <!-- Terminal header -->
  <div class="flex items-center justify-between p-3 border-b {themeColors[theme].border} bg-cyber-bg-secondary">
    <div class="flex items-center gap-2">
      <div class="w-3 h-3 rounded-full bg-cyber-error"></div>
      <div class="w-3 h-3 rounded-full bg-cyber-warning"></div>
      <div class="w-3 h-3 rounded-full bg-cyber-success"></div>
    </div>
    
    <div class="text-cyber-text-tertiary text-xs">
      TERMINAL_SESSION_ACTIVE
    </div>
    
    <div class="text-cyber-text-tertiary text-xs">
      {new Date().toLocaleTimeString()}
    </div>
  </div>

  <!-- Terminal content -->
  <div 
    bind:this={terminalElement}
    class="p-4 h-80 overflow-y-auto scrollbar-thin scrollbar-track-transparent scrollbar-thumb-cyber-border-mint"
  >
    <!-- Command history -->
    {#each lines as line, i}
      <div class="text-cyber-text-secondary mb-1 whitespace-pre-wrap break-all">
        {#if typewriter && i === lines.length - 1}
          <span class="animate-typewriter">{line}</span>
        {:else}
          {line}
        {/if}
      </div>
    {/each}

    <!-- Current input line -->
    {#if allowInput}
      <div class="flex items-center gap-1 {themeColors[theme].text}">
        <span>{prompt}</span>
        <input
          bind:this={inputElement}
          bind:value={currentInput}
          on:keydown={handleKeydown}
          class="flex-1 bg-transparent outline-none text-cyber-text-primary"
          placeholder="Type command..."
          autocomplete="off"
          spellcheck="false"
        />
        <span class="animate-cursor-blink">|</span>
      </div>
    {/if}
  </div>

  <!-- Scan lines effect -->
  <div class="absolute inset-0 bg-scan-lines opacity-30 pointer-events-none" />
  
  <!-- Terminal corners -->
  <div class="absolute top-0 left-0 w-2 h-2 border-t border-l {themeColors[theme].border}" />
  <div class="absolute top-0 right-0 w-2 h-2 border-t border-r {themeColors[theme].border}" />
  <div class="absolute bottom-0 left-0 w-2 h-2 border-b border-l {themeColors[theme].border}" />
  <div class="absolute bottom-0 right-0 w-2 h-2 border-b border-r {themeColors[theme].border}" />
</div>

<style>
  .cyber-terminal {
    min-height: 300px;
  }
  
  .cyber-terminal input::placeholder {
    color: rgba(112, 112, 112, 0.6);
  }
  
  /* Custom scrollbar */
  .scrollbar-thin {
    scrollbar-width: thin;
  }
  
  .scrollbar-thin::-webkit-scrollbar {
    width: 4px;
  }
  
  .scrollbar-thin::-webkit-scrollbar-track {
    background: transparent;
  }
  
  .scrollbar-thin::-webkit-scrollbar-thumb {
    background-color: rgba(159, 239, 223, 0.3);
    border-radius: 0;
  }
  
  .scrollbar-thin::-webkit-scrollbar-thumb:hover {
    background-color: rgba(159, 239, 223, 0.5);
  }
</style>