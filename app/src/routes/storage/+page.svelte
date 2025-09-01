<script lang="ts">
	import { onMount } from 'svelte';
	import { storageService } from '$lib/services/storage';
	import type { FileInfo } from '$lib/services/storage';
	import { successToast, errorToast } from '$lib/stores/toast';

	let uploadedReceipts: FileInfo[] = [];
	let isLoading = false;
	let dragOver = false;
	let serviceAvailable = false;

	onMount(async () => {
		// Check if storage service is available
		serviceAvailable = await storageService.isServiceAvailable();

		if (serviceAvailable) {
			await loadStoredFiles();
		} else {
			console.warn('Storage service not available, using local storage fallback');
			// Fallback to local storage for demo
			const stored = localStorage.getItem('storedReceipts');
			if (stored) {
				uploadedReceipts = JSON.parse(stored);
			}
		}
	});

	async function loadStoredFiles() {
		try {
			isLoading = true;
			const files = await storageService.listFiles();
			uploadedReceipts = files;
		} catch (error) {
			console.error('Failed to load stored files:', error);
			errorToast('Failed to load stored files');
		} finally {
			isLoading = false;
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		dragOver = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		dragOver = false;
	}

	async function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragOver = false;

		const files = Array.from(e.dataTransfer?.files || []);
		await uploadFiles(files);
	}

	async function handleFileSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		const files = Array.from(input.files || []);
		await uploadFiles(files);
	}

	async function uploadFiles(files: File[]) {
		isLoading = true;

		try {
			for (const file of files) {
				if (serviceAvailable) {
					// Real upload via SynapseSDK
					const result = await storageService.uploadFile(file);

					const fileInfo: FileInfo = {
						cid: result.cid,
						size: result.size,
						dealId: result.dealId || '',
						storageCost: result.cost,
						status: result.status || 'stored',
						metadata: {
							filename: file.name,
							contentType: file.type,
							uploadedAt: result.timestamp
						},
						createdAt: result.timestamp,
						name: file.name,
						type: file.type,
						uploadedAt: result.timestamp
					};

					uploadedReceipts = [fileInfo, ...uploadedReceipts];
					successToast(`File uploaded successfully: ${file.name}`);
				} else {
					// Fallback mock upload
					const mockCID = `Qm${Math.random().toString(36).substring(2, 15)}`;

					const receipt: FileInfo = {
						cid: mockCID,
						size: file.size,
						dealId: '',
						storageCost: '0.001 FIL',
						status: 'stored',
						metadata: {
							filename: file.name,
							contentType: file.type,
							uploadedAt: new Date().toISOString()
						},
						createdAt: new Date().toISOString(),
						name: file.name,
						type: file.type,
						uploadedAt: new Date().toISOString()
					};

					uploadedReceipts = [receipt, ...uploadedReceipts];
				}
			}

			if (!serviceAvailable) {
				// Save to local storage for demo fallback
				localStorage.setItem('storedReceipts', JSON.stringify(uploadedReceipts));
			}
		} catch (error) {
			console.error('Upload failed:', error);
			errorToast(`Upload failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
		} finally {
			isLoading = false;
		}
	}

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	async function downloadFromIPFS(cid: string, filename: string) {
		try {
			if (serviceAvailable) {
				// Real download via storage service
				await storageService.downloadFile(cid, filename);
				successToast(`Downloaded: ${filename}`);
			} else {
				// Mock download - would connect to IPFS gateway
				console.log(`Downloading ${filename} from IPFS: ${cid}`);
				// Open IPFS gateway URL as fallback
				const gatewayUrl = storageService.getIPFSGatewayUrl(cid);
				window.open(gatewayUrl, '_blank');
			}
		} catch (error) {
			console.error('Download failed:', error);
			errorToast(`Download failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
		}
	}
</script>

<svelte:head>
	<title>Decentralized Storage - CrossPay</title>
</svelte:head>

<div class="mx-auto max-w-6xl">
	<div class="breadcrumbs mb-8 text-sm">
		<ul>
			<li><a href="/">Home</a></li>
			<li>Storage</li>
		</ul>
	</div>

	<div class="mb-8">
		<h1 class="mb-4 text-4xl font-bold">Decentralized Storage</h1>
		<p class="text-base-content/70 text-lg">
			Store payment receipts and documents permanently on Filecoin with IPFS addressing.
		</p>
	</div>

	<!-- Upload Section -->
	<div class="card bg-base-100 mb-8 shadow-xl">
		<div class="card-body">
			<h2 class="card-title mb-4">Upload to Filecoin</h2>

			<div
				class="border-base-300 rounded-lg border-2 border-dashed p-8 text-center transition-colors"
				class:border-primary={dragOver}
				class:bg-primary-5={dragOver}
				on:dragover={handleDragOver}
				on:dragleave={handleDragLeave}
				role="button"
				tabindex="0"
				on:drop={handleDrop}
			>
				{#if isLoading}
					<div class="loading loading-spinner loading-lg text-primary"></div>
					<p class="mt-4">Uploading to Filecoin...</p>
				{:else}
					<svg
						class="text-base-content/50 mx-auto mb-4 h-12 w-12"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
						/>
					</svg>

					<p class="mb-2 text-lg font-medium">Drop files here or click to select</p>
					<p class="text-base-content/70 mb-4 text-sm">
						Supported formats: PDF, JSON, images, documents
					</p>

					<input
						type="file"
						multiple
						class="file-input file-input-primary w-full max-w-xs"
						on:change={handleFileSelect}
					/>
				{/if}
			</div>

			<div class="alert alert-info mt-4">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					class="h-6 w-6 shrink-0 stroke-current"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
					></path>
				</svg>
				<span
					>Files are stored permanently on Filecoin network with content-addressable IPFS hashes.</span
				>
			</div>
		</div>
	</div>

	<!-- Stored Files -->
	<div class="card bg-base-100 shadow-xl">
		<div class="card-body">
			<h2 class="card-title mb-4">Stored Files ({uploadedReceipts.length})</h2>

			{#if uploadedReceipts.length === 0}
				<div class="text-base-content/70 py-12 text-center">
					<svg
						class="text-base-content/30 mx-auto mb-4 h-16 w-16"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="1"
							d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
						/>
					</svg>
					<p class="text-lg">No files stored yet</p>
					<p class="text-sm">Upload your first receipt or document to get started.</p>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table-zebra table">
						<thead>
							<tr>
								<th>File Name</th>
								<th>Size</th>
								<th>IPFS CID</th>
								<th>Uploaded</th>
								<th>Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each uploadedReceipts as receipt}
								<tr>
									<td>
										<div class="font-medium">{receipt.name}</div>
										<div class="text-base-content/70 text-sm">{receipt.type}</div>
									</td>
									<td class="font-mono text-sm">
										{formatFileSize(receipt.size)}
									</td>
									<td class="font-mono text-xs">
										<div class="tooltip" data-tip={receipt.cid}>
											{receipt.cid.slice(0, 8)}...{receipt.cid.slice(-8)}
										</div>
									</td>
									<td class="text-sm">
										{new Date(receipt.uploadedAt).toLocaleDateString()}
									</td>
									<td>
										<div class="flex gap-2">
											<button
												class="btn btn-ghost btn-xs"
												on:click={() => downloadFromIPFS(receipt.cid, receipt.name)}
											>
												Download
											</button>
											<button class="btn btn-ghost btn-xs"> Share </button>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>

	<!-- Storage Stats -->
	<div class="mt-8 grid grid-cols-1 gap-6 md:grid-cols-3">
		<div class="stat bg-base-100 rounded-lg shadow-xl">
			<div class="stat-title">Files Stored</div>
			<div class="stat-value text-primary">{uploadedReceipts.length}</div>
			<div class="stat-desc">Across Filecoin network</div>
		</div>

		<div class="stat bg-base-100 rounded-lg shadow-xl">
			<div class="stat-title">Total Size</div>
			<div class="stat-value text-secondary">
				{formatFileSize(uploadedReceipts.reduce((sum, r) => sum + r.size, 0))}
			</div>
			<div class="stat-desc">Permanently stored</div>
		</div>

		<div class="stat bg-base-100 rounded-lg shadow-xl">
			<div class="stat-title">Network</div>
			<div class="stat-value text-accent">Filecoin</div>
			<div class="stat-desc">Decentralized storage</div>
		</div>
	</div>
</div>
