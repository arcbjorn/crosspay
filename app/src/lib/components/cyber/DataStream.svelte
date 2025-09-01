<script lang="ts">
	export let data: string[] = [];
	export let speed: 'slow' | 'medium' | 'fast' = 'medium';
	export let direction: 'left' | 'right' = 'right';
	export let height = 'h-16';
	export let color: 'mint' | 'lavender' | 'success' | 'warning' | 'error' = 'mint';
	export let showBorder = true;
	export let animated = true;

	$: speedConfig = {
		slow: '4s',
		medium: '2s',
		fast: '1s'
	};

	$: colorConfig = {
		mint: 'rgba(159, 239, 223, 0.2)',
		lavender: 'rgba(201, 179, 255, 0.2)',
		success: 'rgba(179, 255, 186, 0.2)',
		warning: 'rgba(255, 228, 181, 0.2)',
		error: 'rgba(255, 179, 186, 0.2)'
	};

	$: streamId = `stream-${Math.random().toString(36).substr(2, 9)}`;
</script>

<div
	class="data-stream relative overflow-hidden {height} {showBorder
		? 'border-cyber-border- border' + color
		: ''} font-mono"
	class:bg-cyber-surface-1={showBorder}
>
	<!-- Animated sweep effect -->
	{#if animated}
		<div
			class="pointer-events-none absolute inset-0 z-10"
			style="background: linear-gradient(90deg, transparent, {colorConfig[color]}, transparent); 
             animation: sweep-{direction} {speedConfig[speed]} ease-in-out infinite;"
		></div>
	{/if}

	<!-- Data content -->
	<div class="relative z-0 flex h-full items-center">
		{#if data.length > 0}
			<!-- Scrolling data text -->
			<div
				class="flex items-center gap-8 whitespace-nowrap text-cyber-text-secondary"
				class:animate-scroll-left={direction === 'left' && animated}
				class:animate-scroll-right={direction === 'right' && animated}
			>
				{#each data as item, i}
					<span class="flex items-center gap-2">
						<span class="text-cyber-{color} text-xs">
							[{i.toString().padStart(2, '0')}]
						</span>
						<span class="text-sm">{item}</span>
					</span>
				{/each}
			</div>
		{:else}
			<!-- Empty state with data blocks -->
			<div class="flex w-full items-center justify-center text-cyber-text-tertiary">
				<div class="flex gap-2">
					{#each Array(12) as _, i}
						<div
							class="h-6 w-3 border border-cyber-text-tertiary/20"
							class:bg-cyber-{color}={i % 3 === 0 && animated}
							class:animate-pulse={i % 3 === 0 && animated}
							style="animation-delay: {i * 0.1}s"
						></div>
					{/each}
				</div>
			</div>
		{/if}
	</div>

	<!-- Terminal grid pattern overlay -->
	<div
		class="pointer-events-none absolute inset-0 opacity-30"
		style="
    background-image: 
      linear-gradient(rgba(159, 239, 223, 0.1) 1px, transparent 1px),
      linear-gradient(90deg, rgba(159, 239, 223, 0.1) 1px, transparent 1px);
    background-size: 10px 10px;
  "
	></div>

	<!-- Connection indicators -->
	<div class="absolute left-2 top-2 flex gap-1">
		<div class="h-2 w-2 bg-cyber-{color} animate-pulse"></div>
		<div class="h-2 w-2 bg-cyber-{color} animate-pulse" style="animation-delay: 0.5s"></div>
	</div>

	<div class="absolute right-2 top-2 font-mono text-xs text-cyber-text-tertiary">
		{#if data.length > 0}
			STREAM_ACTIVE
		{:else}
			WAITING_DATA
		{/if}
	</div>
</div>

<style>
	@keyframes sweep-right {
		0% {
			transform: translateX(-100%) skewX(-12deg);
		}
		100% {
			transform: translateX(200%) skewX(-12deg);
		}
	}

	@keyframes sweep-left {
		0% {
			transform: translateX(200%) skewX(12deg);
		}
		100% {
			transform: translateX(-100%) skewX(12deg);
		}
	}

	@keyframes scroll-right {
		0% {
			transform: translateX(-100%);
		}
		100% {
			transform: translateX(100%);
		}
	}

	@keyframes scroll-left {
		0% {
			transform: translateX(100%);
		}
		100% {
			transform: translateX(-100%);
		}
	}

	.animate-scroll-right {
		animation: scroll-right 10s linear infinite;
	}

	.animate-scroll-left {
		animation: scroll-left 10s linear infinite;
	}
</style>
