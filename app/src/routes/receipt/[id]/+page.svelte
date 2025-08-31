<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { walletStore } from '$lib/stores/wallet';
  import { chainStore, getChain } from '$lib/stores/chain';
  import { PaymentService } from '$lib/services/payment';
  import { onMount } from 'svelte';
  import type { Payment } from '../../../../packages/types/contracts';
  import type { Address } from 'viem';
  
  $: paymentId = $page.params.id;
  $: wallet = $walletStore;
  $: chain = $chainStore;
  
  let payment: Payment | null = null;
  let loading = true;
  let error = '';
  
  // Mock payment data - fallback if loading fails
  const mockPayment = {
    id: parseInt(paymentId || '1'),
    sender: '0x1234567890123456789012345678901234567890',
    recipient: '0x742d35Cc6634C0532925a3b8D5c9a7f53b3e1234',
    amount: '0.5',
    fee: '0.0005',
    token: 'ETH',
    status: 'completed',
    createdAt: Date.now() - 86400000,
    completedAt: Date.now() - 86000000,
    txHash: '0xabcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab',
    chainId: 4202,
    metadataURI: 'ipfs://QmExampleHash123456789abcdef',
    blockNumber: 1234567,
    gasUsed: 85000,
    gasPrice: '20000000000'
  };
  
  const loadPayment = async () => {
    if (!paymentId) {
      error = 'Invalid payment ID';
      loading = false;
      return;
    }
    
    try {
      loading = true;
      error = '';
      
      const paymentService = new PaymentService(chain.id);
      const paymentData = await paymentService.getPayment(BigInt(paymentId));
      
      // Format payment for display
      payment = {
        ...paymentData,
        amount: paymentService.formatAmount(paymentData.amount),
        fee: paymentService.formatAmount(paymentData.fee),
        token: paymentData.token === '0x0000000000000000000000000000000000000000' ? 'ETH' : 'TOKEN',
        createdAt: Number(paymentData.createdAt) * 1000,
        completedAt: paymentData.completedAt > 0n ? Number(paymentData.completedAt) * 1000 : null,
        chainId: chain.id,
      };
      
    } catch (err) {
      console.error('Failed to load payment:', err);
      error = 'Failed to load payment. Using mock data.';
      // Use mock data as fallback
      payment = {
        ...mockPayment,
        amount: mockPayment.amount,
        fee: mockPayment.fee,
        token: mockPayment.token,
        createdAt: mockPayment.createdAt,
        completedAt: mockPayment.completedAt,
        chainId: mockPayment.chainId,
      } as any;
    } finally {
      loading = false;
    }
  };
  
  $: paymentChain = payment ? getChain(payment.chainId) : getChain(mockPayment.chainId);
  
  onMount(() => {
    loadPayment();
  });
  
  const formatAddress = (address: string) => {
    return `${address.slice(0, 8)}...${address.slice(-6)}`;
  };
  
  const formatDate = (timestamp: number) => {
    return new Date(timestamp).toLocaleString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };
  
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'success';
      case 'pending': return 'warning';
      case 'refunded': return 'error';
      case 'cancelled': return 'neutral';
      default: return 'neutral';
    }
  };
  
  const getExplorerUrl = (txHash: string) => {
    if (paymentChain && paymentChain.blockExplorers.length > 0) {
      return `${paymentChain.blockExplorers[0].url}/tx/${txHash}`;
    }
    return '#';
  };
  
  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      // Could add toast notification here
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };
  
  const shareReceipt = async () => {
    const currentPayment = payment || mockPayment;
    const shareData = {
      title: `CrossPay Receipt #${currentPayment.id}`,
      text: `Payment of ${currentPayment.amount} ${currentPayment.token} - View receipt`,
      url: window.location.href
    };
    
    if (navigator.share) {
      try {
        await navigator.share(shareData);
      } catch (err) {
        console.error('Share failed:', err);
      }
    } else {
      // Fallback: copy URL to clipboard
      await copyToClipboard(window.location.href);
    }
  };
  
  $: currentPayment = payment || mockPayment;
</script>

<svelte:head>
  <title>Receipt #{paymentId} - CrossPay</title>
  <meta name="description" content="Payment receipt for CrossPay transaction #{paymentId}" />
