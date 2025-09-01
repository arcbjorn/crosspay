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
			await new Promise((resolve) => setTimeout(resolve, 2000)); // Simulate API call

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
				verifiedAt: Date.now()
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

<div class="mx-auto max-w-4xl">
	<div class="breadcrumbs mb-8 text-sm">
		<ul>
			<li><a href="/">Home</a></li>
			<li>Verify Receipt</li>
		</ul>
	</div>

	<div class="mb-8">
		<h1 class="mb-4 text-4xl font-bold">Receipt Verification</h1>
		<p class="text-base-content/70 text-lg">
			Verify the authenticity and integrity of payment receipts using cryptographic signatures.
		</p>
	</div>

	<!-- Verification Form -->
	<div class="card bg-base-100 mb-8 shadow-xl">
		<div class="card-body">
			<h2 class="card-title mb-4">Verify Receipt</h2>

			<div class="form-control">
				<label class="label" for="receipt-id-input">
					<span class="label-text">Receipt ID or Content Hash</span>
				</label>
				<div class="flex gap-4">
					<input
						id="receipt-id-input"
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
				<div class="label">
					<span class="label-text-alt">
						You can verify receipts using their unique ID, payment ID, or content hash
					</span>
				</div>
			</div>
		</div>
	</div>

	<!-- Verification Result -->
	{#if verificationResult}
		<div class="card bg-base-100 mb-8 shadow-xl">
			<div class="card-body">
				<div class="mb-4 flex items-center justify-between">
					<h2 class="card-title">Verification Result</h2>
					<button class="btn btn-ghost btn-sm" on:click={resetVerification}>
						New Verification
					</button>
				</div>

				<!-- Verification Status -->
				<div class="alert {verificationResult.valid ? 'alert-success' : 'alert-error'} mb-6">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-6 w-6 shrink-0 stroke-current"
						fill="none"
						viewBox="0 0 24 24"
					>
						{#if verificationResult.valid}
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						{:else}
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
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
					<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
						<div class="space-y-4">
							<div>
								<div class="text-base-content/70 mb-1 block text-sm font-medium">Payment ID</div>
								<div class="bg-base-200 rounded p-2 font-mono text-sm">
									{verificationResult.paymentId}
								</div>
							</div>

							<div>
								<div class="text-base-content/70 mb-1 block text-sm font-medium">Sender</div>
								<div class="bg-base-200 break-all rounded p-2 font-mono text-sm">
									{verificationResult.sender}
								</div>
							</div>

							<div>
								<div class="text-base-content/70 mb-1 block text-sm font-medium">Recipient</div>
								<div class="bg-base-200 break-all rounded p-2 font-mono text-sm">
									{verificationResult.recipient}
								</div>
							</div>

							<div>
								<div class="text-base-content/70 mb-1 block text-sm font-medium">Amount</div>
								<div class="text-primary font-mono text-lg font-bold">
									{verificationResult.amount}
								</div>
							</div>
						</div>

						<div class="space-y-4">
							<div>
								<div class="text-base-content/70 mb-1 block text-sm font-medium">Timestamp</div>
								<div class="bg-base-200 rounded p-2 text-sm">
									{new Date(verificationResult.timestamp).toLocaleString()}
								</div>
							</div>

							<div>
								<div class="text-base-content/70 mb-1 block text-sm font-medium">Content Hash</div>
								<div class="bg-base-200 break-all rounded p-2 font-mono text-xs">
									{verificationResult.contentHash}
								</div>
							</div>

							<div>
								<div class="text-base-content/70 mb-1 block text-sm font-medium">Metadata CID</div>
								<div class="bg-base-200 break-all rounded p-2 font-mono text-xs">
									{verificationResult.metadataCID}
								</div>
							</div>

							<div>
								<div class="text-base-content/70 mb-1 block text-sm font-medium">Compliance</div>
								<div class="bg-base-200 rounded p-2 text-sm">
									{verificationResult.complianceFields}
								</div>
							</div>
						</div>
					</div>

					<!-- Signature Verification -->
					<div class="mt-6">
						<div class="text-base-content/70 mb-2 block text-sm font-medium">
							Cryptographic Signature
						</div>
						<div class="bg-base-200 rounded p-4">
							<div class="mb-2 break-all font-mono text-xs">
								{verificationResult.signature}
							</div>
							<div class="text-success flex items-center gap-2 text-sm">
								<svg
									xmlns="http://www.w3.org/2000/svg"
									class="h-4 w-4"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
									/>
								</svg>
								Signature verified successfully
							</div>
						</div>
					</div>

					<!-- Verification Timestamp -->
					<div class="mt-4 text-center">
						<div class="text-base-content/70 text-xs">
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
			<svg
				xmlns="http://www.w3.org/2000/svg"
				class="h-6 w-6 shrink-0 stroke-current"
				fill="none"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
				/>
			</svg>
			<span>{error}</span>
		</div>
	{/if}

	<!-- How It Works -->
	<div class="card bg-base-100 shadow-xl">
		<div class="card-body">
			<h2 class="card-title mb-4">How Receipt Verification Works</h2>

			<div class="grid grid-cols-1 gap-6 md:grid-cols-3">
				<div class="text-center">
					<div
						class="bg-primary/10 mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-full"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="text-primary h-8 w-8"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
							/>
						</svg>
					</div>
					<h3 class="mb-2 font-semibold">1. Cryptographic Hash</h3>
					<p class="text-base-content/70 text-sm">
						Each receipt has a unique content hash generated from payment data
					</p>
				</div>

				<div class="text-center">
					<div
						class="bg-secondary/10 mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-full"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="text-secondary h-8 w-8"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M15 7a2 2 0 012 2m0 0v6a2 2 0 01-2 2H9a2 2 0 01-2-2V9a2 2 0 012-2m6 0V7a2 2 0 00-2-2H9a2 2 0 00-2 2v2m6 0H9"
							/>
						</svg>
					</div>
					<h3 class="mb-2 font-semibold">2. Digital Signature</h3>
					<p class="text-base-content/70 text-sm">
						Signed by sender's private key and verified using public key cryptography
					</p>
				</div>

				<div class="text-center">
					<div
						class="bg-accent/10 mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-full"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="text-accent h-8 w-8"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0121 12a11.955 11.955 0 01-1.382 5.984M15 12a3 3 0 11-6 0 3 3 0 016 0z"
							/>
						</svg>
					</div>
					<h3 class="mb-2 font-semibold">3. Blockchain Verification</h3>
					<p class="text-base-content/70 text-sm">
						Cross-referenced with on-chain payment data for complete authenticity
					</p>
				</div>
			</div>
		</div>
	</div>
</div>
