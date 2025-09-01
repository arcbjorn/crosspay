<script lang="ts">
	export let variant: 'default' | 'mint' | 'lavender' | 'danger' = 'default';
	export let padding: 'sm' | 'md' | 'lg' = 'md';
	export let scan = true;
	export let glow = false;
	export let matrix = false;

	$: variantClass =
		variant === 'mint'
			? 'border-cyber-border-mint'
			: variant === 'lavender'
				? 'border-cyber-border-lavender'
				: variant === 'danger'
					? 'border-cyber-error'
					: 'border-cyber-border-mint';

	$: paddingClass = padding === 'sm' ? 'p-3' : padding === 'lg' ? 'p-8' : 'p-6';

	$: cardClasses = [
		'cyber-card',
		variantClass,
		paddingClass,
		glow ? 'animate-terminal-glow' : '',
		matrix ? 'matrix-pattern' : ''
	]
		.filter(Boolean)
		.join(' ');
</script>

<div class={cardClasses} {...$$restProps}>
	{#if matrix}
		<div class="security-matrix absolute right-2 top-2 opacity-20">
			{#each Array(9) as _, i}
				<div class="matrix-cell h-2 w-2 {i % 3 === 0 ? 'active' : ''}"></div>
			{/each}
		</div>
	{/if}

	<slot />

	{#if scan}
		<div class="scan-line-card"></div>
	{/if}
</div>

<style>
	.scan-line-card::after {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 1px;
		background: linear-gradient(90deg, transparent, rgba(159, 239, 223, 0.4), transparent);
		animation: scanLineCard 4s linear infinite;
		pointer-events: none;
	}

	@keyframes scanLineCard {
		0% {
			transform: translateY(0);
			opacity: 0;
		}
		10% {
			opacity: 1;
		}
		90% {
			opacity: 1;
		}
		100% {
			transform: translateY(100px);
			opacity: 0;
		}
	}

	.matrix-pattern::before {
		background-image:
			radial-gradient(rgba(159, 239, 223, 0.03) 1px, transparent 1px),
			radial-gradient(rgba(201, 179, 255, 0.02) 1px, transparent 1px);
		background-size:
			10px 10px,
			15px 15px;
		background-position:
			0 0,
			5px 5px;
	}
</style>
