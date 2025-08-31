<script lang="ts">
  import { onMount } from 'svelte';
  import { oracleService, type OracleHealthStatus, type RandomNumberRequest, type ProofData } from '$lib/services/oracle';
  import PriceDisplay from './PriceDisplay.svelte';

  let healthStatus: OracleHealthStatus | null = null;
  let randomRequests: RandomNumberRequest[] = [];
  let loadingHealth = true;
  let loadingRandom = false;
  
  // FDC Proof state
  let proofId = '';
  let merkleRoot = '';
  let proofData = '';
  let proofArray: string[] = [];
  let loadingProofSubmit = false;
  let loadingProofVerify = false;
  let proofResults: ProofData[] = [];
  let submitSuccess = false;
  let submitError = '';

  const supportedPairs = oracleService.getSupportedPairs();

  onMount(async () => {
    await loadHealthStatus();
  });

  async function loadHealthStatus() {
    try {
      loadingHealth = true;
      healthStatus = await oracleService.getHealthStatus();
    } catch (error) {
      console.error('Failed to load oracle health:', error);
    } finally {
      loadingHealth = false;
    }
  }

  async function requestRandomNumber() {
    try {
      loadingRandom = true;
      const request = await oracleService.requestRandomNumber();
      
      if (request) {
        randomRequests = [request, ...randomRequests];
        
        // Poll for fulfillment
        const pollForFulfillment = async () => {
          const status = await oracleService.getRandomNumberStatus(request.requestId);
          if (status && status.fulfilled) {
            // Update the request in the list
            randomRequests = randomRequests.map(r => 
              r.requestId === request.requestId ? status : r
            );
          } else if (status) {
            // Continue polling if not fulfilled
            setTimeout(pollForFulfillment, 2000);
          }
        };
        
        setTimeout(pollForFulfillment, 2000);
      }
    } catch (error) {
      console.error('Failed to request random number:', error);
    } finally {
      loadingRandom = false;
    }
  }

  async function submitProof() {
    if (!proofId.trim() || !merkleRoot.trim() || !proofData.trim()) {
      submitError = 'Please fill in all required fields';
      return;
    }

    try {
      loadingProofSubmit = true;
      submitError = '';
      
      // Parse proof array from comma-separated string
      const proofParsed = proofData.split(',').map(p => p.trim()).filter(p => p.length > 0);
      
      const success = await oracleService.submitProof(proofId, merkleRoot, proofParsed, proofData);
      
      if (success) {
        submitSuccess = true;
        // Clear form
        proofId = '';
        merkleRoot = '';
        proofData = '';
        // Auto-verify the submitted proof
        setTimeout(() => verifyProof(proofId), 1000);
      } else {
        submitError = 'Proof submission failed';
      }
    } catch (error) {
      console.error('Failed to submit proof:', error);
      submitError = error instanceof Error ? error.message : 'Proof submission failed';
    } finally {
      loadingProofSubmit = false;
    }
  }

  async function verifyProof(id?: string) {
    const verifyId = id || proofId;
    if (!verifyId.trim()) {
      submitError = 'Please enter a proof ID to verify';
      return;
    }

    try {
      loadingProofVerify = true;
      submitError = '';
      
      const result = await oracleService.verifyProof(verifyId);
      
      if (result) {
        proofResults = [result, ...proofResults.filter(p => p.proofId !== result.proofId)];
        submitSuccess = true;
      } else {
        submitError = 'Proof verification failed or proof not found';
      }
    } catch (error) {
      console.error('Failed to verify proof:', error);
      submitError = error instanceof Error ? error.message : 'Proof verification failed';
    } finally {
      loadingProofVerify = false;
    }
  }
</script>

