<script lang="ts">
  export let networks: Array<{
    id: string;
    name: string;
    status: 'online' | 'offline' | 'syncing';
    connections?: number;
    latency?: number;
  }> = [];
  export let showConnections = true;
  export let animated = true;
  export let size: 'sm' | 'md' | 'lg' = 'md';

  $: sizeConfig = {
    sm: { nodeSize: 'w-8 h-8', fontSize: 'text-xs', gridSize: 'gap-4' },
    md: { nodeSize: 'w-12 h-12', fontSize: 'text-sm', gridSize: 'gap-6' },
    lg: { nodeSize: 'w-16 h-16', fontSize: 'text-base', gridSize: 'gap-8' }
  };

  $: statusConfig = {
    online: { color: 'border-cyber-success bg-cyber-success/20', pulse: 'animate-pulse' },
    offline: { color: 'border-cyber-error bg-cyber-error/20', pulse: '' },
    syncing: { color: 'border-cyber-warning bg-cyber-warning/20', pulse: 'animate-pulse' }
  };

  function getConnectionLines() {
    if (!showConnections || networks.length < 2) return [];
    
    const lines = [];
    const onlineNetworks = networks.filter(n => n.status === 'online');
    
    for (let i = 0; i < onlineNetworks.length - 1; i++) {
      lines.push({
        from: i,
        to: i + 1,
        strength: Math.random() * 0.8 + 0.2
      });
    }
    
    return lines;
  }

  $: connectionLines = getConnectionLines();
</script>

<div class="network-grid relative font-mono">
  <!-- Network topology header -->
  <div class="terminal-text text-cyber-mint text-lg mb-6 flex items-center gap-2">
    <span class="text-cyber-text-tertiary">></span>
    NETWORK_TOPOLOGY
    <span class="animate-cursor-blink text-cyber-mint">|</span>
  </div>

  <!-- Network stats bar -->
  <div class="flex gap-6 mb-8 text-xs text-cyber-text-secondary">
    <span>NODES: {networks.length}</span>
    <span>ONLINE: {networks.filter(n => n.status === 'online').length}</span>
    <span>SYNCING: {networks.filter(n => n.status === 'syncing').length}</span>
    <span>OFFLINE: {networks.filter(n => n.status === 'offline').length}</span>
  </div>

  <!-- Network grid -->
  <div class="relative bg-cyber-surface-1 border border-cyber-border-mint p-8">
    <!-- Background grid pattern -->
    <div class="absolute inset-0 opacity-20" style="
      background-image: 
        linear-gradient(rgba(159, 239, 223, 0.1) 1px, transparent 1px),
        linear-gradient(90deg, rgba(159, 239, 223, 0.1) 1px, transparent 1px);
      background-size: 40px 40px;
    " />

    <!-- Connection lines -->
    {#if showConnections}
      <svg class="absolute inset-0 w-full h-full pointer-events-none z-10">
        {#each connectionLines as line}
          <line
            x1="25%"
            y1="25%"
            x2="75%"
            y2="75%"
            stroke="rgba(159, 239, 223, {line.strength})"
            stroke-width="1"
            stroke-dasharray="5,5"
            class:animate-pulse={animated}
          >
            <animate
              attributeName="stroke-dashoffset"
              values="0;10"
              dur="2s"
              repeatCount="indefinite"
            />
          </line>
        {/each}
      </svg>
    {/if}

    <!-- Network nodes -->
    <div class="relative z-20 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 {sizeConfig[size].gridSize} justify-items-center">
      {#each networks as network, i}
        <div class="flex flex-col items-center group">
          <!-- Node -->
          <div 
            class="relative {sizeConfig[size].nodeSize} {statusConfig[network.status].color} border-2 flex items-center justify-center cursor-pointer transition-all duration-300 hover:scale-110"
            class:animate-pulse={animated && statusConfig[network.status].pulse}
          >
            <!-- Node ID -->
            <span class="text-cyber-text-primary font-bold {sizeConfig[size].fontSize}">
              {network.id.slice(0, 2).toUpperCase()}
            </span>
            
            <!-- Status indicator -->
            <div class="absolute -top-1 -right-1 w-3 h-3 rounded-full {statusConfig[network.status].color} border border-cyber-bg-primary" />
            
            <!-- Scan line effect -->
            <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 animate-pulse" />
          </div>

          <!-- Node info -->
          <div class="mt-2 text-center {sizeConfig[size].fontSize}">
            <div class="text-cyber-text-primary font-medium truncate max-w-20">
              {network.name}
            </div>
            <div class="text-cyber-text-tertiary text-xs mt-1 space-y-0.5">
              {#if network.connections !== undefined}
                <div>C: {network.connections}</div>
              {/if}
              {#if network.latency !== undefined}
                <div>{network.latency}ms</div>
              {/if}
            </div>
          </div>

          <!-- Hover tooltip -->
          <div class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none z-30">
            <div class="bg-cyber-surface-2 border border-cyber-border-mint p-2 text-xs whitespace-nowrap">
              <div class="text-cyber-text-primary">{network.name}</div>
              <div class="text-cyber-text-secondary">Status: {network.status.toUpperCase()}</div>
              {#if network.connections}
                <div class="text-cyber-text-secondary">Connections: {network.connections}</div>
              {/if}
              {#if network.latency}
                <div class="text-cyber-text-secondary">Latency: {network.latency}ms</div>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    </div>

    <!-- Terminal corners -->
    <div class="absolute top-0 left-0 w-4 h-4 border-t border-l border-cyber-border-mint" />
    <div class="absolute top-0 right-0 w-4 h-4 border-t border-r border-cyber-border-mint" />
    <div class="absolute bottom-0 left-0 w-4 h-4 border-b border-l border-cyber-border-mint" />
    <div class="absolute bottom-0 right-0 w-4 h-4 border-b border-r border-cyber-border-mint" />
  </div>

  <!-- Legend -->
  <div class="mt-6 flex gap-6 text-xs text-cyber-text-tertiary">
    <div class="flex items-center gap-2">
      <div class="w-3 h-3 border-2 border-cyber-success bg-cyber-success/20" />
      <span>ONLINE</span>
    </div>
    <div class="flex items-center gap-2">
      <div class="w-3 h-3 border-2 border-cyber-warning bg-cyber-warning/20" />
      <span>SYNCING</span>
    </div>
    <div class="flex items-center gap-2">
      <div class="w-3 h-3 border-2 border-cyber-error bg-cyber-error/20" />
      <span>OFFLINE</span>
    </div>
  </div>
</div>