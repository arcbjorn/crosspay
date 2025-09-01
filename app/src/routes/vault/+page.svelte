<script lang="ts">
	import { onMount } from 'svelte';
	import { writable } from 'svelte/store';
	import VaultManager from '$lib/components/VaultManager.svelte';

	let userAddress = '';
	let isConnected = false;

	const portfolioMetrics = writable({
		totalValue: BigInt('7500000000000000000000'), // 7.5k ETH
		totalRewards: BigInt('450000000000000000000'), // 450 ETH
		apy: 12.8,
		riskScore: 'Medium',
		positions: [
			{ tranche: 'senior', amount: BigInt('5000000000000000000000'), apy: 8.5 },
			{ tranche: 'mezzanine', amount: BigInt('2000000000000000000000'), apy: 12.8 },
			{ tranche: 'junior', amount: BigInt('500000000000000000000'), apy: 18.2 }
		]
	});

	onMount(() => {
		// Mock wallet connection
		userAddress = '0x742d35Cc6634C0532925a3b8D4ba9f4e6ad1B6AF';
		isConnected = true;
	});

	function formatEth(wei: bigint): string {
		return (Number(wei) / 1e18).toLocaleString(undefined, { maximumFractionDigits: 2 });
	}

	async function connectWallet() {
		// Mock wallet connection
		userAddress = '0x742d35Cc6634C0532925a3b8D4ba9f4e6ad1B6AF';
		isConnected = true;
	}
</script>

<svelte:head>
	<title>CrossPay Vault - Tranche Yield Farming</title>
	<meta name="description" content="Earn yield through risk-stratified tranche deposits" />
</svelte:head>

