<script lang="ts">
	import { writable } from 'svelte/store';
	import { createFhevmInstance } from '$lib/fhe';
	import type { FHEInstance } from '$lib/fhe';

	export let paymentId: string;
	export let isPrivate: boolean = false;
	export let userRole: 'sender' | 'recipient' | 'compliance' | 'auditor' | 'viewer' = 'viewer';

	interface DisclosureRequest {
		id: string;
		paymentId: string;
		requester: string;
		reason: string;
		requestTime: number;
		approved: boolean;
		status: 'pending' | 'approved' | 'disclosed';
	}

	let fheInstance: FHEInstance | null = null;
	let decryptedAmount: bigint | null = null;
	let decryptedFee: bigint | null = null;
	let disclosureRequests: DisclosureRequest[] = [];
	let newDisclosureReason = '';
	let loading = false;
	let showDecrypted = false;

	const error = writable<string>('');

	// Mock data for demo
	$: {
		disclosureRequests = [
			{
				id: '1',
				paymentId,
				requester: '0x742d35Cc6634C0532925a3b8D4ba9f4e6ad1B6AF',
				reason: 'Tax compliance audit',
				requestTime: Date.now() - 86400000, // 1 day ago
				approved: true,
				status: 'approved'
			},
			{
				id: '2',
				paymentId,
				requester: '0x8ba1f109551bD432803012645Hac136c4c5688dC',
				reason: 'AML investigation',
				requestTime: Date.now() - 3600000, // 1 hour ago
				approved: false,
				status: 'pending'
			}
		];
	}

	async function requestDisclosure() {
		if (!newDisclosureReason.trim()) {
			error.set('Please provide a reason for disclosure');
			return;
		}

		loading = true;
		error.set('');

		try {
			// Mock contract call
			console.log('Requesting disclosure:', {
				paymentId,
				reason: newDisclosureReason
			});

			// Add to local list for demo
			const newRequest: DisclosureRequest = {
				id: Math.random().toString(),
				paymentId,
				requester: '0x' + '0'.repeat(40), // Current user
				reason: newDisclosureReason,
				requestTime: Date.now(),
				approved: false,
				status: 'pending'
			};

			disclosureRequests = [...disclosureRequests, newRequest];
			newDisclosureReason = '';
		} catch (e) {
			error.set(`Failed to request disclosure: ${e}`);
		} finally {
			loading = false;
		}
	}

	async function approveDisclosure(requestId: string) {
		loading = true;
		error.set('');

		try {
			// Mock contract call
			console.log('Approving disclosure:', { paymentId, requestId });

			// Update local state for demo
			disclosureRequests = disclosureRequests.map((req) =>
				req.id === requestId ? { ...req, approved: true, status: 'approved' as const } : req
			);
		} catch (e) {
			error.set(`Failed to approve disclosure: ${e}`);
		} finally {
			loading = false;
		}
	}

	async function revealPayment() {
		if (!fheInstance) {
			try {
				fheInstance = await createFhevmInstance();
			} catch (e) {
				error.set('Failed to initialize FHE instance');
				return;
			}
		}

		loading = true;
		error.set('');

		try {
			// Mock contract call to get encrypted handles
			const mockEncryptedAmountHandle = '0x' + 'a'.repeat(64);
			const mockEncryptedFeeHandle = '0x' + 'b'.repeat(64);

			// Generate keypair for reencryption
			const { publicKey, privateKey } = fheInstance.generateKeypair();

			// Create EIP712 signature (mock)
			const contractAddress = '0x' + '1'.repeat(40);
			const userAddress = '0x' + '0'.repeat(40);

			const eip712 = fheInstance.createEIP712(publicKey, contractAddress);

			// Mock signature process
			const signature = '0x' + 'c'.repeat(130);

			// Decrypt values
			decryptedAmount = await fheInstance.reencrypt(
				mockEncryptedAmountHandle,
				privateKey,
				publicKey,
				signature,
				contractAddress,
				userAddress
			);

			decryptedFee = await fheInstance.reencrypt(
				mockEncryptedFeeHandle,
				privateKey,
				publicKey,
				signature,
				contractAddress,
				userAddress
			);

			showDecrypted = true;
		} catch (e) {
			error.set(`Failed to reveal payment: ${e}`);
		} finally {
			loading = false;
		}
	}

	function grantPermission(viewer: string) {
		console.log('Granting disclosure permission:', { paymentId, viewer });
		// Mock contract call
	}

	function revokePermission(viewer: string) {
		console.log('Revoking disclosure permission:', { paymentId, viewer });
		// Mock contract call
	}
</script>

