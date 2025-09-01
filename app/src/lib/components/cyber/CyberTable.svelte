<script lang="ts">
  export let data: Array<Record<string, any>> = [];
  export let columns: Array<{
    key: string;
    label: string;
    align?: 'left' | 'center' | 'right';
    width?: string;
    format?: (value: any) => string;
  }> = [];
  export let loading = false;
  export let emptyMessage = 'NO_DATA_FOUND';
  
  $: hasData = data && data.length > 0;
</script>

<div class="cyber-table-container relative overflow-hidden">
  <table class="cyber-table">
    <thead>
      <tr>
        {#each columns as column}
          <th 
            class="text-{column.align || 'left'}" 
            style={column.width ? `width: ${column.width}` : ''}
          >
            [{column.label.toUpperCase()}]
          </th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#if loading}
        <tr>
          <td colspan={columns.length} class="text-center py-8">
            <div class="ascii-loader text-cyber-mint">
              <div class="animate-pulse">
                LOADING_DATA...
              </div>
            </div>
          </td>
        </tr>
      {:else if hasData}
        {#each data as row, index}
          <tr class="hover:bg-cyber-surface-1/20 transition-colors data-stream">
            {#each columns as column}
              <td class="text-{column.align || 'left'}">
                {#if column.format}
                  {column.format(row[column.key])}
                {:else}
                  {row[column.key] || '--'}
                {/if}
              </td>
            {/each}
          </tr>
        {/each}
      {:else}
        <tr>
          <td colspan={columns.length} class="text-center py-8">
            <div class="terminal-text text-cyber-text-tertiary">
              <div class="text-lg mb-2">
                ┌─────────────┐<br>
                │   EMPTY     │<br>
                │ ░░░░░░░░░░░ │<br>
                │ ░░░░░░░░░░░ │<br>
                └─────────────┘
              </div>
              <div class="text-sm">
                {emptyMessage}
              </div>
            </div>
          </td>
        </tr>
      {/if}
    </tbody>
  </table>
  
  {#if hasData && !loading}
    <div class="terminal-text text-cyber-text-tertiary text-xs mt-2 text-right">
      RECORDS_TOTAL: {data.length}
    </div>
  {/if}
</div>

<style>
  .cyber-table-container {
    border: 1px solid rgba(159, 239, 223, 0.3);
    background: rgba(28, 28, 31, 0.5);
  }
  
  .cyber-table-container::before {
    content: '';
    position: absolute;
    inset: 0;
    background: 
      repeating-linear-gradient(
        0deg, 
        transparent, 
        transparent 2px, 
        rgba(159, 239, 223, 0.02) 2px, 
        rgba(159, 239, 223, 0.02) 4px
      );
    pointer-events: none;
  }
</style>