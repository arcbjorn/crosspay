<script lang="ts">
	import '../app.css';
	import WalletConnect from '$lib/components/WalletConnect.svelte';
	import WagmiProvider from '$lib/components/WagmiProvider.svelte';
	import ToastContainer from '$lib/components/ToastContainer.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	
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
				<!-- Mobile menu -->
				<div class="dropdown lg:hidden">
					<div tabindex="0" role="button" class="btn btn-ghost">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h8m-8 6h16" />
						</svg>
					</div>
					<ul role="menu" class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52">
						<li><a href="/">Home</a></li>
						<li><a href="/pay">Pay</a></li>
						<li><a href="/receipts">Receipts</a></li>
						<li>
							<span>Services</span>
							<ul class="p-2">
								<li><a href="/oracles">Oracle Feeds</a></li>
								<li><a href="/storage">Storage</a></li>
								<li><a href="/ens">ENS Names</a></li>
								<li><a href="/verify">Verify Receipt</a></li>
							</ul>
						</li>
					</ul>
				</div>
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
					<li>
						<details>
							<summary>Services</summary>
							<ul class="p-2 bg-base-100 shadow-lg min-w-48">
								<li><a href="/oracles">Oracle Feeds</a></li>
								<li><a href="/storage">Storage</a></li>
								<li><a href="/ens">ENS Names</a></li>
								<li><a href="/verify">Verify Receipt</a></li>
							</ul>
						</details>
					</li>
				</ul>
			</div>
			<div class="navbar-end flex items-center gap-2">
				<ThemeToggle />
				<WalletConnect />
			</div>
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
	
	<!-- Global Toast Container -->
	<ToastContainer />
</WagmiProvider>