<div class="space-y-6">
  <!-- Oracle Health Status -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="flex justify-between items-center">
        <h2 class="card-title">Oracle Health Status</h2>
        <button class="btn btn-ghost btn-sm" on:click={loadHealthStatus}>
          {loadingHealth ? 'Checking...' : 'Refresh'}
        </button>
      </div>

      {#if loadingHealth}
        <div class="flex gap-4">
          <div class="skeleton h-4 w-20"></div>
          <div class="skeleton h-4 w-20"></div>
          <div class="skeleton h-4 w-20"></div>
        </div>
      {:else if healthStatus}
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div class="stat">
            <div class="stat-title">Overall</div>
            <div class="stat-value text-sm {healthStatus.healthy ? 'text-success' : 'text-error'}">
              {healthStatus.healthy ? '✓ Healthy' : '✗ Unhealthy'}
            </div>
          </div>
          
          <div class="stat">
            <div class="stat-title">FTSO</div>
            <div class="stat-value text-sm {healthStatus.ftsoHealthy ? 'text-success' : 'text-error'}">
              {healthStatus.ftsoHealthy ? '✓ Active' : '✗ Down'}
            </div>
          </div>
          
          <div class="stat">
            <div class="stat-title">RNG</div>
            <div class="stat-value text-sm {healthStatus.randomHealthy ? 'text-success' : 'text-error'}">
              {healthStatus.randomHealthy ? '✓ Active' : '✗ Down'}
            </div>
          </div>
          
          <div class="stat">
            <div class="stat-title">FDC</div>
            <div class="stat-value text-sm {healthStatus.fdcHealthy ? 'text-success' : 'text-error'}">
              {healthStatus.fdcHealthy ? '✓ Active' : '✗ Down'}
            </div>
          </div>
        </div>

        <div class="text-xs text-base-content/70 mt-4">
          Last health check: {new Date(healthStatus.lastHealthCheck).toLocaleString()}
        </div>
      {:else}
        <div class="text-error">
          Failed to load oracle health status
        </div>
      {/if}
    </div>
  </div>

  <!-- Real-time Price Feeds -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title mb-4">Real-time Price Feeds</h2>
      
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {#each supportedPairs as pair}
          <PriceDisplay symbol={pair} showChange={true} />
        {/each}
      </div>
    </div>
  </div>

  <!-- Random Number Generator -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="flex justify-between items-center mb-4">
        <h2 class="card-title">Secure Random Numbers</h2>
        <button 
          class="btn btn-primary" 
          on:click={requestRandomNumber}
          disabled={loadingRandom}
        >
          {loadingRandom ? 'Requesting...' : 'Request Random Number'}
        </button>
      </div>

      {#if randomRequests.length > 0}
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th>Request ID</th>
                <th>Timestamp</th>
                <th>Status</th>
                <th>Seed</th>
              </tr>
            </thead>
            <tbody>
              {#each randomRequests.slice(0, 10) as request}
                <tr>
                  <td class="font-mono text-xs">
                    {request.requestId.slice(0, 8)}...{request.requestId.slice(-8)}
                  </td>
                  <td class="text-sm">
                    {new Date(request.timestamp).toLocaleTimeString()}
                  </td>
                  <td>
                    {#if request.fulfilled}
                      <div class="badge badge-success badge-sm">Fulfilled</div>
                    {:else}
                      <div class="badge badge-warning badge-sm">Pending</div>
                    {/if}
                  </td>
                  <td class="font-mono text-xs">
                    {#if request.seed}
                      {request.seed.slice(0, 10)}...
                    {:else}
                      <span class="text-base-content/50">-</span>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <div class="text-center text-base-content/70 py-8">
          No random number requests yet. Click "Request Random Number" to get started.
        </div>
      {/if}
    </div>
  </div>

  <!-- External Proof Verification -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title mb-4">External Proof Verification (FDC)</h2>
      
      <div class="alert alert-info mb-4">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>
        <span>FDC (Flare Data Connector) allows verification of external blockchain data through Merkle proofs.</span>
      </div>

      {#if submitSuccess}
        <div class="alert alert-success mb-4">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>Proof operation completed successfully!</span>
        </div>
      {/if}

      {#if submitError}
        <div class="alert alert-error mb-4">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{submitError}</span>
        </div>
      {/if}

      <div class="grid md:grid-cols-2 gap-6">
        <!-- Submit Proof Form -->
        <div>
          <h3 class="text-lg font-semibold mb-3">Submit New Proof</h3>
          
          <div class="form-control w-full mb-4">
            <label class="label">
              <span class="label-text">Proof ID</span>
            </label>
            <input 
              bind:value={proofId}
              type="text" 
              placeholder="Enter unique proof identifier" 
              class="input input-bordered w-full" 
            />
          </div>

          <div class="form-control w-full mb-4">
            <label class="label">
              <span class="label-text">Merkle Root</span>
            </label>
            <input 
              bind:value={merkleRoot}
              type="text" 
              placeholder="0x..." 
              class="input input-bordered w-full font-mono text-sm" 
            />
          </div>

          <div class="form-control w-full mb-4">
            <label class="label">
              <span class="label-text">Proof Data (comma-separated)</span>
            </label>
            <textarea 
              bind:value={proofData}
              placeholder="0x123..., 0x456..., 0x789..." 
              class="textarea textarea-bordered w-full font-mono text-sm" 
              rows="3"
            ></textarea>
          </div>
          
          <button 
            class="btn btn-primary w-full" 
            on:click={submitProof}
            disabled={loadingProofSubmit}
          >
            {loadingProofSubmit ? 'Submitting...' : 'Submit Proof'}
          </button>
        </div>

        <!-- Verify Proof Form -->
        <div>
          <h3 class="text-lg font-semibold mb-3">Verify Existing Proof</h3>
          
          <div class="form-control w-full mb-4">
            <label class="label">
              <span class="label-text">Proof ID to Verify</span>
            </label>
            <input 
              bind:value={proofId}
              type="text" 
              placeholder="Enter proof ID to verify" 
              class="input input-bordered w-full" 
            />
          </div>
          
          <button 
            class="btn btn-secondary w-full" 
            on:click={() => verifyProof()}
            disabled={loadingProofVerify}
          >
            {loadingProofVerify ? 'Verifying...' : 'Verify Proof'}
          </button>
        </div>
      </div>

      <!-- Proof Results -->
      {#if proofResults.length > 0}
        <div class="divider"></div>
        <h3 class="text-lg font-semibold mb-3">Verification Results</h3>
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th>Proof ID</th>
                <th>Merkle Root</th>
                <th>Status</th>
                <th>Timestamp</th>
                <th>Data</th>
              </tr>
            </thead>
            <tbody>
              {#each proofResults as result}
                <tr>
                  <td class="font-mono text-xs">
                    {result.proofId.slice(0, 8)}...
                  </td>
                  <td class="font-mono text-xs">
                    {result.merkleRoot.slice(0, 8)}...{result.merkleRoot.slice(-6)}
                  </td>
                  <td>
                    {#if result.verified}
                      <div class="badge badge-success badge-sm">Verified</div>
                    {:else}
                      <div class="badge badge-error badge-sm">Invalid</div>
                    {/if}
                  </td>
                  <td class="text-sm">
                    {new Date(result.timestamp).toLocaleString()}
                  </td>
                  <td class="font-mono text-xs">
                    {result.data.slice(0, 10)}...
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .text-success {
    color: hsl(var(--su));
  }
  
  .text-error {
    color: hsl(var(--er));
  }
</style>