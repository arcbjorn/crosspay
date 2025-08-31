<script lang="ts">
  import { walletStore } from '$lib/stores/wallet';
  import { chainStore, SUPPORTED_CHAINS, setChain } from '$lib/stores/chain';
  import { goto } from '$app/navigation';
  import { PaymentService } from '$lib/services/payment';
  import type { Address } from 'viem';
  
  $: wallet = $walletStore;
  $: chain = $chainStore;
  
  let recipient = '';
  let amount = '';
  let token = '0x0000000000000000000000000000000000000000'; // ETH
  let metadataURI = '';
  let selectedChain = chain.id;
  
  let isSubmitting = false;
  let error = '';
  let success = '';
  
  $: fee = amount ? (parseFloat(amount) * 0.001).toFixed(4) : '0';
  $: total = amount ? (parseFloat(amount) + parseFloat(fee)).toFixed(4) : '0';
  
  const handleSubmit = async () => {
    if (!wallet.isConnected) {
      error = 'Please connect your wallet first';
      return;
    }
    
    if (!recipient || !amount) {
      error = 'Please fill in all required fields';
      return;
    }
    
    if (parseFloat(amount) <= 0) {
      error = 'Amount must be greater than 0';
      return;
    }
    
    // Basic ETH address validation
    if (!/^0x[a-fA-F0-9]{40}$/.test(recipient)) {
      error = 'Please enter a valid Ethereum address';
      return;
    }
    
    isSubmitting = true;
    error = '';
    
    try {
      // Update chain if different
      if (selectedChain !== chain.id) {
        setChain(selectedChain);
      }
      
      // Create payment service for selected chain
      const paymentService = new PaymentService(selectedChain);
      
      // Create payment on blockchain
      const result = await paymentService.createPayment(
        recipient as Address,
        token as Address,
        amount,
        metadataURI || '',
        wallet.address as Address
      );
      
      success = `Payment created successfully! Payment ID: ${result.paymentId}`;
      console.log('Payment created:', {
        hash: result.hash,
        paymentId: result.paymentId.toString(),
        recipient,
        amount,
        token,
        metadataURI,
        chain: selectedChain
      });
      
      // Reset form
      recipient = '';
      amount = '';
      metadataURI = '';
      
      // Redirect to receipt page
      setTimeout(() => {
        goto(`/receipt/${result.paymentId}`);
      }, 2000);
      
    } catch (err) {
      console.error('Payment creation failed:', err);
      error = err instanceof Error ? err.message : 'Failed to create payment';
    } finally {
      isSubmitting = false;
    }
  };
  
  const resolveENS = async () => {
    if (recipient.endsWith('.eth')) {
      // Mock ENS resolution - in real implementation, this would use ENS
      console.log('Resolving ENS name:', recipient);
      // For now, just show that we would resolve it
    }
  };
  
  $: if (recipient.endsWith('.eth')) {
    resolveENS();
  }
</script>

<svelte:head>
  <title>Send Payment - CrossPay</title>
</svelte:head>

<div class="max-w-2xl mx-auto">
  <div class="breadcrumbs text-sm mb-8">
    <ul>
      <li><a href="/">Home</a></li>
      <li>Send Payment</li>
    </ul>
  </div>

  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title text-3xl mb-6">Send Payment</h2>
      
      {#if !wallet.isConnected}
        <div class="alert alert-warning">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
          <span>Please connect your wallet to send payments</span>
        </div>
      {:else}
        <form on:submit|preventDefault={handleSubmit} class="space-y-6">
          <!-- Chain Selection -->
          <div class="form-control">
            <label class="label" for="chain">
              <span class="label-text">Network</span>
            </label>
            <select 
              id="chain"
              class="select select-bordered w-full"
              bind:value={selectedChain}
            >
              {#each Object.values(SUPPORTED_CHAINS) as chainOption}
                <option value={chainOption.id}>
                  {chainOption.name} {chainOption.testnet ? '(Testnet)' : ''}
                </option>
              {/each}
            </select>
          </div>
          
          <!-- Recipient Address -->
          <div class="form-control">
            <label class="label" for="recipient">
              <span class="label-text">Recipient Address or ENS Name</span>
            </label>
            <input
              id="recipient"
              type="text"
              placeholder="0x... or username.eth"
              class="input input-bordered w-full"
              class:input-success={recipient.endsWith('.eth')}
              bind:value={recipient}
              required
            />
            <label class="label">
              <span class="label-text-alt">
                {#if recipient.endsWith('.eth')}
                  ENS name detected - will resolve to address
                {:else if recipient && !/^0x[a-fA-F0-9]{40}$/.test(recipient)}
                  Please enter a valid address
                {/if}
              </span>
            </label>
          </div>
          
          <!-- Token Selection -->
          <div class="form-control">
            <label class="label" for="token">
              <span class="label-text">Token</span>
            </label>
            <select 
              id="token"
              class="select select-bordered w-full"
              bind:value={token}
            >
              <option value="0x0000000000000000000000000000000000000000">ETH (Native)</option>
              <!-- Mock USDC addresses would go here -->
              <option value="0xa0b86991c31cc0c24b383c0d36b8ef073b3b2c0f8">USDC (Mock)</option>
            </select>
          </div>
          
          <!-- Amount -->
          <div class="form-control">
            <label class="label" for="amount">
              <span class="label-text">Amount</span>
            </label>
            <input
              id="amount"
              type="number"
              step="0.001"
              min="0"
              placeholder="0.0"
              class="input input-bordered w-full"
              bind:value={amount}
              required
            />
            <label class="label">
              <span class="label-text-alt">Available: {wallet.balance ? (Number(wallet.balance) / 10**18).toFixed(4) : '0'} ETH</span>
            </label>
          </div>
          
          <!-- Metadata URI -->
          <div class="form-control">
            <label class="label" for="metadata">
              <span class="label-text">Metadata URI (Optional)</span>
            </label>
            <input
              id="metadata"
              type="url"
              placeholder="ipfs://... or https://..."
              class="input input-bordered w-full"
              bind:value={metadataURI}
            />
            <label class="label">
              <span class="label-text-alt">Link to additional payment information</span>
            </label>
          </div>
          
          <!-- Fee Summary -->
          {#if amount}
            <div class="alert">
              <div class="flex justify-between w-full">
                <span>Payment Amount:</span>
                <span class="font-mono">{amount} ETH</span>
              </div>
              <div class="flex justify-between w-full">
                <span>Protocol Fee (0.1%):</span>
                <span class="font-mono">{fee} ETH</span>
              </div>
              <div class="divider my-2"></div>
              <div class="flex justify-between w-full font-bold">
                <span>Total:</span>
                <span class="font-mono">{total} ETH</span>
              </div>
            </div>
          {/if}
          
          <!-- Submit Button -->
          <div class="form-control mt-8">
            <button
              type="submit"
              class="btn btn-primary btn-lg"
              class:loading={isSubmitting}
              disabled={isSubmitting || !wallet.isConnected}
            >
              {isSubmitting ? 'Creating Payment...' : 'Send Payment'}
            </button>
          </div>
        </form>
      {/if}
      
      <!-- Error/Success Messages -->
      {#if error}
        <div class="alert alert-error mt-4">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{error}</span>
        </div>
      {/if}
      
      {#if success}
        <div class="alert alert-success mt-4">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{success}</span>
        </div>
      {/if}
    </div>
  </div>
</div>
