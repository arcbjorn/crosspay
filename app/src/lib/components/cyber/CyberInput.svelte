<script lang="ts">
	export let value = '';
	export let type = 'text';
	export let placeholder = '';
	export let disabled = false;
	export let error = false;
	export let label: string | undefined = undefined;
	export let terminal = true;

	let inputElement: HTMLInputElement;
	let focused = false;

	$: inputClasses = [
		'cyber-input',
		'w-full',
		error ? 'border-cyber-error focus:border-cyber-error' : '',
		disabled ? 'opacity-50 cursor-not-allowed' : '',
		terminal ? 'terminal-cursor' : ''
	]
		.filter(Boolean)
		.join(' ');
</script>

<div class="cyber-input-container">
	{#if label}
		<label
			for={`cyber-input-${Date.now()}`}
			class="terminal-text mb-2 block text-sm text-cyber-text-secondary"
		>
			> {label}
		</label>
	{/if}

	<div class="relative">
		<input
			bind:this={inputElement}
			bind:value
			{type}
			{placeholder}
			{disabled}
			id={`cyber-input-${Date.now()}`}
			class={inputClasses}
			on:focus={() => (focused = true)}
			on:blur={() => (focused = false)}
			on:input
			on:change
			on:keydown
			{...$$restProps}
		/>

		{#if terminal && focused}
			<div
				class="absolute right-2 top-1/2 -translate-y-1/2 transform animate-cursor-blink text-cyber-mint"
			>
				|
			</div>
		{/if}
	</div>

	{#if $$slots.help}
		<div class="terminal-text mt-1 text-xs text-cyber-text-tertiary">
			<slot name="help" />
		</div>
	{/if}

	{#if error && $$slots.error}
		<div class="terminal-text mt-1 text-xs text-cyber-error">
			ERROR: <slot name="error" />
		</div>
	{/if}
</div>

<style>
	.cyber-input-container .cyber-input:focus {
		animation: terminal-glow 2s ease-in-out infinite alternate;
	}
</style>
