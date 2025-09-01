<script lang="ts">
	import { onMount } from 'svelte';
	import { writable } from 'svelte/store';
	import ConfidentialPayment from '$lib/components/ConfidentialPayment.svelte';
	import DisclosureManager from '$lib/components/DisclosureManager.svelte';

	let activeTab: 'create' | 'manage' = 'create';
	let userRole: 'sender' | 'recipient' | 'compliance' | 'auditor' | 'viewer' = 'sender';
	let selectedPaymentId = '';
	let userAddress = '';

	const recentPayments = writable<any[]>([]);

	onMount(() => {
		// Mock user address - would get from wallet connection
		userAddress = '0x742d35Cc6634C0532925a3b8D4ba9f4e6ad1B6AF';

		// Load recent payments
		recentPayments.set([
			{
				id: '123456',
				recipient: '0x8ba1f109551bD432803012645Hac136c4c5688dC',
				amount: '2.5',
				isPrivate: true,
				status: 'completed',
				createdAt: Date.now() - 86400000
			},
			{
				id: '789012',
				recipient: '0x1234567890123456789012345678901234567890',
				amount: '0.8',
				isPrivate: false,
				status: 'pending',
				createdAt: Date.now() - 3600000
			}
		]);
	});

	function handlePaymentCreated(paymentId: string) {
		selectedPaymentId = paymentId;
		activeTab = 'manage';

		// Add to recent payments
		recentPayments.update((payments) => [
			{
				id: paymentId,
				recipient: '0x' + '0'.repeat(40),
				amount: 'encrypted',
				isPrivate: true,
				status: 'pending',
				createdAt: Date.now()
			},
			...payments
		]);
	}

	function selectPayment(paymentId: string) {
		selectedPaymentId = paymentId;
		activeTab = 'manage';
	}
</script>

<svelte:head>
	<title>CrossPay Privacy - Confidential Payments</title>
	<meta
		name="description"
		content="Create and manage confidential payments with selective disclosure"
	/>
</svelte:head>

<div class="privacy-page">
	<header class="page-header">
		<h1>🔐 Privacy Dashboard</h1>
		<p>Create confidential payments and manage disclosure permissions</p>
	</header>

	<!-- User Role Selector -->
	<div class="role-selector">
		<label for="role">Your Role:</label>
		<select id="role" bind:value={userRole}>
			<option value="sender">Payment Sender</option>
			<option value="recipient">Payment Recipient</option>
			<option value="compliance">Compliance Officer</option>
			<option value="auditor">External Auditor</option>
			<option value="viewer">Public Viewer</option>
		</select>
	</div>

	<!-- Tab Navigation -->
	<div class="tab-navigation">
		<button
			class="tab"
			class:active={activeTab === 'create'}
			on:click={() => (activeTab = 'create')}
		>
			💰 Create Payment
		</button>
		<button
			class="tab"
			class:active={activeTab === 'manage'}
			on:click={() => (activeTab = 'manage')}
		>
			🔍 Manage Disclosures
		</button>
	</div>

	<!-- Tab Content -->
	{#if activeTab === 'create'}
		<div class="tab-content">
			<ConfidentialPayment onPaymentCreated={handlePaymentCreated} />

			<!-- Recent Payments -->
			<div class="recent-payments">
				<h3>Recent Payments</h3>
				{#each $recentPayments as payment}
					<div
						class="payment-item"
						class:private={payment.isPrivate}
						role="button"
						tabindex="0"
						on:click={() => selectPayment(payment.id)}
						on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && selectPayment(payment.id)}
					>
						<div class="payment-info">
							<div class="payment-id">#{payment.id}</div>
							<div class="payment-details">
								<span>To: {payment.recipient.slice(0, 8)}...{payment.recipient.slice(-6)}</span>
								<span>Amount: {payment.isPrivate ? '🔒 Private' : payment.amount + ' ETH'}</span>
							</div>
						</div>
						<div class="payment-status status-{payment.status}">
							{payment.status}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{:else if activeTab === 'manage'}
		<div class="tab-content">
			{#if selectedPaymentId}
				<div class="selected-payment">
					<h3>Managing Payment #{selectedPaymentId}</h3>
				</div>

				<DisclosureManager paymentId={selectedPaymentId} isPrivate={true} {userRole} />
			{:else}
				<div class="no-selection">
					<p>Select a payment from the list above to manage its privacy settings</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.privacy-page {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem;
	}

	.page-header {
		text-align: center;
		margin-bottom: 3rem;
	}

	.page-header h1 {
		font-size: 2.5rem;
		margin-bottom: 1rem;
		color: #333;
	}

	.page-header p {
		font-size: 1.1rem;
		color: #666;
	}

	.role-selector {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 2rem;
		padding: 1rem;
		background: #f8f9fa;
		border-radius: 8px;
	}

	.role-selector label {
		font-weight: 600;
	}

	.role-selector select {
		padding: 0.5rem;
		border: 1px solid #ddd;
		border-radius: 4px;
		font-size: 1rem;
	}

	.tab-navigation {
		display: flex;
		gap: 1rem;
		margin-bottom: 2rem;
		border-bottom: 2px solid #eee;
	}

	.tab {
		padding: 1rem 2rem;
		border: none;
		background: none;
		border-bottom: 3px solid transparent;
		cursor: pointer;
		font-size: 1.1rem;
		font-weight: 500;
		transition: all 0.2s;
	}

	.tab.active {
		border-bottom-color: #667eea;
		color: #667eea;
		background: #f8f9fa;
	}

	.tab:hover {
		background: #f8f9fa;
	}

	.tab-content {
		min-height: 600px;
	}

	.recent-payments {
		margin-top: 3rem;
		padding: 2rem;
		border: 1px solid #e0e0e0;
		border-radius: 12px;
		background: white;
	}

	.recent-payments h3 {
		margin-bottom: 1.5rem;
		color: #333;
	}

	.payment-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		border: 1px solid #eee;
		border-radius: 8px;
		margin-bottom: 1rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.payment-item:hover {
		border-color: #667eea;
		box-shadow: 0 2px 8px rgba(102, 126, 234, 0.1);
	}

	.payment-item.private {
		border-left: 4px solid #764ba2;
	}

	.payment-info {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.payment-id {
		font-weight: 600;
		color: #333;
	}

	.payment-details {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		font-size: 0.9rem;
		color: #666;
	}

	.payment-status {
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.status-pending {
		background: #fff3cd;
		color: #856404;
	}
	.status-completed {
		background: #d4edda;
		color: #155724;
	}
	.status-failed {
		background: #f8d7da;
		color: #721c24;
	}

	.selected-payment {
		margin-bottom: 2rem;
		padding: 1rem;
		background: #e7f3ff;
		border: 1px solid #b8daff;
		border-radius: 8px;
	}

	.selected-payment h3 {
		margin: 0;
		color: #0c5aa6;
	}

	.no-selection {
		text-align: center;
		padding: 4rem;
		color: #666;
	}

	@media (max-width: 768px) {
		.privacy-page {
			padding: 1rem;
		}

		.tab-navigation {
			flex-direction: column;
		}

		.payment-item {
			flex-direction: column;
			gap: 1rem;
			align-items: flex-start;
		}

		.role-selector {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
