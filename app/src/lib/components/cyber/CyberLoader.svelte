<script lang="ts">
  export let type: 'spinner' | 'dots' | 'progress' | 'matrix' = 'spinner';
  export let size: 'sm' | 'md' | 'lg' = 'md';
  export let message = '';
  
  let frame = 0;
  
  const spinnerFrames = ['/', '-', '\\', '|'];
  const matrixFrames = ['█', '▓', '▒', '░', ' ', '░', '▒', '▓'];
  
  setInterval(() => {
    frame = (frame + 1) % (type === 'matrix' ? matrixFrames.length : spinnerFrames.length);
  }, 200);
  
  $: sizeClasses = {
    'sm': 'text-sm',
    'md': 'text-base', 
    'lg': 'text-xl'
  }[size];
</script>

<div class="ascii-loader {sizeClasses} flex flex-col items-center gap-2">
  {#if type === 'spinner'}
    <div class="font-mono animate-pulse">
      [{spinnerFrames[frame]}] PROCESSING...
    </div>
  {:else if type === 'dots'}
    <div class="font-mono animate-pulse">
      LOADING{'.'.repeat((frame % 3) + 1)}
    </div>
  {:else if type === 'progress'}
    <div class="font-mono">
      <div class="mb-1">PROGRESS:</div>
      <div class="border border-cyber-border-mint w-32 h-2 relative">
        <div 
          class="bg-cyber-mint h-full animate-pulse"
          style="width: {((frame % 10) + 1) * 10}%"
        ></div>
      </div>
    </div>
  {:else if type === 'matrix'}
    <div class="font-mono grid grid-cols-8 gap-1">
      {#each Array(16) as _, i}
        <div class="w-2 h-2 animate-pulse">
          {matrixFrames[(frame + i) % matrixFrames.length]}
        </div>
      {/each}
    </div>
  {/if}
  
  {#if message}
    <div class="terminal-text text-cyber-text-secondary text-sm">
      {message}
    </div>
  {/if}
</div>