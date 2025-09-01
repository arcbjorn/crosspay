<script lang="ts">
	import { onMount } from 'svelte';
	import { walletStore } from '@stores/wallet';
	import { ensService } from '@services/ens';
	import { successToast, errorToast } from '@stores/toast';
	import type { Address } from 'viem';
	import type { ENSResolutionResponse } from '@services/ens';

	$: wallet = $walletStore;

	let searchQuery = '';
	let resolveResult: ENSResolutionResponse | null = null;
	let isResolving = false;
	let userENSName = '';
	let isLoadingUserENS = false;

	// Subname registration
	let subname = '';
	let domain = 'crosspay.eth';
	let isRegistering = false;

	onMount(async () => {
		if (wallet.isConnected && wallet.address) {
			await loadUserENS();
		}
	});

	$: if (wallet.isConnected && wallet.address) {
		loadUserENS();
	}

	async function loadUserENS() {
		if (!wallet.address) return;

		isLoadingUserENS = true;
		try {
			const name = await ensService.lookupAddress(wallet.address as Address);
			userENSName = name || '';
		} catch (error) {
			console.error('Failed to lookup user ENS:', error);
		} finally {
			isLoadingUserENS = false;
		}
	}

	async function resolveENS() {
		if (!searchQuery.trim()) return;

		isResolving = true;
		resolveResult = null;

		try {
			if (searchQuery.endsWith('.eth')) {
				const profile = await ensService.getProfile(searchQuery);
				resolveResult = profile;
			} else if (searchQuery.startsWith('0x')) {
				const name = await ensService.lookupAddress(searchQuery as Address);
				if (name) {
					const profile = await ensService.getProfile(name);
					resolveResult = profile;
				}
			}

			if (!resolveResult) {
				errorToast('Could not resolve ENS name or address');
			}
		} catch (error) {
			console.error('ENS resolution failed:', error);
			errorToast('ENS resolution failed');
		} finally {
			isResolving = false;
		}
	}

	async function registerSubname() {
		if (!wallet.isConnected || !wallet.address) {
			errorToast('Please connect your wallet');
			return;
		}

		if (!subname.trim()) {
			errorToast('Please enter a subname');
			return;
		}

		isRegistering = true;
		try {
			const result = await ensService.registerSubname(
				subname,
				domain,
				wallet.address as Address,
				wallet.address as Address
			);

			if (result?.success) {
				successToast(`Successfully registered ${subname}.${domain}`);
				subname = '';
			} else {
				errorToast(result?.error || 'Registration failed');
			}
		} catch (error) {
			console.error('Subname registration failed:', error);
			errorToast('Registration failed');
		} finally {
			isRegistering = false;
		}
	}

	function handleSearchKeyPress(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			resolveENS();
		}
	}
</script>

<svelte:head>
	<title>ENS Management - CrossPay</title>
</svelte:head>

