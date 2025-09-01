<script lang="ts">
  export let cols: number = 12;
  export let gap: 'sm' | 'md' | 'lg' = 'md';
  export let responsive = true;
  export let showGrid = false;
  export let animated = false;

  $: gapClass = {
    sm: 'gap-2',
    md: 'gap-4', 
    lg: 'gap-6'
  }[gap];

  $: gridCols = responsive ? 
    `grid-cols-1 sm:grid-cols-2 md:grid-cols-${Math.min(cols, 4)} lg:grid-cols-${Math.min(cols, 6)} xl:grid-cols-${cols}` :
    `grid-cols-${cols}`;
</script>

<div 
  class="cyber-grid-system grid {gridCols} {gapClass} relative"
  class:show-grid={showGrid}
  class:animate-pulse={animated}
>
  <!-- Grid visualization overlay -->
  {#if showGrid}
    <div class="absolute inset-0 pointer-events-none z-0 opacity-30">
      <!-- Vertical lines -->
      {#each Array(cols + 1) as _, i}
        <div 
          class="absolute top-0 bottom-0 w-px bg-cyber-border-mint"
          style="left: {(i / cols) * 100}%"
        />
      {/each}
      
      <!-- Grid pattern -->
      <div class="absolute inset-0" style="
        background-image: 
          linear-gradient(rgba(159, 239, 223, 0.1) 1px, transparent 1px),
          linear-gradient(90deg, rgba(159, 239, 223, 0.1) 1px, transparent 1px);
        background-size: {100 / cols}% 20px;
      " />
    </div>
  {/if}

  <!-- Grid content -->
  <div class="relative z-10 contents">
    <slot />
  </div>
</div>

<style>
  .cyber-grid-system.show-grid {
    min-height: 200px;
  }
  
  .cyber-grid-system.show-grid::before {
    content: '';
    position: absolute;
    top: -8px;
    left: 0;
    right: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(159, 239, 223, 0.4), transparent);
  }
  
  .cyber-grid-system.show-grid::after {
    content: '';
    position: absolute;
    bottom: -8px;
    left: 0;
    right: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(159, 239, 223, 0.4), transparent);
  }
</style>