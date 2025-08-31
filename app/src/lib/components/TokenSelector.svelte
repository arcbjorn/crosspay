<script lang="ts">
  import { onMount } from 'svelte';
  import { walletStore } from '$lib/stores/wallet';
  import { getSupportedTokens, getTokenLogoUrl, isNativeToken } from '$lib/config/tokens';
  import { ERC20Service } from '$lib/services/erc20';
  import type { TokenInfo } from '$lib/config/tokens';\n  import type { ERC20TokenBalance } from '$lib/services/erc20';
  import type { Address } from 'viem';

  export let selectedToken: string = '0x0000000000000000000000000000000000000000';
  export let chainId: number;
  export let disabled: boolean = false;

  $: wallet = $walletStore;

  let tokenBalances: ERC20TokenBalance[] = [];
  let loadingBalances = false;
  let erc20Service: ERC20Service;

  $: if (chainId) {
    erc20Service = new ERC20Service(chainId);
    loadTokenBalances();
  }

  $: supportedTokens = getSupportedTokens(chainId);
  $: selectedTokenInfo = supportedTokens.find(t => t.address === selectedToken);

  onMount(() => {
    loadTokenBalances();
  });

  async function loadTokenBalances() {
    if (!wallet.isConnected || !wallet.address || !erc20Service) {
      return;
    }

    loadingBalances = true;
    try {
      tokenBalances = await erc20Service.getTokenBalances(supportedTokens, wallet.address as Address);
    } catch (error) {
      console.error('Failed to load token balances:', error);
      // Fallback to tokens without balances
      tokenBalances = supportedTokens.map(token => ({
        token,
        balance: 0n,
        formattedBalance: '0',
      }));
    } finally {
      loadingBalances = false;
    }
  }

  // Reactive statement to reload balances when wallet changes
  $: if (wallet.isConnected && wallet.address) {
    loadTokenBalances();
  }

  function getTokenBalance(tokenAddress: string): ERC20TokenBalance | undefined {
    return tokenBalances.find(tb => tb.token.address === tokenAddress);
  }

  function handleTokenChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    selectedToken = target.value;
  }
</script>

<div class="form-control">
  <label class="label" for="token-selector">
    <span class="label-text">Token</span>
    <span class="label-text-alt">
      {#if loadingBalances}
        <span class="loading loading-spinner loading-xs"></span>
        Loading balances...
      {:else if selectedTokenInfo}
        Balance: {getTokenBalance(selectedToken)?.formattedBalance || '0'} {selectedTokenInfo.symbol}
      {/if}
    </span>
  </label>
  
  <select 
    id="token-selector"
    class="select select-bordered w-full"
    bind:value={selectedToken}
    on:change={handleTokenChange}
    {disabled}
  >
    {#each supportedTokens as token}
      {@const tokenBalance = getTokenBalance(token.address)}
      <option value={token.address}>
        {token.symbol} - {token.name}
        {#if tokenBalance}
          (Balance: {tokenBalance.formattedBalance})
        {/if}
      </option>
    {/each}
  </select>

  <!-- Token Details -->
  {#if selectedTokenInfo}
    <div class="mt-2 p-3 bg-base-200 rounded-lg">
      <div class="flex items-center gap-3">
        <img 
          src={getTokenLogoUrl(selectedTokenInfo.symbol)} 
          alt={selectedTokenInfo.symbol}
          class="w-8 h-8 rounded-full"
          loading="lazy"
        />
        <div class="flex-1">
          <div class="font-medium text-sm">{selectedTokenInfo.name}</div>
          <div class="text-xs text-base-content/70">
            {selectedTokenInfo.symbol} • {selectedTokenInfo.decimals} decimals
          </div>
        </div>
        <div class="text-right">
          {#if getTokenBalance(selectedToken)}
            {@const balance = getTokenBalance(selectedToken)}
            <div class="font-mono text-sm">
              {balance?.formattedBalance || '0'}
            </div>
            <div class="text-xs text-base-content/70">
              {selectedTokenInfo.symbol}
            </div>
          {:else}
            <div class="skeleton h-4 w-16"></div>
          {/if}
        </div>
      </div>

      <!-- Native token indicator -->
      {#if selectedTokenInfo.isNative}
        <div class="mt-2">
          <div class="badge badge-primary badge-sm">Native Token</div>
        </div>
      {:else}
        <div class="mt-2 text-xs text-base-content/70">
          Contract: 
          <span class="font-mono">
            {selectedTokenInfo.address.slice(0, 6)}...{selectedTokenInfo.address.slice(-4)}
          </span>
        </div>
      {/if}
    </div>
  {/if}

  <!-- Refresh button -->
  <div class="flex justify-end mt-2">
    <button 
      class="btn btn-ghost btn-xs"
      on:click={loadTokenBalances}
      disabled={loadingBalances || !wallet.isConnected}
    >
      {#if loadingBalances}
        <span class="loading loading-spinner loading-xs"></span>
      {:else}
        ↻
      {/if}
      Refresh Balances
    </button>
  </div>
</div>

<style>
  .select option {
    font-family: inherit;
  }
</style>