<div class="disclosure-manager">
	<h3>Privacy Controls</h3>

	{#if !isPrivate}
		<div class="public-notice">🔓 This payment is public - amounts are visible to all</div>
	{:else}
		<div class="private-notice">🔒 This payment is private - amounts are encrypted</div>

		<!-- Disclosure Requests -->
		<div class="section">
			<h4>Disclosure Requests</h4>

			{#if disclosureRequests.length === 0}
				<p>No disclosure requests</p>
			{:else}
				{#each disclosureRequests as request}
					<div class="request-card" class:approved={request.approved}>
						<div class="request-header">
							<span class="requester"
								>From: {request.requester.slice(0, 8)}...{request.requester.slice(-6)}</span
							>
							<span class="status status-{request.status}">{request.status}</span>
						</div>

						<div class="request-reason">
							<strong>Reason:</strong>
							{request.reason}
						</div>

						<div class="request-time">
							Requested: {new Date(request.requestTime).toLocaleString()}
						</div>

						{#if !request.approved && (userRole === 'sender' || userRole === 'recipient')}
							<button
								class="approve-btn"
								on:click={() => approveDisclosure(request.id)}
								disabled={loading}
							>
								Approve Disclosure
							</button>
						{/if}
					</div>
				{/each}
			{/if}
		</div>

		<!-- Request Disclosure -->
		{#if userRole === 'compliance' || userRole === 'auditor'}
			<div class="section">
				<h4>Request Disclosure</h4>
				<div class="request-form">
					<textarea
						bind:value={newDisclosureReason}
						placeholder="Reason for requesting disclosure (e.g., tax audit, AML investigation)"
						rows="3"
					></textarea>
					<button
						on:click={requestDisclosure}
						disabled={loading || !newDisclosureReason.trim()}
						class="request-btn"
					>
						Request Disclosure
					</button>
				</div>
			</div>
		{/if}

		<!-- Reveal Payment (for authorized viewers) -->
		{#if userRole === 'sender' || userRole === 'recipient' || userRole === 'compliance'}
			<div class="section">
				<h4>Reveal Encrypted Amount</h4>

				{#if !showDecrypted}
					<button on:click={revealPayment} disabled={loading} class="reveal-btn">
						{loading ? 'Decrypting...' : '🔍 Reveal Amount'}
					</button>
				{:else}
					<div class="decrypted-values">
						<div class="value-row">
							<span>Amount:</span>
							<span class="amount">{(Number(decryptedAmount) / 1e18).toFixed(4)} ETH</span>
						</div>
						<div class="value-row">
							<span>Fee:</span>
							<span class="fee">{(Number(decryptedFee) / 1e18).toFixed(6)} ETH</span>
						</div>
					</div>

					<button on:click={() => (showDecrypted = false)} class="hide-btn"> Hide Values </button>
				{/if}
			</div>
		{/if}

		<!-- Permission Management (for payment participants) -->
		{#if userRole === 'sender' || userRole === 'recipient'}
			<div class="section">
				<h4>Manage Permissions</h4>
				<div class="permission-controls">
					<input
						type="text"
						placeholder="0x... (address to grant/revoke access)"
						class="address-input"
					/>
					<div class="permission-buttons">
						<button class="grant-btn">Grant Access</button>
						<button class="revoke-btn">Revoke Access</button>
					</div>
				</div>
			</div>
		{/if}
	{/if}

	{#if $error}
		<div class="error">{$error}</div>
	{/if}
</div>

<style>
	.disclosure-manager {
		max-width: 600px;
		margin: 0 auto;
		padding: 1.5rem;
		border: 1px solid #e0e0e0;
		border-radius: 12px;
		background: white;
	}

	.public-notice {
		padding: 1rem;
		background: #e7f3ff;
		border: 1px solid #b8daff;
		border-radius: 8px;
		color: #0c5aa6;
		margin-bottom: 1.5rem;
	}

	.private-notice {
		padding: 1rem;
		background: #f8f9fa;
		border: 1px solid #dee2e6;
		border-radius: 8px;
		color: #495057;
		margin-bottom: 1.5rem;
	}

	.section {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid #eee;
	}

	.section:last-child {
		border-bottom: none;
		margin-bottom: 0;
	}

	h3,
	h4 {
		margin-bottom: 1rem;
		color: #333;
	}

	.request-card {
		border: 1px solid #ddd;
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 1rem;
		background: #fafafa;
	}

	.request-card.approved {
		border-color: #28a745;
		background: #f8fff9;
	}

	.request-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.requester {
		font-family: monospace;
		font-size: 0.9rem;
	}

	.status {
		padding: 0.25rem 0.5rem;
		border-radius: 12px;
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
	}

	.status-pending {
		background: #fff3cd;
		color: #856404;
	}
	.status-approved {
		background: #d4edda;
		color: #155724;
	}
	.status-disclosed {
		background: #cce5ff;
		color: #004085;
	}

	.request-reason,
	.request-time {
		margin-bottom: 0.5rem;
		font-size: 0.9rem;
	}

	.request-time {
		color: #666;
	}

	.request-form textarea {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid #ddd;
		border-radius: 6px;
		margin-bottom: 0.5rem;
		resize: vertical;
	}

	.decrypted-values {
		background: #f8f9fa;
		border: 1px solid #dee2e6;
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 1rem;
	}

	.value-row {
		display: flex;
		justify-content: space-between;
		margin-bottom: 0.5rem;
	}

	.amount,
	.fee {
		font-weight: 600;
		font-family: monospace;
	}

	.permission-controls {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.address-input {
		padding: 0.5rem;
		border: 1px solid #ddd;
		border-radius: 6px;
		font-family: monospace;
		font-size: 0.9rem;
	}

	.permission-buttons {
		display: flex;
		gap: 0.5rem;
	}

	button {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		font-weight: 500;
		transition: all 0.2s;
	}

	.approve-btn {
		background: #28a745;
		color: white;
	}

	.request-btn {
		background: #17a2b8;
		color: white;
	}

	.reveal-btn {
		background: #6f42c1;
		color: white;
	}

	.hide-btn {
		background: #6c757d;
		color: white;
	}

	.grant-btn {
		background: #28a745;
		color: white;
		flex: 1;
	}

	.revoke-btn {
		background: #dc3545;
		color: white;
		flex: 1;
	}

	button:hover:not(:disabled) {
		opacity: 0.9;
		transform: translateY(-1px);
	}

	button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.error {
		background: #f8d7da;
		border: 1px solid #f5c6cb;
		color: #721c24;
		padding: 0.75rem;
		border-radius: 6px;
		margin-top: 1rem;
	}
</style>