<div class="vault-page">
	<header class="page-header">
		<h1>🏦 Tranche Vault</h1>
		<p>Deposit into risk-stratified tranches to earn yield while protecting against losses</p>
	</header>

	{#if !isConnected}
		<div class="connect-wallet">
			<h2>Connect Your Wallet</h2>
			<p>Connect your wallet to view and manage your vault positions</p>
			<button on:click={connectWallet} class="connect-button"> 🔗 Connect Wallet </button>
		</div>
	{:else}
		<!-- Portfolio Overview -->
		<div class="portfolio-overview">
			<h2>Your Portfolio</h2>
			<div class="metrics-grid">
				<div class="metric-card total">
					<h3>Total Value</h3>
					<div class="value">{formatEth($portfolioMetrics.totalValue)} ETH</div>
					<div class="change positive">+{formatEth($portfolioMetrics.totalRewards)} ETH earned</div>
				</div>

				<div class="metric-card apy">
					<h3>Weighted APY</h3>
					<div class="value">{$portfolioMetrics.apy}%</div>
					<div class="change">{$portfolioMetrics.riskScore} Risk</div>
				</div>

				<div class="metric-card positions">
					<h3>Active Positions</h3>
					<div class="value">{$portfolioMetrics.positions.length}</div>
					<div class="change">Across all tranches</div>
				</div>
			</div>

			<!-- Position Breakdown -->
			<div class="position-breakdown">
				<h3>Position Breakdown</h3>
				<div class="positions-list">
					{#each $portfolioMetrics.positions as position}
						<div class="position-item">
							<div class="position-info">
								<span class="tranche-name"
									>{position.tranche.charAt(0).toUpperCase() + position.tranche.slice(1)} Tranche</span
								>
								<span class="position-amount">{formatEth(position.amount)} ETH</span>
							</div>
							<div class="position-apy">
								{position.apy}% APY
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>

		<!-- Vault Manager Component -->
		<VaultManager {userAddress} />

		<!-- Risk Education -->
		<div class="risk-education">
			<h2>Understanding Tranche Risks</h2>
			<div class="risk-grid">
				<div class="risk-card senior">
					<h4>🟢 Senior Tranche</h4>
					<ul>
						<li>Lowest risk, first to receive yields</li>
						<li>Last to absorb losses in liquidation waterfall</li>
						<li>Lower APY but capital protection priority</li>
						<li>24-hour withdrawal delay</li>
					</ul>
				</div>

				<div class="risk-card mezzanine">
					<h4>🟡 Mezzanine Tranche</h4>
					<ul>
						<li>Moderate risk and yield</li>
						<li>Second in liquidation waterfall</li>
						<li>Balanced risk/reward profile</li>
						<li>48-hour withdrawal delay</li>
					</ul>
				</div>

				<div class="risk-card junior">
					<h4>🔴 Junior Tranche</h4>
					<ul>
						<li>Highest risk and potential yield</li>
						<li>First to absorb losses (protection buffer)</li>
						<li>Maximum yield but maximum risk</li>
						<li>72-hour withdrawal delay</li>
					</ul>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.vault-page {
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
		max-width: 600px;
		margin: 0 auto;
	}

	.connect-wallet {
		text-align: center;
		padding: 4rem;
		border: 2px dashed #ddd;
		border-radius: 12px;
		margin: 2rem 0;
	}

	.connect-button {
		padding: 1rem 2rem;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		color: white;
		border: none;
		border-radius: 8px;
		font-size: 1.1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}

	.connect-button:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
	}

	.portfolio-overview {
		background: white;
		border: 1px solid #e0e0e0;
		border-radius: 12px;
		padding: 2rem;
		margin-bottom: 2rem;
	}

	.portfolio-overview h2 {
		margin-bottom: 1.5rem;
		color: #333;
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.metric-card {
		padding: 1.5rem;
		border-radius: 8px;
		text-align: center;
	}

	.metric-card.total {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		color: white;
	}

	.metric-card.apy {
		background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
		color: white;
	}

	.metric-card.positions {
		background: linear-gradient(135deg, #ffc107 0%, #fd7e14 100%);
		color: white;
	}

	.metric-card h3 {
		margin: 0 0 0.5rem 0;
		font-size: 0.9rem;
		opacity: 0.9;
		text-transform: uppercase;
	}

	.metric-card .value {
		font-size: 2rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
	}

	.metric-card .change {
		font-size: 0.9rem;
		opacity: 0.8;
	}

	.position-breakdown h3 {
		margin-bottom: 1rem;
		color: #333;
	}

	.positions-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.position-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		border: 1px solid #eee;
		border-radius: 8px;
		background: #fafafa;
	}

	.position-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.tranche-name {
		font-weight: 600;
		color: #333;
	}

	.position-amount {
		color: #666;
		font-size: 0.9rem;
	}

	.position-apy {
		font-weight: 600;
		color: #28a745;
	}

	.risk-education {
		background: white;
		border: 1px solid #e0e0e0;
		border-radius: 12px;
		padding: 2rem;
		margin-top: 2rem;
	}

	.risk-education h2 {
		margin-bottom: 1.5rem;
		color: #333;
	}

	.risk-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1.5rem;
	}

	.risk-card {
		padding: 1.5rem;
		border-radius: 8px;
		border: 1px solid #eee;
	}

	.risk-card.senior {
		border-left: 4px solid #28a745;
	}

	.risk-card.mezzanine {
		border-left: 4px solid #ffc107;
	}

	.risk-card.junior {
		border-left: 4px solid #dc3545;
	}

	.risk-card h4 {
		margin: 0 0 1rem 0;
		color: #333;
	}

	.risk-card ul {
		margin: 0;
		padding-left: 1.2rem;
	}

	.risk-card li {
		margin-bottom: 0.5rem;
		color: #666;
		line-height: 1.4;
	}

	@media (max-width: 768px) {
		.vault-page {
			padding: 1rem;
		}

		.metrics-grid {
			grid-template-columns: 1fr;
		}

		.risk-grid {
			grid-template-columns: 1fr;
		}

		.position-item {
			flex-direction: column;
			gap: 0.5rem;
			align-items: flex-start;
		}
	}
</style>
