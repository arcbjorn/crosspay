<script lang="ts">
	import OracleDashboard from '$lib/components/OracleDashboard.svelte';
</script>

<svelte:head>
	<title>Oracle Services - CrossPay</title>
</svelte:head>

<div class="mx-auto max-w-7xl">
	<div class="breadcrumbs mb-8 text-sm">
		<ul>
			<li><a href="/">Home</a></li>
			<li>Oracle Services</li>
		</ul>
	</div>

	<div class="mb-8">
		<h1 class="mb-4 text-4xl font-bold">Oracle Services</h1>
		<p class="text-base-content/70 text-lg">
			Real-time price feeds, secure random numbers, and external data verification powered by Flare
			Network.
		</p>
	</div>

	<OracleDashboard />
</div>
