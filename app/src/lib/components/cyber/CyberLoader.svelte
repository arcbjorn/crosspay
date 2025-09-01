<script lang="ts">
	export let type: 'spinner' | 'dots' | 'progress' | 'matrix' = 'spinner';
	export let size: 'sm' | 'md' | 'lg' = 'md';
	export let message = '';

	let frame = 0;

	const spinnerFrames = ['/', '-', '\\', '|'];
	const matrixFrames = ['█', '▓', '▒', '░', ' ', '░', '▒', '▓'];

	setInterval(() => {
		frame = (frame + 1) % (type === 'matrix' ? matrixFrames.length : spinnerFrames.length);
	}, 200);

	$: sizeClasses = {
		sm: 'text-sm',
		md: 'text-base',
		lg: 'text-xl'
	}[size];
</script>

<div class="ascii-loader {sizeClasses} flex flex-col items-center gap-2">
	{#if type === 'spinner'}
		<div class="animate-pulse font-mono">
			[{spinnerFrames[frame]}] PROCESSING...
		</div>
	{:else if type === 'dots'}
		<div class="animate-pulse font-mono">
			LOADING{'.'.repeat((frame % 3) + 1)}
		</div>
	{:else if type === 'progress'}
		<div class="font-mono">
			<div class="mb-1">PROGRESS:</div>
			<div class="relative h-2 w-32 border border-cyber-border-mint">
				<div
					class="h-full animate-pulse bg-cyber-mint"
					style="width: {((frame % 10) + 1) * 10}%"
				></div>
			</div>
		</div>
	{:else if type === 'matrix'}
		<div class="grid grid-cols-8 gap-1 font-mono">
			{#each Array(16) as _, i}
				<div class="h-2 w-2 animate-pulse">
					{matrixFrames[(frame + i) % matrixFrames.length]}
				</div>
			{/each}
		</div>
	{/if}

	{#if message}
		<div class="terminal-text text-sm text-cyber-text-secondary">
			{message}
		</div>
	{/if}
</div>