</svelte:head>

<div class="max-w-4xl mx-auto">
  <div class="breadcrumbs text-sm mb-8">
    <ul>
      <li><a href="/">Home</a></li>
      <li><a href="/receipts">Receipts</a></li>
      <li>Receipt #{paymentId}</li>
    </ul>
  </div>

  {#if loading}
    <div class="flex justify-center items-center py-16">
      <span class="loading loading-spinner loading-lg"></span>
      <span class="ml-4">Loading receipt...</span>
    </div>
  {:else if error && !payment}
    <div class="alert alert-error mb-8">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>{error}</span>
    </div>
  {:else}
    {#if error}
      <div class="alert alert-warning mb-4">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
        <span>{error}</span>
      </div>
    {/if}

    <div class="flex items-center justify-between mb-8">
      <h1 class="text-4xl font-bold">Payment Receipt</h1>
    <div class="flex gap-2">
      <button class="btn btn-outline btn-sm" on:click={shareReceipt}>
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z"></path>
        </svg>
        Share
      </button>
      <button class="btn btn-outline btn-sm" on:click={() => window.print()}>
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"></path>
        </svg>
        Print
      </button>
    </div>
  </div>

  <!-- Receipt Card -->
  <div class="card bg-base-100 shadow-xl mb-8">
    <div class="card-body">
      <!-- Header with Status -->
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-4">
          <div class="text-6xl">🧾</div>
          <div>
            <h2 class="text-2xl font-bold">Receipt #{currentPayment.id}</h2>
            <div class="badge badge-{getStatusColor(currentPayment.status)} badge-lg">
              {currentPayment.status.toUpperCase()}
            </div>
          </div>
        </div>
        <div class="text-right">
          <div class="text-sm opacity-70">Network</div>
          <div class="badge badge-outline">
            {paymentChain?.name || `Chain ${currentPayment.chainId}`}
          </div>
        </div>
      </div>

      <!-- Payment Details Grid -->
      <div class="grid md:grid-cols-2 gap-6">
        <!-- Left Column -->
        <div class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold mb-3">Payment Details</h3>
            <div class="space-y-3">
              <div class="flex justify-between">
                <span class="text-sm opacity-70">Amount</span>
                <span class="font-mono font-bold">{currentPayment.amount} {currentPayment.token}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-sm opacity-70">Protocol Fee</span>
                <span class="font-mono">{currentPayment.fee} {currentPayment.token}</span>
              </div>
              <div class="divider my-2"></div>
              <div class="flex justify-between font-bold">
                <span>Total Processed</span>
                <span class="font-mono">{(parseFloat(currentPayment.amount) + parseFloat(currentPayment.fee)).toFixed(4)} {currentPayment.token}</span>
              </div>
            </div>
          </div>

          <div>
            <h3 class="text-lg font-semibold mb-3">Participants</h3>
            <div class="space-y-3">
              <div>
                <div class="text-sm opacity-70">From (Sender)</div>
                <div class="font-mono text-sm flex items-center gap-2">
                  {formatAddress(currentPayment.sender)}
                  <button 
                    class="btn btn-ghost btn-xs" 
                    on:click={() => copyToClipboard(currentPayment.sender)}
                    title="Copy address"
                  >
                    📋
                  </button>
                </div>
              </div>
              <div>
                <div class="text-sm opacity-70">To (Recipient)</div>
                <div class="font-mono text-sm flex items-center gap-2">
                  {formatAddress(currentPayment.recipient)}
                  <button 
                    class="btn btn-ghost btn-xs" 
                    on:click={() => copyToClipboard(currentPayment.recipient)}
                    title="Copy address"
                  >
                    📋
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Right Column -->
        <div class="space-y-4">
          <div>
            <h3 class="text-lg font-semibold mb-3">Transaction Info</h3>
            <div class="space-y-3">
              <div>
                <div class="text-sm opacity-70">Transaction Hash</div>
                <div class="font-mono text-sm flex items-center gap-2">
                  <a 
                    href={getExplorerUrl(currentPayment.txHash)} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    class="link link-primary"
                  >
                    {formatAddress(currentPayment.txHash)}
                  </a>
                  <button 
                    class="btn btn-ghost btn-xs" 
                    on:click={() => copyToClipboard(currentPayment.txHash)}
                    title="Copy hash"
                  >
                    📋
                  </button>
                </div>
              </div>
              <div class="flex justify-between">
                <span class="text-sm opacity-70">Block Number</span>
                <span class="font-mono text-sm">{currentPayment.blockNumber?.toLocaleString()}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-sm opacity-70">Gas Used</span>
                <span class="font-mono text-sm">{currentPayment.gasUsed?.toLocaleString()}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-sm opacity-70">Gas Price</span>
                <span class="font-mono text-sm">{(parseInt(currentPayment.gasPrice || '0') / 1e9).toFixed(2)} Gwei</span>
              </div>
            </div>
          </div>

          <div>
            <h3 class="text-lg font-semibold mb-3">Timestamps</h3>
            <div class="space-y-3">
              <div>
                <div class="text-sm opacity-70">Created At</div>
                <div class="text-sm">{formatDate(currentPayment.createdAt)}</div>
              </div>
              {#if currentPayment.completedAt}
                <div>
                  <div class="text-sm opacity-70">Completed At</div>
                  <div class="text-sm">{formatDate(currentPayment.completedAt)}</div>
                </div>
                <div>
                  <div class="text-sm opacity-70">Processing Time</div>
                  <div class="text-sm">
                    {Math.round((currentPayment.completedAt - currentPayment.createdAt) / 60000)} minutes
                  </div>
                </div>
              {/if}
            </div>
          </div>
        </div>
      </div>

      <!-- Metadata Section -->
      {#if currentPayment.metadataURI}
        <div class="divider"></div>
        <div>
          <h3 class="text-lg font-semibold mb-3">Additional Information</h3>
          <div class="alert alert-info">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <div>
              <h4 class="font-bold">Metadata Attached</h4>
              <div class="text-sm">
                <a 
                  href={currentPayment.metadataURI} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  class="link link-primary"
                >
                  View attached data: {currentPayment.metadataURI}
                </a>
              </div>
            </div>
          </div>
        </div>
      {/if}

      <!-- Action Buttons -->
      <div class="divider"></div>
      <div class="flex gap-4 justify-center">
        <a href="/receipts" class="btn btn-outline">
          ← Back to Receipts
        </a>
        <a 
          href={getExplorerUrl(currentPayment.txHash)} 
          target="_blank" 
          rel="noopener noreferrer"
          class="btn btn-primary"
        >
          View on Explorer
        </a>
        {#if currentPayment.status === 'pending' && wallet.isConnected}
          <button class="btn btn-success">
            Complete Payment
          </button>
          <button class="btn btn-warning">
            Request Refund
          </button>
        {/if}
      </div>
    </div>
  </div>

  <!-- Technical Details (Collapsible) -->
  <div class="collapse collapse-arrow bg-base-200">
    <input type="checkbox" />
    <div class="collapse-title text-xl font-medium">
      Technical Details
    </div>
    <div class="collapse-content">
      <div class="grid md:grid-cols-2 gap-4 text-sm">
        <div>
          <h4 class="font-semibold mb-2">Contract Interaction</h4>
          <div class="space-y-1">
            <div class="flex justify-between">
              <span>Contract Address:</span>
              <span class="font-mono">{formatAddress('0x' + '1'.repeat(40))}</span>
            </div>
            <div class="flex justify-between">
              <span>Method Called:</span>
              <span class="font-mono">createPayment</span>
            </div>
            <div class="flex justify-between">
              <span>Input Data:</span>
              <span class="font-mono">0x{paymentId}...</span>
            </div>
          </div>
        </div>
        
        <div>
          <h4 class="font-semibold mb-2">Network Details</h4>
          <div class="space-y-1">
            <div class="flex justify-between">
              <span>Chain ID:</span>
              <span class="font-mono">{currentPayment.chainId}</span>
            </div>
            <div class="flex justify-between">
              <span>Network:</span>
              <span>{paymentChain?.name || 'Unknown'}</span>
            </div>
            <div class="flex justify-between">
              <span>Block Confirmations:</span>
              <span class="font-mono">1,234</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  {/if}
</div>