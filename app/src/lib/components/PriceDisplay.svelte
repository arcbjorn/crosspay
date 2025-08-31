<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { oracleService, type PriceData } from '$lib/services/oracle';

  export let symbol: string;
  export let showChange: boolean = true;
  export let compact: boolean = false;

  let price: PriceData | null = null;
  let loading = true;
  let error = '';
  let unsubscribe: (() => void) | null = null;
  let previousPrice: number = 0;
  let priceDirection: 'up' | 'down' | 'same' = 'same';

  onMount(async () => {
    // Get initial price
    await loadPrice();

    // Subscribe to real-time updates
    unsubscribe = oracleService.subscribeToPriceUpdates(symbol, (newPrice) => {
      if (price) {
        previousPrice = price.price;
        priceDirection = newPrice.price > previousPrice ? 'up' : 
                         newPrice.price < previousPrice ? 'down' : 'same';
      }
      price = newPrice;
      loading = false;
    });
  });

  onDestroy(() => {
    if (unsubscribe) {
      unsubscribe();
    }
  });

  async function loadPrice() {
    try {
      loading = true;
      error = '';
      const priceData = await oracleService.getCurrentPrice(symbol);
      
      if (priceData) {
        price = priceData;
        previousPrice = priceData.price;
      } else {
        error = `Failed to load ${symbol} price`;
      }
    } catch (err) {
      error = `Error loading ${symbol} price`;
      console.error('Price loading error:', err);
    } finally {
      loading = false;
    }
  }

  $: priceChangeClass = priceDirection === 'up' ? 'text-success' : 
                        priceDirection === 'down' ? 'text-error' : '';
  
  $: isStale = price ? oracleService.isPriceStale(price.timestamp) : false;
</script>

{#if compact}
  <div class="flex items-center gap-2">
    {#if loading}
      <div class="skeleton h-4 w-16"></div>
    {:else if error}
      <span class="text-error text-sm">Error</span>
    {:else if price}
      <span class="font-mono text-sm {priceChangeClass}" class:opacity-50={isStale}>
        ${oracleService.formatPrice(price.price, price.decimals || 2)}
      </span>
      {#if isStale}
        <div class="tooltip tooltip-warning" data-tip="Price data is stale">
          <span class="text-warning text-xs">⚠</span>
        </div>
      {/if}
    {/if}
  </div>
{:else}
  <div class="card bg-base-100 shadow-sm border">
    <div class="card-body p-4">
      <div class="flex justify-between items-start">
        <h3 class="card-title text-lg">{symbol}</h3>
        {#if price && !price.valid}
          <div class="badge badge-warning badge-sm">Invalid</div>
        {:else if isStale}
          <div class="badge badge-warning badge-sm">Stale</div>
        {:else}
          <div class="badge badge-success badge-sm">Live</div>
        {/if}
      </div>

      {#if loading}
        <div class="flex flex-col gap-2">
          <div class="skeleton h-8 w-24"></div>
          <div class="skeleton h-4 w-20"></div>
        </div>
      {:else if error}
        <div class="text-error">
          <p class="text-sm">{error}</p>
          <button class="btn btn-ghost btn-xs mt-2" on:click={loadPrice}>
            Retry
          </button>
        </div>
      {:else if price}
        <div class="mt-2">
          <div class="text-2xl font-mono font-bold {priceChangeClass}">
            ${oracleService.formatPrice(price.price, price.decimals || 2)}
          </div>
          
          {#if showChange && previousPrice > 0 && previousPrice !== price.price}
            <div class="text-sm {priceChangeClass} mt-1">
              {#if priceDirection === 'up'}
                ↗ +{oracleService.formatPrice(price.price - previousPrice, price.decimals || 2)}
              {:else if priceDirection === 'down'}
                ↘ {oracleService.formatPrice(price.price - previousPrice, price.decimals || 2)}
              {/if}
              ({((price.price - previousPrice) / previousPrice * 100).toFixed(2)}%)
            </div>
          {/if}

          <div class="text-xs text-base-content/70 mt-2">
            Last updated: {new Date(price.timestamp).toLocaleTimeString()}
          </div>
        </div>
      {/if}

      <div class="card-actions justify-end mt-2">
        <button 
          class="btn btn-ghost btn-xs" 
          on:click={loadPrice}
          disabled={loading}
        >
          {loading ? 'Refreshing...' : 'Refresh'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .text-success {
    color: hsl(var(--su));
  }
  
  .text-error {
    color: hsl(var(--er));
  }
  
  .text-warning {
    color: hsl(var(--wa));
  }
</style>