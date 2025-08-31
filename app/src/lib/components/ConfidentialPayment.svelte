<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import { createFhevmInstance, isHighValuePayment, formatEncryptedAmount } from '$lib/fhe';
  import type { FHEInstance } from '$lib/fhe';

  export let recipient: string = '';
  export let token: string = '0x0000000000000000000000000000000000000000'; // ETH
  export let onPaymentCreated: (paymentId: string) => void = () => {};

  let amount = '';
  let metadataURI = '';
  let privateMode = false;
  let loading = false;
  let fheInstance: FHEInstance | null = null;
  let showValidatorWarning = false;
  
  const error = writable<string>('');

  onMount(async () => {
    try {
      fheInstance = await createFhevmInstance();
    } catch (e) {
      error.set('Failed to initialize FHE instance');
    }
  });

  $: {
    const amountBigInt = amount ? BigInt(amount) * 10n**18n : 0n;
    showValidatorWarning = isHighValuePayment(amountBigInt);
  }

  async function createPayment() {
    if (!fheInstance || !amount || !recipient) {
      error.set('Please fill all required fields');
      return;
    }

    loading = true;
    error.set('');

    try {
      const amountWei = BigInt(amount) * 10n**18n;
      
      if (privateMode) {
        // Create encrypted payment using ConfidentialPayments contract
        const userAddress = '0x' + '0'.repeat(40); // Would get from wallet
        const contractAddress = '0x' + '1'.repeat(40); // Would get from config
        
        // Create encrypted input
        const input = fheInstance.createEncryptedInput(contractAddress, userAddress);
        input.add256(amountWei);
        
        const encryptedData = await input.encrypt();
        
        // Call confidential payment contract
        console.log('Creating confidential payment:', {
          recipient,
          token,
          encryptedAmount: encryptedData.handles[0],
          inputProof: encryptedData.inputProof,
          metadataURI,
          makePrivate: true
        });
        
        // Mock contract call for demo
        const paymentId = Math.floor(Math.random() * 1000000).toString();
        onPaymentCreated(paymentId);
        
      } else {
        // Create regular payment using PaymentCore
        console.log('Creating public payment:', {
          recipient,
          token,
          amount: amountWei.toString(),
          metadataURI
        });
        
        // Mock contract call for demo
        const paymentId = Math.floor(Math.random() * 1000000).toString();
        onPaymentCreated(paymentId);
      }
      
      // Reset form
      amount = '';
      recipient = '';
      metadataURI = '';
      privateMode = false;
      
    } catch (e) {
      error.set(`Payment creation failed: ${e}`);
    } finally {
      loading = false;
    }
  }
</script>

<div class="payment-form">
  <h2>Create Payment</h2>
  
  <div class="form-group">
    <label for="recipient">Recipient Address</label>
    <input 
      id="recipient"
      type="text" 
      bind:value={recipient} 
      placeholder="0x..."
      class="address-input"
    />
  </div>

  <div class="form-group">
    <label for="amount">Amount (ETH)</label>
    <input 
      id="amount"
      type="number" 
      bind:value={amount} 
      placeholder="1.0"
      step="0.001"
      min="0"
    />
    
    {#if showValidatorWarning}
      <div class="warning">
        ⚠️ High-value payment detected. This will require validator approval before completion.
      </div>
    {/if}
  </div>

  <div class="form-group">
    <label for="metadata">Metadata URI (optional)</label>
    <input 
      id="metadata"
      type="text" 
      bind:value={metadataURI} 
      placeholder="ipfs://..."
    />
  </div>

  <div class="privacy-toggle">
    <input 
      id="privacy-mode"
      type="checkbox" 
      bind:checked={privateMode}
    />
    <label for="privacy-mode">
      🔒 Private Payment
      {#if privateMode}
        <span class="privacy-info">
          Amount will be encrypted and hidden from public view
        </span>
      {/if}
    </label>
  </div>

  {#if $error}
    <div class="error">{$error}</div>
  {/if}

  <button 
    on:click={createPayment}
    disabled={loading || !amount || !recipient}
    class="create-button"
    class:private={privateMode}
  >
    {#if loading}
      Creating...
    {:else if privateMode}
      🔐 Create Private Payment
    {:else}
      💸 Create Payment
    {/if}
  </button>
</div>

<style>
  .payment-form {
    max-width: 500px;
    margin: 0 auto;
    padding: 2rem;
    border: 1px solid #e0e0e0;
    border-radius: 12px;
    background: white;
    box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 600;
    color: #333;
  }

  input[type="text"], input[type="number"] {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ddd;
    border-radius: 8px;
    font-size: 1rem;
  }

  .address-input {
    font-family: 'Monaco', 'Consolas', monospace;
    font-size: 0.9rem;
  }

  .privacy-toggle {
    margin: 1.5rem 0;
    padding: 1rem;
    background: #f8f9fa;
    border-radius: 8px;
  }

  .privacy-toggle input[type="checkbox"] {
    margin-right: 0.5rem;
    transform: scale(1.2);
  }

  .privacy-toggle label {
    display: inline;
    cursor: pointer;
    margin-bottom: 0;
  }

  .privacy-info {
    display: block;
    font-size: 0.9rem;
    color: #666;
    margin-top: 0.5rem;
    font-weight: normal;
  }

  .warning {
    margin-top: 0.5rem;
    padding: 0.75rem;
    background: #fff3cd;
    border: 1px solid #ffeaa7;
    border-radius: 6px;
    color: #856404;
    font-size: 0.9rem;
  }

  .error {
    background: #f8d7da;
    border: 1px solid #f5c6cb;
    color: #721c24;
    padding: 0.75rem;
    border-radius: 6px;
    margin: 1rem 0;
  }

  .create-button {
    width: 100%;
    padding: 1rem;
    border: none;
    border-radius: 8px;
    font-size: 1.1rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    background: #007bff;
    color: white;
  }

  .create-button:hover:not(:disabled) {
    background: #0056b3;
    transform: translateY(-1px);
  }

  .create-button.private {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  }

  .create-button.private:hover:not(:disabled) {
    background: linear-gradient(135deg, #5a6fd8 0%, #6a4190 100%);
  }

  .create-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
  }

  h2 {
    text-align: center;
    margin-bottom: 2rem;
    color: #333;
  }
</style>