<script lang="ts">
	import CyberCard from './CyberCard.svelte';
	import { onMount } from 'svelte';

	export let validators: Array<{
		id: string;
		address: string;
		status: 'active' | 'inactive' | 'syncing';
		stake: number;
		uptime: number;
	}> = [];

	export let consensusStatus: 'finalized' | 'pending' | 'error' = 'pending';
	export let securityLevel: 'low' | 'medium' | 'high' | 'maximum' = 'high';

	let matrixCells = Array(64)
		.fill(0)
		.map((_, i) => ({
			id: i,
			active: false,
			type: Math.random() > 0.7 ? 'validator' : 'data',
			strength: Math.random()
		}));

	let animationFrame = 0;

	onMount(() => {
		const interval = setInterval(() => {
			matrixCells = matrixCells.map((cell) => ({
				...cell,
				active: Math.random() > 0.8,
				strength: Math.random()
			}));
			animationFrame++;
		}, 1000);

		return () => clearInterval(interval);
	});

	$: activeValidators = validators.filter((v) => v.status === 'active').length;
	$: totalStake = validators.reduce((sum, v) => sum + v.stake, 0);
	$: averageUptime =
		validators.length > 0
			? validators.reduce((sum, v) => sum + v.uptime, 0) / validators.length
			: 0;

	$: securityColor = {
		low: 'text-cyber-error',
		medium: 'text-cyber-warning',
		high: 'text-cyber-mint',
		maximum: 'text-cyber-success'
	}[securityLevel];

	$: consensusColor = {
		finalized: 'text-cyber-success',
		pending: 'text-cyber-warning',
		error: 'text-cyber-error'
	}[consensusStatus];
</script>

<CyberCard variant="default" padding="lg">
	<div class="terminal-text mb-6 text-xl text-cyber-mint">[SECURITY_MATRIX]</div>

	<!-- Matrix Visualization -->
	<div
		class="mb-6 grid grid-cols-8 gap-1 border border-cyber-border-mint/20 bg-cyber-bg-secondary/50 p-4"
	>
		{#each matrixCells as cell}
			<div
				class="aspect-square border transition-all duration-500 {cell.active
					? cell.type === 'validator'
						? 'animate-pulse border-cyber-success bg-cyber-success'
						: 'border-cyber-mint bg-cyber-mint'
					: 'border-cyber-text-tertiary/20 bg-cyber-surface-2'}"
				style="opacity: {cell.strength * 0.8 + 0.2}"
			></div>
		{/each}
	</div>

	<!-- Status Grid -->
	<div class="mb-6 grid gap-6 md:grid-cols-2">
		<!-- Consensus Status -->
		<div class="terminal-text">
			<div class="mb-2 text-sm text-cyber-text-secondary">CONSENSUS_STATUS:</div>
			<div class="{consensusColor} font-mono text-lg">
				[{consensusStatus.toUpperCase()}]
			</div>
			{#if consensusStatus === 'pending'}
				<div class="mt-1 animate-pulse text-xs text-cyber-text-tertiary">
					> Waiting for validator signatures...
				</div>
			{:else if consensusStatus === 'finalized'}
				<div class="mt-1 text-xs text-cyber-text-tertiary">> Block finalized and verified</div>
			{:else}
				<div class="mt-1 text-xs text-cyber-text-tertiary">> Consensus failure detected</div>
			{/if}
		</div>

		<!-- Security Level -->
		<div class="terminal-text">
			<div class="mb-2 text-sm text-cyber-text-secondary">SECURITY_LEVEL:</div>
			<div class="{securityColor} font-mono text-lg">
				[{securityLevel.toUpperCase()}]
			</div>
			<div class="mt-1 text-xs text-cyber-text-tertiary">
				> {activeValidators}/{validators.length} validators active
			</div>
		</div>
	</div>

	<!-- Validator Stats -->
	<div class="border-t border-cyber-border-mint/20 pt-4">
		<div class="terminal-text mb-3 text-sm text-cyber-text-secondary">VALIDATOR_METRICS:</div>

		<div class="grid grid-cols-3 gap-4 text-center">
			<div class="cyber-card bg-cyber-surface-2 p-3">
				<div class="font-mono text-lg text-cyber-mint">
					{activeValidators}
				</div>
				<div class="text-xs text-cyber-text-tertiary">ACTIVE_NODES</div>
			</div>

			<div class="cyber-card bg-cyber-surface-2 p-3">
				<div class="font-mono text-lg text-cyber-lavender">
					{totalStake.toFixed(2)}K
				</div>
				<div class="text-xs text-cyber-text-tertiary">TOTAL_STAKE</div>
			</div>

			<div class="cyber-card bg-cyber-surface-2 p-3">
				<div class="font-mono text-lg text-cyber-success">
					{averageUptime.toFixed(1)}%
				</div>
				<div class="text-xs text-cyber-text-tertiary">AVG_UPTIME</div>
			</div>
		</div>
	</div>

	<!-- Individual Validators -->
	{#if validators.length > 0}
		<div class="mt-6 border-t border-cyber-border-mint/20 pt-4">
			<div class="terminal-text mb-3 text-sm text-cyber-text-secondary">VALIDATOR_LIST:</div>

			<div class="max-h-32 space-y-2 overflow-y-auto">
				{#each validators.slice(0, 5) as validator}
					<div class="terminal-text flex items-center justify-between text-xs">
						<div class="flex items-center gap-3">
							<div
								class="h-2 w-2 rounded-full {validator.status === 'active'
									? 'animate-pulse bg-cyber-success'
									: validator.status === 'syncing'
										? 'bg-cyber-warning'
										: 'bg-cyber-error'}"
							></div>
							<span class="font-mono text-cyber-text-primary">
								{validator.address.slice(0, 8)}...{validator.address.slice(-6)}
							</span>
						</div>

						<div class="flex gap-4 text-cyber-text-tertiary">
							<span>{validator.stake.toFixed(1)}K</span>
							<span>{validator.uptime.toFixed(1)}%</span>
							<span class="uppercase">{validator.status}</span>
						</div>
					</div>
				{/each}

				{#if validators.length > 5}
					<div class="pt-2 text-center text-xs text-cyber-text-tertiary">
						... and {validators.length - 5} more validators
					</div>
				{/if}
			</div>
		</div>
	{/if}
</CyberCard>
