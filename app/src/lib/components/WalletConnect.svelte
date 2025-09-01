<script lang="ts">
  import { walletStore, connectWallet, disconnectWallet, switchChain } from '@stores/wallet';
  import { chainStore, isChainSupported } from '@stores/chain';
  
  $: wallet = $walletStore;
  $: chain = $chainStore;
  
  let showWalletOptions = false;

  const handleConnect = async (connectorType?: 'metamask' | 'walletconnect') => {
    showWalletOptions = false;
    await connectWallet(connectorType);
  };

  const toggleWalletOptions = () => {
    showWalletOptions = !showWalletOptions;
  };
  
  const handleDisconnect = () => {
    disconnectWallet();
  };
  
  const handleSwitchChain = async () => {
    if (wallet.chainId && !isChainSupported(wallet.chainId)) {
      try {
        await switchChain(chain.id);
      } catch (error) {
        console.error('Failed to switch chain:', error);
      }
    }
  };
  
  $: needsChainSwitch = wallet.isConnected && wallet.chainId && !isChainSupported(wallet.chainId);
  
  const formatAddress = (address: string) => {
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  };
  
  const formatBalance = (balance: bigint) => {
    return (Number(balance) / 10**18).toFixed(4);
  };
</script>

<div class="flex items-center">
  {#if wallet.isConnected}
    <div class="dropdown dropdown-end">
      <div tabindex="0" role="button" class="btn btn-ghost">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 bg-green-500 rounded-full"></div>
          <span class="font-mono text-sm">{formatAddress(wallet.address || '')}</span>
        </div>
      </div>
      <ul role="menu" class="dropdown-content menu p-2 shadow bg-base-100 rounded-box w-80">
        <li class="menu-title">Wallet Info</li>
        <li>
          <div class="flex flex-col items-start">
            <span class="font-mono text-xs opacity-70">{wallet.address}</span>
            <span class="text-sm">
              Balance: {wallet.balance ? formatBalance(wallet.balance) : '0'} ETH
            </span>
            <span class="text-xs opacity-70">
              Chain: {wallet.chainId ? `${wallet.chainId}` : 'Unknown'}
            </span>
          </div>
        </li>
        
        {#if needsChainSwitch}
          <li>
            <button 
              class="btn btn-warning btn-sm" 
              on:click={handleSwitchChain}
            >
              Switch to {chain.name}
            </button>
          </li>
        {/if}
        
        <li>
          <button class="btn btn-outline btn-sm" on:click={handleDisconnect}>
            Disconnect
          </button>
        </li>
      </ul>
    </div>
  {:else}
    <div class="dropdown dropdown-end">
      <div tabindex="0" role="button">
        <button 
          class="btn btn-primary"
          class:loading={wallet.isConnecting}
          disabled={wallet.isConnecting}
          on:click={toggleWalletOptions}
        >
          {wallet.isConnecting ? 'Connecting...' : 'Connect Wallet'}
        </button>
      </div>
      {#if showWalletOptions}
        <ul role="menu" class="dropdown-content menu p-2 shadow bg-base-100 rounded-box w-60">
          <li class="menu-title">Choose Wallet</li>
          <li>
            <button on:click={() => handleConnect('metamask')}>
              <span>🦊</span>
              <span>MetaMask</span>
            </button>
          </li>
          <li>
            <button on:click={() => handleConnect('walletconnect')}>
              <span>🔗</span>
              <span>WalletConnect</span>
            </button>
          </li>
          <li>
            <button on:click={() => handleConnect()}>
              <span>💳</span>
              <span>Browser Wallet</span>
            </button>
          </li>
        </ul>
      {/if}
    </div>
  {/if}
</div>

{#if wallet.error}
  <div class="toast toast-end">
    <div class="alert alert-error">
      <span>{wallet.error}</span>
    </div>
  </div>
{/if}
