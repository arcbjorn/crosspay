<script lang="ts">
	import '../app.css';
	import WalletConnect from '$lib/components/WalletConnect.svelte';
	import WagmiProvider from '$lib/components/WagmiProvider.svelte';
	
	let { children } = $props();
</script>

<svelte:head>
	<title>CrossPay Protocol</title>
	<meta name="description" content="Verifiable, private, cross-chain payment infrastructure" />
</svelte:head>

<WagmiProvider>
	<div class="min-h-screen bg-base-100">
		<div class="navbar bg-base-100 border-b border-base-300">
			<div class="navbar-start">
				<a href="/" class="btn btn-ghost text-xl font-bold">
					<span class="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
						CrossPay
					</span>
				</a>
			</div>
			<div class="navbar-center hidden lg:flex">
				<ul class="menu menu-horizontal px-1">
					<li><a href="/">Home</a></li>
					<li><a href="/pay">Pay</a></li>
					<li><a href="/receipts">Receipts</a></li>
				</ul>
			</div>
			<WalletConnect />
		</div>

		<main class="container mx-auto px-4 py-8">
			{@render children?.()}
		</main>

		<footer class="footer footer-center p-4 bg-base-200 text-base-content mt-auto">
			<div>
				<p>© 2025 CrossPay Protocol. Built for the future of payments.</p>
			</div>
		</footer>
	</div>
</WagmiProvider>
