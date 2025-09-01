<script lang="ts">
	import '../app.css';
	// import { walletStore } from '$lib/stores/wallet';
	import { onMount } from 'svelte';

	let typewriterText = '';
	const fullText = 'CrossPay Protocol v2.1.0';

	onMount(() => {
		let i = 0;
		const timer = setInterval(() => {
			typewriterText = fullText.slice(0, i);
			i++;
			if (i > fullText.length) {
				clearInterval(timer);
			}
		}, 100);

		return () => clearInterval(timer);
	});

	// $: wallet = $walletStore;
	const wallet = { isConnected: false }; // Placeholder
</script>

<!-- Terminal Header -->
<div class="relative flex min-h-screen flex-col">
	<!-- ASCII Art Header -->
	<header class="terminal-text relative z-10 p-8 text-center">
		<pre class="mb-4 text-sm text-cyber-mint md:text-base">
{`
╔═══════════════════════════════════════════════════════════════╗
║  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  ║
║  ░░ ██████╗██████╗  ██████╗ ███████╗███████╗██████╗  █████╗  ░░  ║
║  ░░██╔════╝██╔══██╗██╔═══██╗██╔════╝██╔════╝██╔══██╗██╔══██╗ ░░  ║
║  ░░██║     ██████╔╝██║   ██║███████╗███████╗██████╔╝███████║ ░░  ║
║  ░░██║     ██╔══██╗██║   ██║╚════██║╚════██║██╔═══╝ ██╔══██║ ░░  ║
║  ░░╚██████╗██║  ██║╚██████╔╝███████║███████║██║     ██║  ██║ ░░  ║
║  ░░ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚══════╝╚═╝     ╚═╝  ╚═╝ ░░  ║
║  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  ║
╚═══════════════════════════════════════════════════════════════╝
`}
    </pre>

		<div class="mb-2 font-mono text-cyber-text-primary">
			> Initializing {typewriterText}<span class="animate-cursor-blink">|</span>
		</div>

		<div class="terminal-text text-sm text-cyber-text-secondary">
			[SYSTEM STATUS: ONLINE] [VALIDATORS: ACTIVE] [PRIVACY: ENABLED]
		</div>
	</header>

	<!-- Main Hero Section -->
	<main class="relative z-10 flex flex-1 items-center justify-center px-4">
		<div class="max-w-4xl text-center">
			<h1
				class="glitch-text mb-6 font-mono text-2xl text-cyber-text-primary md:text-4xl"
				data-text="VERIFIABLE CROSS-CHAIN PAYMENTS"
			>
				VERIFIABLE CROSS-CHAIN PAYMENTS
			</h1>

			<p class="mx-auto mb-8 max-w-2xl text-lg leading-relaxed text-cyber-text-secondary">
				Secure payment infrastructure across 9 blockchain networks. Private transactions with
				selective disclosure. Cryptographic proofs for high-value transfers.
			</p>

			{#if wallet.isConnected}
				<div class="flex flex-col justify-center gap-4 sm:flex-row">
					<a href="/pay" class="cyber-btn-primary glitch-hover px-8 py-3 text-lg">
						[INITIATE_PAYMENT]
					</a>
					<a href="/receipts" class="cyber-btn-secondary glitch-hover px-8 py-3 text-lg">
						[VIEW_RECEIPTS]
					</a>
				</div>
			{:else}
				<div class="cyber-card mx-auto mb-8 max-w-md">
					<div class="flex items-center gap-3">
						<div class="terminal-text text-cyber-warning">⚠</div>
						<span class="terminal-text">WALLET_CONNECTION_REQUIRED</span>
					</div>
					<div class="terminal-text mt-2 text-sm text-cyber-text-tertiary">
						> Execute wallet.connect() to proceed
					</div>
				</div>

				<button class="cyber-btn-primary glitch-hover px-8 py-3 text-lg"> [CONNECT_WALLET] </button>
			{/if}
		</div>
	</main>

	<!-- Feature Matrix -->
	<section class="relative z-10 py-16">
		<div class="container mx-auto px-4">
			<div class="terminal-text mb-12 text-center">
				<div class="mb-2 text-lg text-cyber-mint">SYSTEM_MODULES:</div>
			</div>

			<div class="mx-auto grid max-w-6xl gap-6 md:grid-cols-3">
				<!-- Privacy Module -->
				<div class="cyber-card group transition-all duration-300 hover:border-cyber-mint-hover">
					<div class="security-matrix mb-4">
						{#each Array(16) as _, i}
							<div class="matrix-cell {i % 7 === 0 ? 'active' : ''}"></div>
						{/each}
					</div>

					<div class="terminal-text mb-2 text-lg text-cyber-mint">[PRIVACY_SHIELD]</div>
					<div class="mb-4 text-sm text-cyber-text-secondary">
						MODULE_STATUS: ACTIVE<br />
						ENCRYPTION: ZAMA_FHE<br />
						DISCLOSURE: SELECTIVE
					</div>
					<div class="text-xs text-cyber-text-tertiary">
						Payment amounts encrypted with homomorphic encryption. Role-based decryption for
						compliance and auditing.
					</div>
				</div>

				<!-- Cross-Chain Module -->
				<div class="cyber-card group transition-all duration-300 hover:border-cyber-lavender-hover">
					<div class="data-stream mb-4 flex h-16 items-center justify-center">
						<div class="terminal-text text-cyber-lavender">
							╔═══╗ ═══════════ ╔═══╗<br />
							║ A ║ ═══════════ ║ B ║<br />
							╚═══╝ ═══════════ ╚═══╝
						</div>
					</div>

					<div class="terminal-text mb-2 text-lg text-cyber-lavender">[CROSS_CHAIN]</div>
					<div class="mb-4 text-sm text-cyber-text-secondary">
						NETWORKS: 9_ACTIVE<br />
						ADAPTERS: DEPLOYED<br />
						INTEROP: UNIFIED
					</div>
					<div class="text-xs text-cyber-text-tertiary">
						Seamless payments across Ethereum, Base, Lisk, Polygon, and 5 additional networks with
						unified infrastructure.
					</div>
				</div>

				<!-- Security Module -->
				<div class="cyber-card group transition-all duration-300 hover:border-cyber-success">
					<div class="mb-4 flex justify-center">
						<div class="terminal-text text-center text-cyber-success">
							┌─────────────────┐<br />
							│ VALIDATOR_POOL │<br />
							│ ███████████████ │<br />
							│ STATUS: SECURED │<br />
							└─────────────────┘
						</div>
					</div>

					<div class="terminal-text mb-2 text-lg text-cyber-success">[VALIDATOR_NETWORK]</div>
					<div class="mb-4 text-sm text-cyber-text-secondary">
						CONSENSUS: BFT<br />
						PROOFS: CRYPTOGRAPHIC<br />
						SECURITY: SYMBIOTIC
					</div>
					<div class="text-xs text-cyber-text-tertiary">
						High-value transactions secured by distributed validator network with cryptographic
						proof aggregation.
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Terminal Footer -->
	<footer class="terminal-text relative z-10 py-8 text-center text-xs text-cyber-text-tertiary">
		<div>CROSSPAY_PROTOCOL © 2024 | SECURITY_LEVEL: MAXIMUM | UPTIME: 99.99%</div>
		<div class="mt-1">
			VERSION: 2.1.0 | BUILD: STABLE | LAST_DEPLOY: {new Date().toISOString().split('T')[0]}
		</div>
	</footer>
</div>

<!-- Background Effects -->
<div class="pointer-events-none fixed inset-0">
	<!-- Matrix Rain Effect (subtle) -->
	<div class="absolute inset-0 opacity-5">
		{#each Array(20) as _, i}
			<div
				class="absolute w-px animate-matrix-rain bg-cyber-mint"
				style="left: {Math.random() * 100}%; animation-delay: {Math.random() * 10}s; height: 100vh;"
			></div>
		{/each}
	</div>
</div>
