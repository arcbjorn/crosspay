<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { walletStore } from '@stores/wallet';
  import { chainStore, getChain } from '@stores/chain';
  import { PaymentService } from '@services/payment';
  import { getTokenInfo } from '@config/tokens';
  import { onMount } from 'svelte';
  import type { Payment } from '$lib/types/contracts';
  import type { Address } from 'viem';
  
  $: paymentId = $page.params.id;
  $: wallet = $walletStore;
  $: chain = $chainStore;
  
  let payment: Payment | null = null;
  let paymentChainId = 4202;
  let txHash = '';
  let blockNumber = 0;
  let gasUsed = 0;
  let gasPrice = '';
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
    receiptCID: 'QmExampleReceiptHash123456789abcdef',
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
      
      // Get token info for proper symbol
      const tokenInfo = getTokenInfo(chain.id, paymentData.token);
      
      // Format payment for display
      payment = {
        ...paymentData,
        token: tokenInfo?.symbol || paymentData.token,
      };
      paymentChainId = chain.id;
      // Set transaction details if available (these may come from extended data)
      txHash = paymentData?.['txHash'] || '';
      blockNumber = paymentData?.['blockNumber'] || 0;
      gasUsed = paymentData?.['gasUsed'] || 0;
      gasPrice = paymentData?.['gasPrice'] || '';
      
    } catch (err) {
      console.error('Failed to load payment:', err);
      error = 'Failed to load payment. Using mock data.';
      // Use mock data as fallback
      payment = {
        id: BigInt(mockPayment.id),
        sender: mockPayment.sender,
        recipient: mockPayment.recipient,
        token: mockPayment.token,
        amount: BigInt(mockPayment.amount.replace('.', '').padEnd(18, '0')),
        fee: BigInt(mockPayment.fee.replace('.', '').padEnd(18, '0')),
        status: 'completed',
        createdAt: BigInt(mockPayment.createdAt),
        completedAt: BigInt(mockPayment.completedAt),
        metadataURI: mockPayment.metadataURI,
        receiptCID: '',
        senderENS: '',
        recipientENS: '',
        oraclePrice: '',
        randomSeed: '',
      };
      paymentChainId = mockPayment.chainId;
      txHash = mockPayment.txHash;
      blockNumber = mockPayment.blockNumber;
      gasUsed = mockPayment.gasUsed;
      gasPrice = mockPayment.gasPrice;
    } finally {
      loading = false;
    }
  };
  
  $: paymentChain = getChain(paymentChainId);
  
  onMount(() => {
    loadPayment();
  });
  
  const formatAddress = (address: string) => {
    return `${address.slice(0, 8)}...${address.slice(-6)}`;
  };

  const formatBigIntAmount = (amount: bigint | string | number, decimals = 18) => {
    if (typeof amount === 'string' || typeof amount === 'number') {
      return amount.toString();
    }
    // Convert from wei to tokens
    const divisor = 10n ** BigInt(decimals);
    const whole = amount / divisor;
    const remainder = amount % divisor;
    if (remainder === 0n) {
      return whole.toString();
    }
    const decimalPart = remainder.toString().padStart(decimals, '0').replace(/0+$/, '');
    return decimalPart ? `${whole}.${decimalPart}` : whole.toString();
  };
  
  const formatDate = (timestamp: number | bigint) => {
    const ts = typeof timestamp === 'bigint' ? Number(timestamp) : timestamp;
    return new Date(ts).toLocaleString('en-US', {
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
            {paymentChain?.name || `Chain ${paymentChainId}`}
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
                <span class="font-mono font-bold">{formatBigIntAmount(currentPayment.amount)} {currentPayment.token}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-sm opacity-70">Protocol Fee</span>
                <span class="font-mono">{formatBigIntAmount(currentPayment.fee)} {currentPayment.token}</span>
              </div>
              <div class="divider my-2"></div>
              <div class="flex justify-between font-bold">
                <span>Total Processed</span>
                <span class="font-mono">{formatBigIntAmount((typeof currentPayment.amount === 'bigint' ? currentPayment.amount : BigInt(0)) + (typeof currentPayment.fee === 'bigint' ? currentPayment.fee : BigInt(0)))} {currentPayment.token}</span>
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
                    href={getExplorerUrl(txHash)} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    class="link link-primary"
                  >
                    {formatAddress(txHash)}
                  </a>
                  <button 
                    class="btn btn-ghost btn-xs" 
                    on:click={() => copyToClipboard(txHash)}
                    title="Copy hash"
                  >
                    📋
                  </button>
                </div>
              </div>
              <div class="flex justify-between">
                <span class="text-sm opacity-70">Block Number</span>
                <span class="font-mono text-sm">{blockNumber?.toLocaleString()}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-sm opacity-70">Gas Used</span>
                <span class="font-mono text-sm">{gasUsed?.toLocaleString()}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-sm opacity-70">Gas Price</span>
                <span class="font-mono text-sm">{(parseInt(gasPrice || '0') / 1e9).toFixed(2)} Gwei</span>
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
                    {Math.round((Number(currentPayment.completedAt) - Number(currentPayment.createdAt)) / 60000)} minutes
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

      <!-- Receipt Storage Section -->
      {#if currentPayment.receiptCID}
        <div class="divider"></div>
        <div>
          <h3 class="text-lg font-semibold mb-3">Permanent Receipt Storage</h3>
          <div class="bg-base-200 rounded-lg p-4">
            <div class="flex items-center justify-between mb-3">
              <div>
                <div class="font-medium">Receipt stored on Filecoin</div>
                <div class="text-sm opacity-70">Permanent, immutable receipt data</div>
              </div>
              <div class="badge badge-success">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                </svg>
                Verified
              </div>
            </div>
            
            <div class="flex items-center gap-2 mb-3">
              <span class="text-sm opacity-70">Content ID:</span>
              <code class="bg-base-300 px-2 py-1 rounded text-xs font-mono">
                {currentPayment.receiptCID}
              </code>
              <button 
                class="btn btn-ghost btn-xs" 
                on:click={() => copyToClipboard(currentPayment.receiptCID)}
                title="Copy CID"
              >
                📋
              </button>
            </div>
            
            <div class="flex gap-2">
              <a 
                href="/storage?cid={currentPayment.receiptCID}" 
                class="btn btn-primary btn-sm"
              >
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3M3 17V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v10a2 2 0 01-2 2H5a2 2 0 01-2-2z"></path>
                </svg>
                Download Receipt
              </a>
              <a 
                href="/verify?cid={currentPayment.receiptCID}" 
                class="btn btn-outline btn-sm"
              >
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.031 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
                </svg>
                Verify Receipt
              </a>
              <a 
                href="https://ipfs.io/ipfs/{currentPayment.receiptCID}" 
                target="_blank" 
                rel="noopener noreferrer"
                class="btn btn-ghost btn-sm"
              >
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"></path>
                </svg>
                View on IPFS
              </a>
            </div>
          </div>
        </div>
      {:else}
        <div class="divider"></div>
        <div>
          <h3 class="text-lg font-semibold mb-3">Receipt Storage</h3>
          <div class="alert alert-warning">
            <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 w-6 h-6" fill="none" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path>
            </svg>
            <div>
              <div class="font-bold">Receipt not yet stored</div>
              <div class="text-sm">The permanent receipt is being generated and will be available shortly.</div>
            </div>
          </div>
        </div>
      {/if}

      <!-- ENS Names Section -->
      {#if currentPayment.senderENS || currentPayment.recipientENS}
        <div class="divider"></div>
        <div>
          <h3 class="text-lg font-semibold mb-3">ENS Names</h3>
          <div class="grid md:grid-cols-2 gap-4">
            {#if currentPayment.senderENS}
              <div class="bg-base-200 rounded-lg p-3">
                <div class="text-sm opacity-70">Sender ENS</div>
                <div class="font-medium">{currentPayment.senderENS}</div>
              </div>
            {/if}
            {#if currentPayment.recipientENS}
              <div class="bg-base-200 rounded-lg p-3">
                <div class="text-sm opacity-70">Recipient ENS</div>
                <div class="font-medium">{currentPayment.recipientENS}</div>
              </div>
            {/if}
          </div>
        </div>
      {/if}

      <!-- Oracle Data Section -->
      {#if currentPayment.oraclePrice && parseFloat(currentPayment.oraclePrice) > 0}
        <div class="divider"></div>
        <div>
          <h3 class="text-lg font-semibold mb-3">Price Oracle Data</h3>
          <div class="bg-base-200 rounded-lg p-3">
            <div class="flex justify-between items-center">
              <div>
                <div class="font-medium">Exchange Rate at Payment</div>
                <div class="text-sm opacity-70">Flare FTSO price feed</div>
              </div>
              <div class="font-mono font-bold">
                ${parseFloat(currentPayment.oraclePrice).toFixed(2)}
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
          href={getExplorerUrl(txHash)} 
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
              <span class="font-mono">{paymentChainId}</span>
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