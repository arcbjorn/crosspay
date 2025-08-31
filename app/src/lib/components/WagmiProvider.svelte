<script lang="ts">
  import { onMount } from 'svelte';
  import { wagmiConfig } from '$lib/wagmi';
  import { reconnect } from 'wagmi/actions';

  let mounted = false;

  onMount(async () => {
    // Initialize wagmi and attempt to reconnect
    try {
      await reconnect(wagmiConfig);
    } catch (error) {
      console.log('No previous wallet connection found');
    }
    mounted = true;
  });
</script>

{#if mounted}
  <slot />
{:else}
  <div class="flex items-center justify-center min-h-screen">
    <div class="loading loading-spinner loading-lg"></div>
  </div>
{/if}