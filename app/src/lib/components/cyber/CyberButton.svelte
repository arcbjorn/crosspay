<script lang="ts">
  export let variant: 'primary' | 'secondary' | 'danger' = 'primary';
  export let size: 'sm' | 'md' | 'lg' = 'md';
  export let disabled = false;
  export let loading = false;
  export let href: string | undefined = undefined;
  export let glitch = true;
  
  $: sizeClass = size === 'sm' ? 'text-sm px-3 py-1' :
                   size === 'lg' ? 'text-lg px-6 py-3' :
                   'text-base px-4 py-2';
  
  $: buttonClasses = [
    'cyber-btn',
    `cyber-btn-${variant}`,
    sizeClass,
    glitch ? 'glitch-hover' : '',
    disabled ? 'opacity-50 cursor-not-allowed' : '',
    loading ? 'animate-pulse' : ''
  ].filter(Boolean).join(' ');
</script>

{#if href}
  <a 
    {href}
    class={buttonClasses}
    class:pointer-events-none={disabled}
    on:click
    on:mouseover
    on:mouseleave
    {...$$restProps}
  >
    {#if loading}
      <span class="ascii-loader inline-block mr-2">
        {'>'}_{'<'}
      </span>
    {/if}
    <slot />
  </a>
{:else}
  <button
    class={buttonClasses}
    {disabled}
    on:click
    on:mouseover
    on:mouseleave
    {...$$restProps}
  >
    {#if loading}
      <span class="ascii-loader inline-block mr-2">
        {'>'}_{'<'}
      </span>
    {/if}
    <slot />
  </button>
{/if}