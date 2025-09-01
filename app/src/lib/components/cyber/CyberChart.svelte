<script lang="ts">
	export let data: { label: string; value: number; color?: string }[] = [];
	export let title = '';
	export let type: 'bar' | 'line' = 'bar';
	export let height = 200;
	export let showGrid = true;
	export let animated = true;

	$: maxValue = Math.max(...data.map((d) => d.value), 1);
	$: chartId = `chart-${Math.random().toString(36).substr(2, 9)}`;

	function getBarHeight(value: number): number {
		return (value / maxValue) * (height - 40);
	}

	function getColor(item: { color?: string }, index: number): string {
		if (item.color) return item.color;
		const colors = [
			'rgba(159, 239, 223, 0.6)', // cyber-mint
			'rgba(201, 179, 255, 0.5)', // cyber-lavender
			'rgba(179, 255, 186, 0.5)', // cyber-success
			'rgba(255, 228, 181, 0.6)', // cyber-warning
			'rgba(255, 179, 186, 0.5)' // cyber-error
		];
		return colors[index % colors.length];
	}

	function formatValue(value: number): string {
		if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`;
		if (value >= 1000) return `${(value / 1000).toFixed(1)}K`;
		return value.toString();
	}
</script>

<div class="cyber-chart font-mono text-cyber-text-secondary">
	{#if title}
		<div class="terminal-text mb-4 flex items-center gap-2 text-lg text-cyber-mint">
			<span class="text-cyber-text-tertiary">></span>
			{title}
			<span class="animate-cursor-blink text-cyber-mint">|</span>
		</div>
	{/if}

	<div class="chart-container relative border border-cyber-border-mint bg-cyber-surface-1 p-4">
		<!-- Grid lines -->
		{#if showGrid}
			<div class="pointer-events-none absolute inset-4">
				{#each Array(5) as _, i}
					<div
						class="absolute w-full border-t border-cyber-text-tertiary/20"
						style="top: {(i / 4) * 100}%"
					></div>
				{/each}
			</div>
		{/if}

		<!-- Chart content -->
		<div class="relative z-10" style="height: {height}px">
			{#if type === 'bar'}
				<!-- Bar chart -->
				<div class="flex h-full items-end justify-between gap-2 px-4 pb-8">
					{#each data as item, i}
						<div class="flex min-w-0 flex-1 flex-col items-center">
							<!-- Bar -->
							<div
								class="group relative w-full max-w-12 overflow-hidden border border-cyber-border-mint transition-all duration-300"
								style="height: {getBarHeight(item.value)}px; background: {getColor(item, i)}"
								class:animate-pulse={animated}
							>
								<!-- Scan line effect -->
								<div
									class="absolute inset-0 bg-gradient-to-t from-transparent via-white/10 to-transparent opacity-0 transition-opacity duration-300 group-hover:opacity-100"
								></div>

								<!-- Value display on hover -->
								<div
									class="absolute -top-8 left-1/2 -translate-x-1/2 transform whitespace-nowrap text-xs text-cyber-mint opacity-0 transition-opacity group-hover:opacity-100"
								>
									{formatValue(item.value)}
								</div>
							</div>

							<!-- Label -->
							<div class="mt-2 w-full truncate text-center text-xs text-cyber-text-tertiary">
								{item.label}
							</div>
						</div>
					{/each}
				</div>
			{:else}
				<!-- Line chart -->
				<div class="relative h-full p-4">
					<svg width="100%" height="100%" class="overflow-visible">
						<!-- Grid -->
						{#if showGrid}
							<defs>
								<pattern id="grid-{chartId}" width="20" height="20" patternUnits="userSpaceOnUse">
									<path
										d="M 20 0 L 0 0 0 20"
										fill="none"
										stroke="rgba(112, 112, 112, 0.2)"
										stroke-width="1"
									/>
								</pattern>
							</defs>
							<rect width="100%" height="100%" fill="url(#grid-{chartId})" />
						{/if}

						<!-- Line path -->
						{#if data.length > 1}
							<path
								d="M {data
									.map(
										(item, i) =>
											`${(i / (data.length - 1)) * 100}% ${100 - (item.value / maxValue) * 100}%`
									)
									.join(' L ')}"
								fill="none"
								stroke="rgba(159, 239, 223, 0.8)"
								stroke-width="2"
								class="transition-all duration-500"
								class:animate-pulse={animated}
							/>
						{/if}

						<!-- Data points -->
						{#each data as item, i}
							<circle
								cx="{(i / (data.length - 1)) * 100}%"
								cy="{100 - (item.value / maxValue) * 100}%"
								r="4"
								fill={getColor(item, i)}
								stroke="rgba(159, 239, 223, 0.8)"
								stroke-width="2"
								class="hover:r-6 transition-all duration-200"
							>
								<title>{item.label}: {formatValue(item.value)}</title>
							</circle>
						{/each}
					</svg>

					<!-- Labels -->
					<div
						class="absolute bottom-0 left-0 right-0 flex justify-between px-4 text-xs text-cyber-text-tertiary"
					>
						{#each data as item}
							<span class="truncate">{item.label}</span>
						{/each}
					</div>
				</div>
			{/if}
		</div>

		<!-- Y-axis labels -->
		<div
			class="absolute bottom-8 left-0 top-4 flex w-8 flex-col justify-between text-xs text-cyber-text-tertiary"
		>
			{#each Array(5) as _, i}
				<span class="pr-1 text-right">
					{formatValue((maxValue * (4 - i)) / 4)}
				</span>
			{/each}
		</div>

		<!-- Terminal-style corner brackets -->
		<div class="absolute left-0 top-0 h-4 w-4 border-l border-t border-cyber-border-mint"></div>
		<div class="absolute right-0 top-0 h-4 w-4 border-r border-t border-cyber-border-mint"></div>
		<div class="absolute bottom-0 left-0 h-4 w-4 border-b border-l border-cyber-border-mint"></div>
		<div class="absolute bottom-0 right-0 h-4 w-4 border-b border-r border-cyber-border-mint"></div>
	</div>

	<!-- Legend/Stats -->
	{#if data.length > 0}
		<div class="mt-4 text-xs text-cyber-text-tertiary">
			<div class="flex gap-4">
				<span>SAMPLES: {data.length}</span>
				<span>MAX: {formatValue(maxValue)}</span>
				<span
					>AVG: {formatValue(data.reduce((sum, item) => sum + item.value, 0) / data.length)}</span
				>
			</div>
		</div>
	{/if}
</div>

<style>
	.cyber-chart .chart-container::before {
		content: '';
		position: absolute;
		inset: 0;
		background: repeating-linear-gradient(
			0deg,
			transparent,
			transparent 2px,
			rgba(159, 239, 223, 0.02) 2px,
			rgba(159, 239, 223, 0.02) 4px
		);
		pointer-events: none;
		z-index: 1;
	}
</style>
