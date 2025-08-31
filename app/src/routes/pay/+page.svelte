<script lang="ts">
  import { walletStore } from '$lib/stores/wallet';
  import { chainStore, SUPPORTED_CHAINS, setChain } from '$lib/stores/chain';
  import { goto } from '$app/navigation';
  import { PaymentService } from '$lib/services/payment';
  import { ensService } from '$lib/services/ens';
  import { ERC20Service } from '$lib/services/erc20';
  import { isNativeToken, getSupportedTokens } from '$lib/config/tokens';
  import PriceDisplay from '$lib/components/PriceDisplay.svelte';
  import TokenSelector from '$lib/components/TokenSelector.svelte';
  import ApprovalFlow from '$lib/components/ApprovalFlow.svelte';
  import { successToast, errorToast, warningToast } from '$lib/stores/toast';
  import type { Address } from 'viem';
  
  $: wallet = $walletStore;
  $: chain = $chainStore;
  
  let recipient = '';
  let amount = '';
  let selectedToken = '0x0000000000000000000000000000000000000000'; // ETH
  let metadataURI = '';
  let selectedChain = chain.id;
  let senderENS = '';
  let recipientENS = '';
  
  let isSubmitting = false;
  let error = '';
  let success = '';
  let resolvedAddress = '';
  let isResolvingENS = false;
  let senderENSFromWallet = '';
  let showApprovalFlow = false;
  let approvalRequired = false;
  let erc20Service: ERC20Service;
  let paymentService: PaymentService;
  
  $: fee = amount ? (parseFloat(amount) * 0.001).toFixed(4) : '0';
  $: total = amount ? (parseFloat(amount) + parseFloat(fee)).toFixed(4) : '0';
  $: selectedChainConfig = SUPPORTED_CHAINS[selectedChain];
  $: nativeSymbol = selectedChainConfig?.nativeCurrency.symbol || 'ETH';
  $: supportedTokens = getSupportedTokens(selectedChain);
  $: selectedTokenInfo = supportedTokens.find(t => t.address === selectedToken);
  $: isNative = isNativeToken(selectedToken as Address);
  
  // Initialize services when chain changes
  $: if (selectedChain) {
    erc20Service = new ERC20Service(selectedChain);
    paymentService = new PaymentService(selectedChain);
  }
  
  const handleSubmit = async () => {
    if (!wallet.isConnected) {
      error = 'Please connect your wallet first';
      errorToast('Please connect your wallet first');
      return;
    }
    
    if (!recipient || !amount) {
      error = 'Please fill in all required fields';
      errorToast('Please fill in all required fields');
      return;
    }
    
    if (parseFloat(amount) <= 0) {
      error = 'Amount must be greater than 0';
      errorToast('Amount must be greater than 0');
      return;
    }
    
    // Address validation - either direct address or resolved ENS
    let finalRecipient = recipient;
    if (recipient.endsWith('.eth')) {
      if (!resolvedAddress) {
        error = 'ENS name could not be resolved. Please check the name or enter a direct address.';
        errorToast('ENS name could not be resolved');
        return;
      }
      finalRecipient = resolvedAddress;
    } else if (!/^0x[a-fA-F0-9]{40}$/.test(recipient)) {
      error = 'Please enter a valid Ethereum address or ENS name';
      errorToast('Please enter a valid Ethereum address or ENS name');
      return;
    }
    
    isSubmitting = true;
    error = '';
    
    try {
      // Update chain if different
      if (selectedChain !== chain.id) {
        setChain(selectedChain);
      }
      
      // Check if approval is needed for ERC20 tokens
      if (!isNative && wallet.address && paymentService && erc20Service) {
        const spenderAddress = paymentService.getContractAddress();
        const parsedAmount = erc20Service.parseTokenAmount(amount, selectedTokenInfo?.decimals || 18);
        const approvalStatus = await erc20Service.getApprovalStatus(
          selectedToken as Address,
          wallet.address as Address,
          spenderAddress,
          parsedAmount
        );
        
        if (approvalStatus.needsApproval) {
          approvalRequired = true;
          showApprovalFlow = true;
          isSubmitting = false;
          return;
        }
      }
      
      // Create payment on blockchain  
      const result = await paymentService.createPayment(
        finalRecipient as Address,
        selectedToken as Address,
        amount,
        metadataURI || '',
        wallet.address as Address,
        senderENS,
        recipientENS || (recipient.endsWith('.eth') ? recipient : '')
      );
      
      success = `Payment created successfully! Payment ID: ${result.paymentId}`;
      successToast(`💸 Payment created! ID: ${result.paymentId}`, 3000);
      
      console.log('Payment created:', {
        hash: result.hash,
        paymentId: result.paymentId.toString(),
        recipient,
        amount,
        token: selectedToken,
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
      errorToast(err instanceof Error ? err.message : 'Failed to create payment');
    } finally {
      isSubmitting = false;
    }
  };
  
  const resolveENS = async () => {
    if (!recipient.endsWith('.eth') || !ensService.isValidENSName(recipient)) {
      resolvedAddress = '';
      return;
    }

    isResolvingENS = true;
    
    try {
      const address = await ensService.resolveName(recipient);
      if (address) {
        resolvedAddress = address;
        console.log('Resolved ENS name:', recipient, '->', address);
      } else {
        resolvedAddress = '';
        console.warn('Could not resolve ENS name:', recipient);
      }
    } catch (error) {
      console.error('ENS resolution failed:', error);
      resolvedAddress = '';
    } finally {
      isResolvingENS = false;
    }
  };
  
  // Reverse ENS lookup for sender
  const lookupSenderENS = async () => {
    if (!wallet.address) return;
    
    try {
      const name = await ensService.lookupAddress(wallet.address as Address);
      if (name && !senderENS) {
        senderENSFromWallet = name;
      }
    } catch (error) {
      console.error('Sender ENS lookup failed:', error);
    }
  };

  // Debounced ENS resolution
  let ensTimeout: NodeJS.Timeout;
  $: if (recipient) {
    clearTimeout(ensTimeout);
    if (recipient.endsWith('.eth')) {
      ensTimeout = setTimeout(resolveENS, 500); // 500ms debounce
    } else {
      resolvedAddress = '';
    }
  }

  // Lookup sender ENS when wallet connects
  $: if (wallet.isConnected && wallet.address) {
    lookupSenderENS();
  }
  
  // Handle approval completion
  function handleApprovalCompleted(event: CustomEvent<{ hash: string }>) {
    showApprovalFlow = false;
    approvalRequired = false;
    successToast('Token approved! You can now proceed with payment.');
    // Automatically retry payment submission
    handleSubmit();
  }
  
  // Handle approval cancellation
  function handleApprovalCancelled() {
    showApprovalFlow = false;
    approvalRequired = false;
    warningToast('Token approval cancelled');
  }
  
  // Handle approval error
  function handleApprovalError(event: CustomEvent<{ error: Error }>) {
    showApprovalFlow = false;
    approvalRequired = false;
    errorToast(`Approval failed: ${event.detail.error.message}`);
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
                {#if isResolvingENS}
                  <span class="loading loading-spinner loading-xs"></span> Resolving ENS name...
                {:else if recipient.endsWith('.eth') && resolvedAddress}
                  ✅ Resolved to {ensService.formatAddress(resolvedAddress)}
                {:else if recipient.endsWith('.eth') && !resolvedAddress}
                  ❌ Could not resolve ENS name
                {:else if recipient && !/^0x[a-fA-F0-9]{40}$/.test(recipient) && !recipient.endsWith('.eth')}
                  Please enter a valid address or ENS name
                {/if}
              </span>
            </label>
          </div>
          
          <!-- Token Selection -->
          <TokenSelector 
            bind:selectedToken={selectedToken}
            chainId={selectedChain}
            disabled={isSubmitting}
          />
          
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
              <span class="label-text-alt">
                {#if selectedTokenInfo}
                  Enter amount in {selectedTokenInfo.symbol}
                {/if}
              </span>
            </label>
          </div>
          
          <!-- ENS Names (Optional) -->
          <div class="form-control">
            <label class="label" for="senderENS">
              <span class="label-text">Your ENS Name (Optional)</span>
            </label>
            <input
              id="senderENS"
              type="text"
              placeholder={senderENSFromWallet || "alice.eth"}
              class="input input-bordered w-full"
              bind:value={senderENS}
            />
            <label class="label">
              <span class="label-text-alt">
                {#if senderENSFromWallet && !senderENS}
                  Found: {senderENSFromWallet} (auto-filled)
                {:else}
                  Display name for sender in receipt
                {/if}
              </span>
            </label>
            {#if senderENSFromWallet && !senderENS}
              <button 
                type="button" 
                class="btn btn-ghost btn-xs mt-1"
                on:click={() => senderENS = senderENSFromWallet}
              >
                Use {senderENSFromWallet}
              </button>
            {/if}
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
          
          <!-- Current Exchange Rates -->
          <div class="card bg-base-200/50 border-none">
            <div class="card-body p-4">
              <h3 class="text-sm font-medium mb-2">Current Exchange Rates</h3>
              <div class="grid grid-cols-2 gap-2">
                <PriceDisplay symbol="ETH/USD" compact={true} showChange={false} />
                <PriceDisplay symbol="BTC/USD" compact={true} showChange={false} />
              </div>
            </div>
          </div>

          <!-- Fee Summary -->
          {#if amount}
            <div class="alert">
              <div class="flex justify-between w-full">
                <span>Payment Amount:</span>
                <span class="font-mono">{amount} {selectedTokenInfo?.symbol || nativeSymbol}</span>
              </div>
              <div class="flex justify-between w-full">
                <span>Protocol Fee (0.1%):</span>
                <span class="font-mono">{fee} {selectedTokenInfo?.symbol || nativeSymbol}</span>
              </div>
              <div class="divider my-2"></div>
              <div class="flex justify-between w-full font-bold">
                <span>Total:</span>
                <span class="font-mono">{total} {selectedTokenInfo?.symbol || nativeSymbol}</span>
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

<!-- Token Approval Flow -->
{#if showApprovalFlow && selectedTokenInfo && wallet.address && paymentService && erc20Service}
  <ApprovalFlow
    chainId={selectedChain}
    tokenAddress={selectedToken as Address}
    spenderAddress={paymentService.getContractAddress()}
    requiredAmount={erc20Service.parseTokenAmount(amount, selectedTokenInfo.decimals)}
    bind:show={showApprovalFlow}
    on:approved={handleApprovalCompleted}
    on:cancelled={handleApprovalCancelled}
    on:error={handleApprovalError}
  />
{/if}
