<script lang="ts">
  import { walletStore } from '$lib/stores/wallet';
  
  $: wallet = $walletStore;
</script>

<div class="hero min-h-[70vh] bg-gradient-to-br from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20">
  <div class="hero-content text-center">
    <div class="max-w-md">
      <h1 class="text-5xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
        CrossPay
      </h1>
      <p class="py-6 text-lg">
        Verifiable, private, cross-chain payment infrastructure. Send payments across 9 blockchain networks with unified security and privacy.
      </p>
      
      {#if wallet.isConnected}
        <div class="flex gap-4 justify-center">
          <a href="/pay" class="btn btn-primary btn-lg">
            Send Payment
          </a>
          <a href="/receipts" class="btn btn-outline btn-lg">
            View Receipts
          </a>
        </div>
      {:else}
        <div class="alert alert-info">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>Connect your wallet to start sending payments</span>
        </div>
      {/if}
    </div>
  </div>
</div>

<div class="grid md:grid-cols-3 gap-8 py-16">
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body items-center text-center">
      <div class="text-4xl mb-4">🔒</div>
      <h2 class="card-title">Private & Secure</h2>
      <p>Zama FHE encryption ensures payment amounts remain confidential with selective disclosure.</p>
    </div>
  </div>
  
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body items-center text-center">
      <div class="text-4xl mb-4">🌐</div>
      <h2 class="card-title">Cross-Chain</h2>
      <p>Send payments across Lisk, Base, and 7 other networks with unified infrastructure.</p>
    </div>
  </div>
  
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body items-center text-center">
      <div class="text-4xl mb-4">✅</div>
      <h2 class="card-title">Verifiable</h2>
      <p>Symbiotic validator network provides cryptographic proofs for high-value transactions.</p>
    </div>
  </div>
</div>
