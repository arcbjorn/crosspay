<script lang="ts">
  import { walletStore } from '$lib/stores/wallet';
  import { chainStore, getChain } from '$lib/stores/chain';
  import { PaymentService } from '$lib/services/payment';
  import { onMount } from 'svelte';
  import type { Address } from 'viem';
  
  $: wallet = $walletStore;
  $: chain = $chainStore;
  
  let payments: any[] = [];
  let loading = true;
  let error = '';
  
  // Mock payment data - this will be replaced with real data
  const mockPayments = [
    {
      id: 1,
      recipient: '0x742d35Cc6634C0532925a3b8D5c9a7f53b3e1234',
      amount: '0.5',
      token: 'ETH',
      status: 'completed',
      createdAt: Date.now() - 86400000, // 1 day ago
      completedAt: Date.now() - 86000000,
      txHash: '0xabcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab',
      chainId: 4202,
      metadataURI: '',
      fee: '0.0005'
    },
    {
      id: 2,
      recipient: '0x8ba1f109551bD432803012645Hac136c7A9B5678',
      amount: '1.2',
      token: 'ETH',
      status: 'pending',
      createdAt: Date.now() - 3600000, // 1 hour ago
      txHash: '0xefgh5678901234567890abcd5678901234567890abcd5678901234567890ef',
      chainId: 84532,
      metadataURI: 'ipfs://QmExampleHash123456789',
      fee: '0.0012'
    },
    {
      id: 3,
      recipient: '0x123456789abcdef123456789abcdef123456789ab',
      amount: '0.8',
      token: 'USDC',
      status: 'refunded',
      createdAt: Date.now() - 172800000, // 2 days ago
      completedAt: Date.now() - 172000000,
      txHash: '0x9876543210abcdef9876543210abcdef9876543210abcdef9876543210abcd',
      chainId: 4202,
      metadataURI: '',
      fee: '0.0008'
    }
  ];
  
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'badge-success';
      case 'pending': return 'badge-warning';
      case 'refunded': return 'badge-error';
      case 'cancelled': return 'badge-neutral';
      default: return 'badge-neutral';
    }
  };
  
  const formatAddress = (address: string) => {
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  };
  
  const formatDate = (timestamp: number) => {
    return new Date(timestamp).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };
  
  const getExplorerUrl = (txHash: string, chainId: number) => {
    const chainConfig = getChain(chainId);
    if (chainConfig && chainConfig.blockExplorers.length > 0) {
      return `${chainConfig.blockExplorers[0].url}/tx/${txHash}`;
    }
    return '#';
  };
  
  const loadPayments = async () => {
    if (!wallet.isConnected || !wallet.address) {
      loading = false;
      return;
    }
    
    try {
      loading = true;
      error = '';
      
      const paymentService = new PaymentService(chain.id);
      
      // Get both sent and received payments
      const [sentPaymentIds, receivedPaymentIds] = await Promise.all([
        paymentService.getUserPayments(wallet.address as Address, true),
        paymentService.getUserPayments(wallet.address as Address, false),
      ]);
      
      // Fetch full payment details for all payment IDs
      const allPaymentIds = [...new Set([...sentPaymentIds, ...receivedPaymentIds])];
      const paymentPromises = allPaymentIds.map(id => paymentService.getPayment(id));
      const paymentResults = await Promise.all(paymentPromises);
      
      // Format payments for display
      payments = paymentResults.map(payment => ({
        id: Number(payment.id),
        sender: payment.sender,
        recipient: payment.recipient,
        amount: paymentService.formatAmount(payment.amount),
        token: payment.token === '0x0000000000000000000000000000000000000000' ? 'ETH' : 'TOKEN',
        status: payment.status,
        createdAt: Number(payment.createdAt) * 1000, // Convert to milliseconds
        completedAt: payment.completedAt > 0n ? Number(payment.completedAt) * 1000 : null,
        chainId: chain.id,
        metadataURI: payment.metadataURI,
        fee: paymentService.formatAmount(payment.fee),
        txHash: '0x' // This would come from transaction events
      }));
      
    } catch (err) {
      console.error('Failed to load payments:', err);
      error = 'Failed to load payment history. Using mock data for now.';
      payments = mockPayments; // Fallback to mock data
    } finally {
      loading = false;
    }
  };
  
  const handleCompletePayment = async (paymentId: number) => {
    if (!wallet.address) return;
    
    try {
      const paymentService = new PaymentService(chain.id);
      await paymentService.completePayment(BigInt(paymentId), wallet.address as Address);
      await loadPayments(); // Reload payments
    } catch (err) {
      console.error('Failed to complete payment:', err);
    }
  };
  
  const handleRefundPayment = async (paymentId: number) => {
    if (!wallet.address) return;
    
    try {
      const paymentService = new PaymentService(chain.id);
      await paymentService.refundPayment(BigInt(paymentId), wallet.address as Address);
      await loadPayments(); // Reload payments
    } catch (err) {
      console.error('Failed to refund payment:', err);
    }
  };
  
  // Load payments when wallet connects or chain changes
  $: if (wallet.isConnected && wallet.address) {
    loadPayments();
  }
  
  onMount(() => {
    if (wallet.isConnected && wallet.address) {
      loadPayments();
    } else {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Payment Receipts - CrossPay</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
  <div class="breadcrumbs text-sm mb-8">
    <ul>
      <li><a href="/">Home</a></li>
      <li>Payment Receipts</li>
    </ul>
  </div>

  <div class="flex items-center justify-between mb-8">
    <h1 class="text-4xl font-bold">Payment History</h1>
    <a href="/pay" class="btn btn-primary">
      New Payment
    </a>
  </div>
  
  {#if !wallet.isConnected}
    <div class="alert alert-info">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
      </svg>
      <span>Connect your wallet to view your payment history</span>
    </div>
  {:else if loading}
    <div class="flex justify-center items-center py-16">
      <span class="loading loading-spinner loading-lg"></span>
      <span class="ml-4">Loading payments...</span>
    </div>
  {:else if error}
    <div class="alert alert-warning">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
      </svg>
      <span>{error}</span>
    </div>
  {:else if payments.length === 0}
    <div class="text-center py-16">
      <div class="text-6xl mb-4">📄</div>
      <h3 class="text-2xl font-bold mb-2">No payments yet</h3>
      <p class="text-base-content/70 mb-6">Your payment history will appear here once you send your first payment.</p>
      <a href="/pay" class="btn btn-primary">Send Your First Payment</a>
    </div>
  {:else}
    <div class="grid gap-6">
      {#each payments as payment}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <div class="flex items-center gap-3 mb-2">
                  <h3 class="card-title">Payment #{payment.id}</h3>
                  <div class="badge {getStatusColor(payment.status)} badge-lg">
                    {payment.status}
                  </div>
                  <div class="badge badge-outline">
                    {getChain(payment.chainId)?.name || `Chain ${payment.chainId}`}
                  </div>
                </div>
                
                <div class="grid md:grid-cols-2 gap-4 mt-4">
                  <div>
                    <p class="text-sm opacity-70">Recipient</p>
                    <p class="font-mono text-sm">{formatAddress(payment.recipient)}</p>
                  </div>
                  
                  <div>
                    <p class="text-sm opacity-70">Amount</p>
                    <p class="font-bold">{payment.amount} {payment.token}</p>
                  </div>
                  
                  <div>
                    <p class="text-sm opacity-70">Created</p>
                    <p class="text-sm">{formatDate(payment.createdAt)}</p>
                  </div>
                  
                  {#if payment.completedAt}
                    <div>
                      <p class="text-sm opacity-70">Completed</p>
                      <p class="text-sm">{formatDate(payment.completedAt)}</p>
                    </div>
                  {/if}
                  
                  <div>
                    <p class="text-sm opacity-70">Transaction Hash</p>
                    <a 
                      href={getExplorerUrl(payment.txHash, payment.chainId)} 
                      target="_blank" 
                      rel="noopener noreferrer"
                      class="link link-primary font-mono text-sm"
                    >
                      {formatAddress(payment.txHash)}
                    </a>
                  </div>
                  
                  <div>
                    <p class="text-sm opacity-70">Fee Paid</p>
                    <p class="text-sm">{payment.fee} {payment.token}</p>
                  </div>
                </div>
                
                {#if payment.metadataURI}
                  <div class="mt-4">
                    <p class="text-sm opacity-70">Metadata</p>
                    <a 
                      href={payment.metadataURI} 
                      target="_blank" 
                      rel="noopener noreferrer"
                      class="link link-secondary text-sm"
                    >
                      View attached data
                    </a>
                  </div>
                {/if}
              </div>
              
              <div class="flex flex-col gap-2 ml-4">
                {#if payment.status === 'pending'}
                  <button 
                    class="btn btn-success btn-sm"
                    on:click={() => handleCompletePayment(payment.id)}
                  >
                    Complete
                  </button>
                  <button 
                    class="btn btn-warning btn-sm"
                    on:click={() => handleRefundPayment(payment.id)}
                  >
                    Refund
                  </button>
                {/if}
                
                <a 
                  href="/receipt/{payment.id}" 
                  class="btn btn-outline btn-sm"
                >
                  View Receipt
                </a>
              </div>
            </div>
          </div>
        </div>
      {/each}
    </div>
    
    <!-- Pagination placeholder -->
    <div class="flex justify-center mt-8">
      <div class="join">
        <button class="join-item btn btn-disabled">«</button>
        <button class="join-item btn btn-active">1</button>
        <button class="join-item btn">2</button>
        <button class="join-item btn">3</button>
        <button class="join-item btn">»</button>
      </div>
    </div>
  {/if}
</div>
