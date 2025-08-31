<script lang="ts">
  import OracleDashboard from '$lib/components/OracleDashboard.svelte';
</script>

<svelte:head>
  <title>Oracle Services - CrossPay</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
  <div class="breadcrumbs text-sm mb-8">
    <ul>
      <li><a href="/">Home</a></li>
      <li>Oracle Services</li>
    </ul>
  </div>

  <div class="mb-8">
    <h1 class="text-4xl font-bold mb-4">Oracle Services</h1>
    <p class="text-lg text-base-content/70">
      Real-time price feeds, secure random numbers, and external data verification powered by Flare Network.
    </p>
  </div>

  <OracleDashboard />
</div>