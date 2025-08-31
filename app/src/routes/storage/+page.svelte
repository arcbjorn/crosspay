<script lang="ts">
  import { onMount } from 'svelte';

  let uploadedReceipts: any[] = [];
  let isLoading = false;
  let dragOver = false;

  onMount(() => {
    // Load stored receipts from local storage for demo
    const stored = localStorage.getItem('storedReceipts');
    if (stored) {
      uploadedReceipts = JSON.parse(stored);
    }
  });

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
    
    for (const file of files) {
      // Mock upload - in real implementation would upload to Filecoin via storage service
      const mockCID = `Qm${Math.random().toString(36).substring(2, 15)}`;
      
      const receipt = {
        id: Date.now() + Math.random(),
        name: file.name,
        size: file.size,
        type: file.type,
        cid: mockCID,
        uploadedAt: new Date().toISOString(),
        status: 'stored'
      };

      uploadedReceipts = [receipt, ...uploadedReceipts];
    }

    // Save to local storage for demo
    localStorage.setItem('storedReceipts', JSON.stringify(uploadedReceipts));
    isLoading = false;
  }

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  async function downloadFromIPFS(cid: string, filename: string) {
    // Mock download - would connect to IPFS gateway
    console.log(`Downloading ${filename} from IPFS: ${cid}`);
    // In real implementation:
    // const response = await fetch(`https://gateway.ipfs.io/ipfs/${cid}`);
    // const blob = await response.blob();
    // ... trigger download
  }
</script>

<svelte:head>
  <title>Decentralized Storage - CrossPay</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
  <div class="breadcrumbs text-sm mb-8">
    <ul>
      <li><a href="/">Home</a></li>
      <li>Storage</li>
    </ul>
  </div>

  <div class="mb-8">
    <h1 class="text-4xl font-bold mb-4">Decentralized Storage</h1>
    <p class="text-lg text-base-content/70">
      Store payment receipts and documents permanently on Filecoin with IPFS addressing.
    </p>
  </div>

  <!-- Upload Section -->
  <div class="card bg-base-100 shadow-xl mb-8">
    <div class="card-body">
      <h2 class="card-title mb-4">Upload to Filecoin</h2>
      
      <div 
        class="border-2 border-dashed border-base-300 rounded-lg p-8 text-center transition-colors"
        class:border-primary={dragOver}
        class:bg-primary/5={dragOver}
        on:dragover={handleDragOver}
        on:dragleave={handleDragLeave}
        on:drop={handleDrop}
      >
        {#if isLoading}
          <div class="loading loading-spinner loading-lg text-primary"></div>
          <p class="mt-4">Uploading to Filecoin...</p>
        {:else}
          <svg class="mx-auto h-12 w-12 text-base-content/50 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          
          <p class="text-lg font-medium mb-2">Drop files here or click to select</p>
          <p class="text-sm text-base-content/70 mb-4">
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
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>
        <span>Files are stored permanently on Filecoin network with content-addressable IPFS hashes.</span>
      </div>
    </div>
  </div>

  <!-- Stored Files -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title mb-4">Stored Files ({uploadedReceipts.length})</h2>
      
      {#if uploadedReceipts.length === 0}
        <div class="text-center text-base-content/70 py-12">
          <svg class="mx-auto h-16 w-16 text-base-content/30 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p class="text-lg">No files stored yet</p>
          <p class="text-sm">Upload your first receipt or document to get started.</p>
        </div>
      {:else}
        <div class="overflow-x-auto">
          <table class="table table-zebra">
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
                    <div class="text-sm text-base-content/70">{receipt.type}</div>
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
                      <button class="btn btn-ghost btn-xs">
                        Share
                      </button>
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
  <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mt-8">
    <div class="stat bg-base-100 shadow-xl rounded-lg">
      <div class="stat-title">Files Stored</div>
      <div class="stat-value text-primary">{uploadedReceipts.length}</div>
      <div class="stat-desc">Across Filecoin network</div>
    </div>
    
    <div class="stat bg-base-100 shadow-xl rounded-lg">
      <div class="stat-title">Total Size</div>
      <div class="stat-value text-secondary">
        {formatFileSize(uploadedReceipts.reduce((sum, r) => sum + r.size, 0))}
      </div>
      <div class="stat-desc">Permanently stored</div>
    </div>
    
    <div class="stat bg-base-100 shadow-xl rounded-lg">
      <div class="stat-title">Network</div>
      <div class="stat-value text-accent">Filecoin</div>
      <div class="stat-desc">Decentralized storage</div>
    </div>
  </div>
</div>