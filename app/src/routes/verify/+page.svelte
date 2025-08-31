<script lang="ts">
  import { onMount } from 'svelte';

  let receiptId = '';
  let verificationResult: any = null;
  let isVerifying = false;
  let error = '';

  async function verifyReceipt() {
    if (!receiptId.trim()) {
      error = 'Please enter a receipt ID';
      return;
    }

    isVerifying = true;
    error = '';
    verificationResult = null;

    try {
      // Mock verification - in real implementation would call verification service
      await new Promise(resolve => setTimeout(resolve, 2000)); // Simulate API call
      
      // Mock successful verification result
      verificationResult = {
        id: receiptId,
        valid: Math.random() > 0.3, // 70% chance of being valid
        paymentId: Math.floor(Math.random() * 1000),
        sender: '0x742d35Cc6634C0532925a3b8D4F742d35Cc6634',
        recipient: '0x8ba1f109551bD432803012645Hac136c30f62043',
        amount: '1.5 ETH',
        timestamp: Date.now() - Math.floor(Math.random() * 86400000),
        signature: '0x1234...abcd',
        contentHash: '0xabcd...1234',
        metadataCID: 'QmX7eZYWX8Y8Y8Y8Y8Y8Y8Y8Y8Y8Y8Y8Y8',
        complianceFields: 'KYC: Verified, AML: Clear',
        verifiedAt: Date.now(),
      };
    } catch (err) {
      error = 'Verification failed. Please try again.';
      console.error('Verification error:', err);
    } finally {
      isVerifying = false;
    }
  }

  function resetVerification() {
    verificationResult = null;
    error = '';
    receiptId = '';
  }
</script>

<svelte:head>
  <title>Receipt Verification - CrossPay</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
  <div class="breadcrumbs text-sm mb-8">
    <ul>
      <li><a href="/">Home</a></li>
      <li>Verify Receipt</li>
    </ul>
  </div>

  <div class="mb-8">
    <h1 class="text-4xl font-bold mb-4">Receipt Verification</h1>
    <p class="text-lg text-base-content/70">
      Verify the authenticity and integrity of payment receipts using cryptographic signatures.
    </p>
  </div>

  <!-- Verification Form -->
  <div class="card bg-base-100 shadow-xl mb-8">
    <div class="card-body">
      <h2 class="card-title mb-4">Verify Receipt</h2>
      
      <div class="form-control">
        <label class="label" for="receiptId">
          <span class="label-text">Receipt ID or Content Hash</span>
        </label>
        <div class="flex gap-4">
          <input
            id="receiptId"
            type="text"
            placeholder="Enter receipt ID, payment ID, or content hash"
            class="input input-bordered flex-1"
            bind:value={receiptId}
            disabled={isVerifying}
          />
          <button 
            class="btn btn-primary" 
            on:click={verifyReceipt}
            disabled={isVerifying || !receiptId.trim()}
          >
            {#if isVerifying}
              <span class="loading loading-spinner loading-sm"></span>
              Verifying...
            {:else}
              Verify
            {/if}
          </button>
        </div>
        <label class="label">
          <span class="label-text-alt">
            You can verify receipts using their unique ID, payment ID, or content hash
          </span>
        </label>
      </div>
    </div>
  </div>

  <!-- Verification Result -->
  {#if verificationResult}
    <div class="card bg-base-100 shadow-xl mb-8">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h2 class="card-title">Verification Result</h2>
          <button class="btn btn-ghost btn-sm" on:click={resetVerification}>
            New Verification
          </button>
        </div>

        <!-- Verification Status -->
        <div class="alert {verificationResult.valid ? 'alert-success' : 'alert-error'} mb-6">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            {#if verificationResult.valid}
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            {:else}
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
            {/if}
          </svg>
          <span class="font-medium">
            {#if verificationResult.valid}
              ✓ Receipt is Valid and Verified
            {:else}
              ✗ Receipt Verification Failed
            {/if}
          </span>
        </div>

        <!-- Receipt Details -->
        {#if verificationResult.valid}
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1">Payment ID</label>
                <div class="font-mono text-sm bg-base-200 p-2 rounded">
                  {verificationResult.paymentId}
                </div>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1">Sender</label>
                <div class="font-mono text-sm bg-base-200 p-2 rounded break-all">
                  {verificationResult.sender}
                </div>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1">Recipient</label>
                <div class="font-mono text-sm bg-base-200 p-2 rounded break-all">
                  {verificationResult.recipient}
                </div>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1">Amount</label>
                <div class="font-mono text-lg font-bold text-primary">
                  {verificationResult.amount}
                </div>
              </div>
            </div>

            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1">Timestamp</label>
                <div class="text-sm bg-base-200 p-2 rounded">
                  {new Date(verificationResult.timestamp).toLocaleString()}
                </div>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1">Content Hash</label>
                <div class="font-mono text-xs bg-base-200 p-2 rounded break-all">
                  {verificationResult.contentHash}
                </div>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1">Metadata CID</label>
                <div class="font-mono text-xs bg-base-200 p-2 rounded break-all">
                  {verificationResult.metadataCID}
                </div>
              </div>
              
              <div>
                <label class="block text-sm font-medium text-base-content/70 mb-1">Compliance</label>
                <div class="text-sm bg-base-200 p-2 rounded">
                  {verificationResult.complianceFields}
                </div>
              </div>
            </div>
          </div>

          <!-- Signature Verification -->
          <div class="mt-6">
            <label class="block text-sm font-medium text-base-content/70 mb-2">Cryptographic Signature</label>
            <div class="bg-base-200 p-4 rounded">
              <div class="font-mono text-xs break-all mb-2">
                {verificationResult.signature}
              </div>
              <div class="flex items-center gap-2 text-sm text-success">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Signature verified successfully
              </div>
            </div>
          </div>

          <!-- Verification Timestamp -->
          <div class="mt-4 text-center">
            <div class="text-xs text-base-content/70">
              Verified on {new Date(verificationResult.verifiedAt).toLocaleString()}
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Error Message -->
  {#if error}
    <div class="alert alert-error mb-8">
      <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>{error}</span>
    </div>
  {/if}

  <!-- How It Works -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title mb-4">How Receipt Verification Works</h2>
      
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="text-center">
          <div class="bg-primary/10 rounded-full w-16 h-16 flex items-center justify-center mx-auto mb-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <h3 class="font-semibold mb-2">1. Cryptographic Hash</h3>
          <p class="text-sm text-base-content/70">
            Each receipt has a unique content hash generated from payment data
          </p>
        </div>
        
        <div class="text-center">
          <div class="bg-secondary/10 rounded-full w-16 h-16 flex items-center justify-center mx-auto mb-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m0 0v6a2 2 0 01-2 2H9a2 2 0 01-2-2V9a2 2 0 012-2m6 0V7a2 2 0 00-2-2H9a2 2 0 00-2 2v2m6 0H9" />
            </svg>
          </div>
          <h3 class="font-semibold mb-2">2. Digital Signature</h3>
          <p class="text-sm text-base-content/70">
            Signed by sender's private key and verified using public key cryptography
          </p>
        </div>
        
        <div class="text-center">
          <div class="bg-accent/10 rounded-full w-16 h-16 flex items-center justify-center mx-auto mb-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0121 12a11.955 11.955 0 01-1.382 5.984M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </div>
          <h3 class="font-semibold mb-2">3. Blockchain Verification</h3>
          <p class="text-sm text-base-content/70">
            Cross-referenced with on-chain payment data for complete authenticity
          </p>
        </div>
      </div>
    </div>
  </div>
</div>