<div class="mx-auto max-w-6xl">
	<div class="breadcrumbs mb-8 text-sm">
		<ul>
			<li><a href="/">Home</a></li>
			<li>ENS Management</li>
		</ul>
	</div>

	<div class="mb-8">
		<h1 class="mb-4 text-4xl font-bold">ENS Management</h1>
		<p class="text-base-content/70 text-lg">
			Resolve ENS names, manage subdomains, and integrate human-readable addresses into your
			payments.
		</p>
	</div>

	<div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
		<!-- ENS Resolution -->
		<div class="card bg-base-100 shadow-xl">
			<div class="card-body">
				<h2 class="card-title mb-4">ENS Resolution</h2>

				<div class="form-control">
					<label class="label" for="ens-search">
						<span class="label-text">ENS Name or Address</span>
					</label>
					<div class="flex gap-2">
						<input
							type="text"
							id="ens-search"
							placeholder="vitalik.eth or 0x..."
							class="input input-bordered flex-1"
							bind:value={searchQuery}
							on:keypress={handleSearchKeyPress}
						/>
						<button
							class="btn btn-primary"
							on:click={resolveENS}
							disabled={isResolving || !searchQuery.trim()}
						>
							{#if isResolving}
								<span class="loading loading-spinner loading-sm"></span>
							{:else}
								Resolve
							{/if}
						</button>
					</div>
				</div>

				{#if resolveResult}
					<div class="bg-base-200 mt-6 rounded-lg p-4">
						<div class="space-y-3">
							<div>
								<span class="text-base-content/70 text-sm font-medium">Address:</span>
								<div class="break-all font-mono text-sm">{resolveResult.address}</div>
							</div>

							{#if resolveResult.name}
								<div>
									<span class="text-base-content/70 text-sm font-medium">ENS Name:</span>
									<div class="font-medium">{resolveResult.name}</div>
								</div>
							{/if}

							{#if resolveResult.avatar}
								<div>
									<span class="text-base-content/70 text-sm font-medium">Avatar:</span>
									<div class="flex items-center gap-2">
										<img src={resolveResult.avatar} alt="Avatar" class="h-8 w-8 rounded-full" />
										<span class="text-sm">{resolveResult.avatar}</span>
									</div>
								</div>
							{/if}

							{#if resolveResult.email}
								<div>
									<span class="text-base-content/70 text-sm font-medium">Email:</span>
									<div>{resolveResult.email}</div>
								</div>
							{/if}

							{#if resolveResult.twitter}
								<div>
									<span class="text-base-content/70 text-sm font-medium">Twitter:</span>
									<div>@{resolveResult.twitter}</div>
								</div>
							{/if}

							{#if resolveResult.github}
								<div>
									<span class="text-base-content/70 text-sm font-medium">GitHub:</span>
									<div>{resolveResult.github}</div>
								</div>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		</div>

		<!-- Your ENS -->
		<div class="card bg-base-100 shadow-xl">
			<div class="card-body">
				<h2 class="card-title mb-4">Your ENS</h2>

				{#if !wallet.isConnected}
					<div class="alert alert-warning">
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
								d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
							/>
						</svg>
						<span>Connect your wallet to view your ENS information</span>
					</div>
				{:else}
					<div class="space-y-4">
						<div>
							<span class="text-base-content/70 text-sm font-medium">Wallet Address:</span>
							<div class="break-all font-mono text-sm">{wallet.address}</div>
						</div>

						<div>
							<span class="text-base-content/70 text-sm font-medium">ENS Name:</span>
							{#if isLoadingUserENS}
								<div class="skeleton h-4 w-32"></div>
							{:else if userENSName}
								<div class="text-primary font-medium">{userENSName}</div>
							{:else}
								<div class="text-base-content/70">No ENS name registered</div>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Subname Registration -->
	<div class="card bg-base-100 mt-8 shadow-xl">
		<div class="card-body">
			<h2 class="card-title mb-4">Register CrossPay Subname</h2>

			<p class="text-base-content/70 mb-4">
				Get your own CrossPay subdomain for easier payment identification.
			</p>

			<div class="form-control max-w-md">
				<label class="label" for="subname-input">
					<span class="label-text">Choose your subname</span>
				</label>
				<div class="flex">
					<input
						id="subname-input"
						type="text"
						placeholder="yourname"
						class="input input-bordered flex-1 rounded-r-none"
						bind:value={subname}
						disabled={isRegistering || !wallet.isConnected}
					/>
					<span
						class="bg-base-200 border-base-300 text-base-content/70 rounded-r-md border border-l-0 px-3 py-2"
					>
						.{domain}
					</span>
				</div>
				{#if subname}
					<div class="label">
						<span class="label-text-alt">Your subname: <strong>{subname}.{domain}</strong></span>
					</div>
				{/if}
			</div>

			<div class="card-actions mt-6 justify-end">
				<button
					class="btn btn-primary"
					on:click={registerSubname}
					disabled={isRegistering || !wallet.isConnected || !subname.trim()}
				>
					{#if isRegistering}
						<span class="loading loading-spinner loading-sm"></span>
						Registering...
					{:else}
						Register Subname
					{/if}
				</button>
			</div>

			{#if !wallet.isConnected}
				<div class="alert alert-info mt-4">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						class="h-6 w-6 shrink-0 stroke-current"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						></path>
					</svg>
					<span>Connect your wallet to register a subname</span>
				</div>
			{/if}
		</div>
	</div>

	<!-- ENS Features -->
	<div class="mt-8 grid grid-cols-1 gap-6 md:grid-cols-3">
		<div class="card from-primary/10 to-primary/5 bg-gradient-to-br shadow-lg">
			<div class="card-body text-center">
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
							d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
						/>
					</svg>
				</div>
				<h3 class="mb-2 font-semibold">Human-Readable Names</h3>
				<p class="text-base-content/70 text-sm">
					Use simple names like alice.eth instead of complex addresses
				</p>
			</div>
		</div>

		<div class="card from-secondary/10 to-secondary/5 bg-gradient-to-br shadow-lg">
			<div class="card-body text-center">
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
							d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0121 12a11.955 11.955 0 01-1.382 5.984M15 12a3 3 0 11-6 0 3 3 0 016 0z"
						/>
					</svg>
				</div>
				<h3 class="mb-2 font-semibold">Decentralized</h3>
				<p class="text-base-content/70 text-sm">
					Built on Ethereum blockchain with no central authority
				</p>
			</div>
		</div>

		<div class="card from-accent/10 to-accent/5 bg-gradient-to-br shadow-lg">
			<div class="card-body text-center">
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
							d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
						/>
					</svg>
				</div>
				<h3 class="mb-2 font-semibold">Rich Profiles</h3>
				<p class="text-base-content/70 text-sm">
					Link social profiles, avatars, and contact information
				</p>
			</div>
		</div>
	</div>
</div>
