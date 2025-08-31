<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { walletStore } from '$lib/stores/wallet';
  import { ERC20Service } from '$lib/services/erc20';
  import { getTokenInfo, isNativeToken } from '$lib/config/tokens';
  import { successToast, errorToast } from '$lib/stores/toast';
  import type { Address } from 'viem';
  import type { ERC20ApprovalStatus } from '$lib/services/erc20';

  export let chainId: number;
  export let tokenAddress: Address;
  export let spenderAddress: Address;
  export let requiredAmount: bigint;
  export let show: boolean = false;

  const dispatch = createEventDispatcher<{
    approved: { hash: string };
    cancelled: void;
    error: { error: Error };
  }>();

  $: wallet = $walletStore;

  let erc20Service: ERC20Service;
  let approvalStatus: ERC20ApprovalStatus | null = null;
  let isLoading = false;
  let isApproving = false;
  let approvalType: 'exact' | 'max' = 'exact';

  $: if (chainId) {
    erc20Service = new ERC20Service(chainId);
  }

  $: tokenInfo = getTokenInfo(chainId, tokenAddress);
  $: isNative = isNativeToken(tokenAddress);

  // Check approval status when component is shown
  $: if (show && erc20Service && wallet.address && !isNative) {
    checkApprovalStatus();
  }

  async function checkApprovalStatus() {
    if (!erc20Service || !wallet.address || isNative) return;

    isLoading = true;
    try {
      approvalStatus = await erc20Service.getApprovalStatus(
        tokenAddress,
        wallet.address as Address,
        spenderAddress,
        requiredAmount
      );
      
      // If already approved, dispatch success immediately
      if (approvalStatus.hasApproval) {
        dispatch('approved', { hash: '' });
        show = false;
      }
    } catch (error) {
      console.error('Failed to check approval status:', error);
      dispatch('error', { error: error as Error });
    } finally {
      isLoading = false;
    }
  }

  async function handleApprove() {
    if (!erc20Service || !wallet.address || !approvalStatus) return;

    isApproving = true;
    try {
      const amount = approvalType === 'max' ? undefined : requiredAmount;
      
      let hash: string;
      if (approvalType === 'max') {
        hash = await erc20Service.approveMax(tokenAddress, spenderAddress, wallet.address as Address);
        successToast('Maximum approval granted successfully!');
      } else {
        hash = await erc20Service.safeApprove(
          tokenAddress, 
          spenderAddress, 
          requiredAmount, 
          wallet.address as Address
        );
        successToast('Token approval granted successfully!');
      }

      dispatch('approved', { hash });
      show = false;
      
    } catch (error) {
      console.error('Approval failed:', error);
      errorToast(`Approval failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
      dispatch('error', { error: error as Error });
    } finally {
      isApproving = false;
    }
  }

  function handleCancel() {
    dispatch('cancelled');
    show = false;
  }

  function formatAmount(amount: bigint): string {
    if (!tokenInfo) return amount.toString();
    return erc20Service.formatTokenAmount(amount, tokenInfo.decimals, 6);
  }
</script>

{#if show}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-4">Token Approval Required</h3>
      
      {#if isLoading}
        <div class="flex flex-col items-center py-8">
          <div class="loading loading-spinner loading-lg text-primary mb-4"></div>
          <p>Checking approval status...</p>
        </div>
      {:else if isNative}
        <div class="alert alert-info">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>No approval needed for native tokens.</span>
        </div>
      {:else if approvalStatus}
        <div class="space-y-4">
          <!-- Token Information -->
          <div class="bg-base-200 p-4 rounded-lg">
            <div class="flex justify-between items-center mb-2">
              <span class="font-medium">Token:</span>
              <span>{tokenInfo?.symbol || 'Unknown'}</span>
            </div>
            <div class="flex justify-between items-center mb-2">
              <span class="font-medium">Required Amount:</span>
              <span class="font-mono">{formatAmount(requiredAmount)}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="font-medium">Current Allowance:</span>
              <span class="font-mono">{formatAmount(approvalStatus.currentAllowance)}</span>
            </div>
          </div>

          <!-- Approval Explanation -->
          <div class="alert alert-info">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <div>
              <div class="font-medium mb-1">Why do I need to approve?</div>
              <div class="text-sm">
                You need to give permission for the CrossPay contract to spend your {tokenInfo?.symbol} tokens.
                This is a one-time transaction that enables token payments.
              </div>
            </div>
          </div>

          <!-- Approval Type Selection -->
          <div class="form-control">
            <label class="label">
              <span class="label-text font-medium">Approval Amount</span>
            </label>
            <div class="flex gap-4">
              <label class="label cursor-pointer">
                <input 
                  type="radio" 
                  name="approval-type" 
                  class="radio radio-primary" 
                  bind:group={approvalType}
                  value="exact"
                />
                <span class="label-text ml-2">Exact Amount</span>
              </label>
              <label class="label cursor-pointer">
                <input 
                  type="radio" 
                  name="approval-type" 
                  class="radio radio-primary" 
                  bind:group={approvalType}
                  value="max"
                />
                <span class="label-text ml-2">Maximum (Unlimited)</span>
              </label>
            </div>
            <label class="label">
              <span class="label-text-alt">
                {#if approvalType === 'exact'}
                  Approve only the amount needed for this transaction
                {:else}
                  Approve unlimited amount to avoid future approvals (recommended)
                {/if}
              </span>
            </label>
          </div>

          <!-- Gas Estimate -->
          <div class="text-sm text-base-content/70">
            <div class="flex justify-between">
              <span>Estimated gas:</span>
              <span>~50,000 - 100,000 gas</span>
            </div>
          </div>
        </div>
      {/if}

      <!-- Actions -->
      <div class="modal-action">
        <button 
          class="btn btn-ghost" 
          on:click={handleCancel}
          disabled={isApproving}
        >
          Cancel
        </button>
        
        {#if !isNative && approvalStatus && approvalStatus.needsApproval}
          <button 
            class="btn btn-primary" 
            on:click={handleApprove}
            disabled={isApproving || isLoading}
          >
            {#if isApproving}
              <span class="loading loading-spinner loading-sm"></span>
              Approving...
            {:else}
              Approve {tokenInfo?.symbol || 'Token'}
            {/if}
          </button>
        {/if}
      </div>
    </div>
    <div class="modal-backdrop" on:click={handleCancel}></div>
  </div>
{/if}

<style>
  .modal-backdrop {
    background-color: rgba(0, 0, 0, 0.5);
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
  }
</style>