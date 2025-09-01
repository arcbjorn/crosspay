<script lang="ts">
	import CyberCard from './CyberCard.svelte';
	import CyberInput from './CyberInput.svelte';
	import CyberButton from './CyberButton.svelte';
	import { createEventDispatcher } from 'svelte';

	export let loading = false;
	export let networks: Array<{ id: string; name: string; symbol: string }> = [];

	const dispatch = createEventDispatcher();

	let amount = '';
	let recipient = '';
	let selectedNetwork = '';
	let memo = '';
	let isPrivate = false;

	function handleSubmit() {
		if (!amount || !recipient || !selectedNetwork) return;

		dispatch('submit', {
			amount: parseFloat(amount),
			recipient,
			network: selectedNetwork,
			memo,
			isPrivate
		});
	}

	function formatAmount(value: string) {
		return value.replace(/[^\d.]/g, '');
	}

	$: isValid = amount && recipient && selectedNetwork;
</script>

<CyberCard variant="mint" padding="lg" matrix>
	<div class="terminal-text mb-6 text-xl text-cyber-mint">[PAYMENT_TERMINAL_v2.1.0]</div>

	<div class="grid gap-6">
		<!-- Network Selection -->
		<div>
			<label class="terminal-text mb-3 block text-sm text-cyber-text-secondary">
				> SELECT_NETWORK:
			</label>
			<div class="grid grid-cols-2 gap-2 md:grid-cols-3">
				{#each networks as network}
					<button
						class="cyber-btn-secondary p-2 text-xs {selectedNetwork === network.id
							? 'bg-cyber-lavender text-cyber-bg-primary'
							: ''}"
						on:click={() => (selectedNetwork = network.id)}
					>
						{network.name}
						<span class="text-cyber-text-tertiary">({network.symbol})</span>
					</button>
				{/each}
			</div>
		</div>

		<!-- Recipient Address -->
		<CyberInput
			bind:value={recipient}
			label="RECIPIENT_ADDRESS"
			placeholder="0x... or ENS name"
			type="text"
		>
			<span slot="help"> > ENS names will be resolved automatically </span>
		</CyberInput>

		<!-- Amount -->
		<CyberInput
			bind:value={amount}
			label="AMOUNT"
			placeholder="0.00"
			type="text"
			on:input={(e) => (amount = formatAmount((e.target as HTMLInputElement).value))}
		>
			<span slot="help">
				> Enter payment amount in {selectedNetwork
					? networks.find((n) => n.id === selectedNetwork)?.symbol
					: 'tokens'}
			</span>
		</CyberInput>

		<!-- Privacy Toggle -->
		<div class="cyber-card bg-cyber-surface-2 p-4">
			<div class="mb-2 flex items-center justify-between">
				<span class="terminal-text text-sm text-cyber-text-primary"> [PRIVACY_MODE] </span>
				<button
					class="relative h-6 w-12 border border-cyber-border-mint bg-transparent transition-all"
					class:bg-cyber-mint={isPrivate}
					on:click={() => (isPrivate = !isPrivate)}
					aria-label="Toggle privacy mode"
				>
					<div
						class="absolute top-0.5 h-4 w-4 bg-cyber-text-primary transition-transform"
						class:translate-x-6={isPrivate}
						class:translate-x-0.5={!isPrivate}
					></div>
				</button>
			</div>
			<div class="terminal-text text-xs text-cyber-text-tertiary">
				{#if isPrivate}
					> AMOUNT_ENCRYPTED: Zama FHE active
				{:else}
					> AMOUNT_PUBLIC: Standard transaction
				{/if}
			</div>
		</div>

		<!-- Optional Memo -->
		<CyberInput
			bind:value={memo}
			label="MEMO (OPTIONAL)"
			placeholder="Payment reference..."
			type="text"
		/>

		<!-- Action Buttons -->
		<div class="mt-6 flex gap-4">
			<CyberButton
				variant="primary"
				size="lg"
				disabled={!isValid || loading}
				{loading}
				on:click={handleSubmit}
				class="flex-1"
			>
				{#if loading}
					PROCESSING...
				{:else}
					[EXECUTE_PAYMENT]
				{/if}
			</CyberButton>

			<CyberButton variant="secondary" size="lg" on:click={() => dispatch('clear')}>
				[CLEAR]
			</CyberButton>
		</div>
	</div>

	<!-- Status Display -->
	<div
		class="terminal-text mt-6 border-t border-cyber-border-mint/20 pt-4 text-xs text-cyber-text-tertiary"
	>
		<div class="grid grid-cols-2 gap-4">
			<div>STATUS: {loading ? 'PROCESSING' : 'READY'}</div>
			<div>PRIVACY: {isPrivate ? 'ENABLED' : 'DISABLED'}</div>
			<div>NETWORK: {selectedNetwork || 'NOT_SELECTED'}</div>
			<div>VALIDATION: {isValid ? 'PASSED' : 'PENDING'}</div>
		</div>
	</div>
</CyberCard>
