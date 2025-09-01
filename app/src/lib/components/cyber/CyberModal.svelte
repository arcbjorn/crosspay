<script lang="ts">
	export let open = false;
	export let title: string | undefined = undefined;
	export let closable = true;

	import { createEventDispatcher } from 'svelte';
	import CyberButton from './CyberButton.svelte';

	const dispatch = createEventDispatcher();

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget && closable) {
			close();
		}
	}

	function close() {
		open = false;
		dispatch('close');
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && closable) {
			close();
		}
	}
</script>

{#if open}
	<!-- svelte-ignore a11y-no-static-element-interactions -->
	<div
		class="cyber-modal"
		on:click={handleBackdropClick}
		on:keydown={handleKeydown}
		tabindex="-1"
		role="dialog"
		aria-modal="true"
	>
		<div class="cyber-modal-content" on:click|stopPropagation on:keydown={handleKeydown}>
			{#if closable}
				<button
					class="absolute right-4 top-4 text-cyber-text-tertiary transition-colors hover:text-cyber-mint"
					on:click={close}
					aria-label="Close modal"
				>
					<span class="terminal-text text-lg">×</span>
				</button>
			{/if}

			{#if title}
				<header class="mb-6">
					<h2 class="terminal-text text-xl text-cyber-mint">
						[{title.toUpperCase()}]
					</h2>
				</header>
			{/if}

			<main class="relative z-10">
				<slot />
			</main>

			{#if $$slots.footer}
				<footer class="mt-6 flex justify-end gap-3">
					<slot name="footer" />
				</footer>
			{/if}
		</div>
	</div>
{/if}
