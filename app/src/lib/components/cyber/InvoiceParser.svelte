<script lang="ts">
	import CyberCard from './CyberCard.svelte';
	import CyberButton from './CyberButton.svelte';
	import CyberLoader from './CyberLoader.svelte';
	import {
		invoiceParserAI,
		type ParsingResult,
		type InvoiceData
	} from '$lib/services/ai/invoiceParser';
	import { createEventDispatcher } from 'svelte';

	const dispatch = createEventDispatcher();

	let files: FileList | null = null;
	let dragover = false;
	let parsing = false;
	let parsingStep = 0;
	let totalSteps = 4;
	let result: ParsingResult | null = null;
	let progressMessage = '';

	const stepMessages = [
		'INITIALIZING_OCR_ENGINE...',
		'EXTRACTING_TEXT_FROM_IMAGE...',
		'ANALYZING_DOCUMENT_STRUCTURE...',
		'PARSING_INVOICE_FIELDS...'
	];

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		dragover = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		dragover = false;
	}

	async function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragover = false;

		const droppedFiles = e.dataTransfer?.files;
		if (droppedFiles && droppedFiles.length > 0) {
			await processFile(droppedFiles[0]);
		}
	}

	async function handleFileSelect(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.files && target.files.length > 0) {
			await processFile(target.files[0]);
		}
	}

	async function processFile(file: File) {
		if (!file.type.startsWith('image/') && file.type !== 'application/pdf') {
			alert('Please select an image or PDF file');
			return;
		}

		parsing = true;
		parsingStep = 0;
		result = null;

		// Simulate processing steps
		for (let i = 0; i < totalSteps; i++) {
			parsingStep = i;
			progressMessage = stepMessages[i];
			await new Promise((resolve) => setTimeout(resolve, 800 + Math.random() * 400));
		}

		// Actually parse the invoice
		result = await invoiceParserAI.parseInvoice(file);
		parsing = false;

		if (result.success && result.data) {
			dispatch('parsed', result.data);
		}
	}

	function clearResults() {
		result = null;
		files = null;
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	$: progressPercentage = parsing ? Math.floor((parsingStep / totalSteps) * 100) : 0;
</script>

<CyberCard variant="lavender" padding="lg">
	<div class="terminal-text mb-6 text-xl text-cyber-lavender">[INVOICE_PARSER_AI]</div>

	{#if !result}
		<!-- File Upload Zone -->
		<div
			class="border-2 border-dashed p-8 text-center transition-all duration-300 {dragover
				? 'border-cyber-mint bg-cyber-mint/10'
				: 'border-cyber-border-lavender'}"
			on:dragover={handleDragOver}
			on:dragleave={handleDragLeave}
			on:drop={handleDrop}
			role="button"
			tabindex="0"
		>
			{#if parsing}
				<div class="space-y-4">
					<CyberLoader type="matrix" size="lg" />

					<!-- Progress Display -->
					<div class="terminal-text text-cyber-lavender">
						<div class="mb-2">PROCESSING_INVOICE:</div>
						<div class="mb-2 font-mono text-sm text-cyber-text-secondary">
							{progressMessage}
						</div>

						<!-- ASCII Progress Bar -->
						<div class="border border-cyber-border-lavender bg-cyber-bg-secondary p-2">
							<div class="font-mono text-xs">
								[{Array.from({ length: 20 }, (_, i) =>
									i < Math.floor(progressPercentage / 5) ? '█' : '░'
								).join('')}] {progressPercentage}%
							</div>
						</div>

						<div class="mt-2 text-xs text-cyber-text-tertiary">
							STEP {parsingStep + 1}/{totalSteps}
						</div>
					</div>
				</div>
			{:else}
				<div class="space-y-4">
					<!-- Upload Icon -->
					<div class="terminal-text text-4xl text-cyber-lavender">
						┌─────────────────┐<br />
						│ ┌─────────────┐ │<br />
						│ │ INVOICE │ │<br />
						│ │ ░░░░░░░░░░░ │ │<br />
						│ │ ░░░░░░░░░░░ │ │<br />
						│ │ ░░░░░░░░░░░ │ │<br />
						│ └─────────────┘ │<br />
						└─────────────────┘
					</div>

					<div class="terminal-text text-cyber-text-primary">DRAG_AND_DROP_INVOICE</div>

					<div class="text-sm text-cyber-text-secondary">
						> Supported formats: PNG, JPG, PDF<br />
						> Maximum file size: 10MB<br />
						> OCR accuracy: ~95%
					</div>

					<div class="flex justify-center gap-4">
						<CyberButton
							variant="secondary"
							on:click={() => document.getElementById('file-input')?.click()}
						>
							[BROWSE_FILES]
						</CyberButton>
					</div>

					<input
						id="file-input"
						type="file"
						accept="image/*,.pdf"
						class="hidden"
						on:change={handleFileSelect}
					/>
				</div>
			{/if}
		</div>
	{:else}
		<!-- Results Display -->
		<div class="space-y-6">
			<!-- Status Header -->
			<div class="flex items-center justify-between">
				<div class="terminal-text {result.success ? 'text-cyber-success' : 'text-cyber-error'}">
					[{result.success ? 'PARSING_SUCCESS' : 'PARSING_FAILED'}]
				</div>
				<div class="terminal-text text-sm text-cyber-text-tertiary">
					{result.processingTime}ms | {Math.floor(result.confidence * 100)}% confidence
				</div>
			</div>

			{#if result.success && result.data}
				<!-- Parsed Invoice Data -->
				<div class="cyber-card bg-cyber-surface-2 p-4">
					<div class="terminal-text mb-4 text-cyber-mint">EXTRACTED_DATA:</div>

					<div class="grid gap-6 md:grid-cols-2">
						<!-- Header Info -->
						<div class="space-y-2">
							<div class="terminal-text text-sm text-cyber-text-secondary">INVOICE_DETAILS:</div>
							<div class="space-y-1 text-xs">
								<div class="flex justify-between">
									<span class="text-cyber-text-tertiary">NUMBER:</span>
									<span class="font-mono text-cyber-text-primary">
										{result.data.invoiceNumber || 'N/A'}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-cyber-text-tertiary">DATE:</span>
									<span class="font-mono text-cyber-text-primary">
										{result.data.date || 'N/A'}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-cyber-text-tertiary">DUE:</span>
									<span class="font-mono text-cyber-text-primary">
										{result.data.dueDate || 'N/A'}
									</span>
								</div>
							</div>
						</div>

						<!-- Vendor Info -->
						<div class="space-y-2">
							<div class="terminal-text text-sm text-cyber-text-secondary">VENDOR:</div>
							<div class="text-xs">
								<div class="font-mono text-cyber-text-primary">
									{result.data.vendor?.name || 'Unknown'}
								</div>
								{#if result.data.vendor?.address}
									<div class="mt-1 text-cyber-text-tertiary">
										{result.data.vendor.address}
									</div>
								{/if}
							</div>
						</div>
					</div>

					<!-- Line Items -->
					{#if result.data.items.length > 0}
						<div class="mt-6">
							<div class="terminal-text mb-3 text-sm text-cyber-text-secondary">LINE_ITEMS:</div>

							<div class="cyber-table-container text-xs">
								<table class="w-full">
									<thead>
										<tr class="border-b border-cyber-border-lavender">
											<th class="px-3 py-2 text-left text-cyber-text-primary">DESCRIPTION</th>
											<th class="px-3 py-2 text-right text-cyber-text-primary">QTY</th>
											<th class="px-3 py-2 text-right text-cyber-text-primary">RATE</th>
											<th class="px-3 py-2 text-right text-cyber-text-primary">TOTAL</th>
										</tr>
									</thead>
									<tbody>
										{#each result.data.items as item}
											<tr class="border-b border-cyber-text-tertiary/20">
												<td class="px-3 py-2 text-cyber-text-secondary">{item.description}</td>
												<td class="px-3 py-2 text-right font-mono text-cyber-text-secondary"
													>{item.quantity}</td
												>
												<td class="px-3 py-2 text-right font-mono text-cyber-text-secondary">
													{formatCurrency(item.unitPrice)}
												</td>
												<td class="px-3 py-2 text-right font-mono text-cyber-text-primary">
													{formatCurrency(item.total)}
												</td>
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						</div>
					{/if}

					<!-- Totals -->
					<div class="mt-6 flex justify-end">
						<div class="min-w-48 space-y-1 text-sm">
							{#if result.data.subtotal}
								<div class="flex justify-between">
									<span class="text-cyber-text-tertiary">SUBTOTAL:</span>
									<span class="font-mono text-cyber-text-primary"
										>{formatCurrency(result.data.subtotal)}</span
									>
								</div>
							{/if}
							{#if result.data.tax}
								<div class="flex justify-between">
									<span class="text-cyber-text-tertiary">TAX:</span>
									<span class="font-mono text-cyber-text-primary"
										>{formatCurrency(result.data.tax)}</span
									>
								</div>
							{/if}
							<div class="flex justify-between border-t border-cyber-border-lavender pt-1">
								<span class="font-bold text-cyber-lavender">TOTAL:</span>
								<span class="font-mono font-bold text-cyber-lavender"
									>{formatCurrency(result.data.total)}</span
								>
							</div>
						</div>
					</div>
				</div>
			{/if}

			<!-- Errors -->
			{#if result.errors.length > 0}
				<div class="cyber-card border-cyber-error p-4">
					<div class="terminal-text mb-2 text-cyber-error">PARSING_ERRORS:</div>
					{#each result.errors as error}
						<div class="text-xs text-cyber-text-tertiary">
							> {error}
						</div>
					{/each}
				</div>
			{/if}

			<!-- Actions -->
			<div class="flex gap-4">
				<CyberButton variant="primary" on:click={() => result && dispatch('use', result.data)}>
					[USE_DATA]
				</CyberButton>
				<CyberButton variant="secondary" on:click={clearResults}>[PARSE_ANOTHER]</CyberButton>
			</div>
		</div>
	{/if}
</CyberCard>
