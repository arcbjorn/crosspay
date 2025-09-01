<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	export let title = 'CrossPay Protocol';
	export let showNav = true;
	export let fullscreen = false;

	let commandPaletteOpen = false;
	let currentPath = '/';

	const routes = [
		{ path: '/', label: 'HOME', shortcut: 'H' },
		{ path: '/pay', label: 'PAYMENT', shortcut: 'P' },
		{ path: '/receipts', label: 'RECEIPTS', shortcut: 'R' },
		{ path: '/security', label: 'SECURITY', shortcut: 'S' },
		{ path: '/ai', label: 'AI_COPILOT', shortcut: 'A' }
	];

	onMount(() => {
		function handleKeydown(event: KeyboardEvent) {
			// Command palette (CMD/CTRL + K)
			if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
				event.preventDefault();
				commandPaletteOpen = !commandPaletteOpen;
			}

			// Quick navigation shortcuts
			if ((event.metaKey || event.ctrlKey) && event.shiftKey) {
				const route = routes.find((r) => r.shortcut.toLowerCase() === event.key.toLowerCase());
				if (route) {
					event.preventDefault();
					window.location.href = route.path;
				}
			}

			// ESC to close command palette
			if (event.key === 'Escape') {
				commandPaletteOpen = false;
			}
		}

		document.addEventListener('keydown', handleKeydown);

		return () => {
			document.removeEventListener('keydown', handleKeydown);
		};
	});

	function navigateTo(path: string) {
		commandPaletteOpen = false;
		window.location.href = path;
	}

	$: {
		currentPath = $page.url.pathname;
	}
</script>

<svelte:head>
	<title>{title}</title>
</svelte:head>

<div class="cyber-layout {fullscreen ? 'fullscreen' : ''}" data-path={currentPath}>
	{#if showNav}
		<!-- Navigation Header -->
		<nav class="cyber-nav">
			<div class="nav-container">
				<!-- Logo -->
				<a href="/" class="nav-logo glitch-hover">
					<span class="terminal-text text-lg text-cyber-mint"> [CROSSPAY] </span>
				</a>

				<!-- Navigation Links -->
				<div class="nav-links">
					{#each routes as route}
						<a
							href={route.path}
							class="nav-link {currentPath === route.path ? 'active' : ''}"
							data-shortcut={route.shortcut}
						>
							{route.label}
							<span class="nav-shortcut">⌘⇧{route.shortcut}</span>
						</a>
					{/each}
				</div>

				<!-- Command Palette Trigger -->
				<button
					class="cyber-btn-secondary nav-command"
					on:click={() => (commandPaletteOpen = true)}
				>
					<span class="terminal-text">⌘K</span>
				</button>
			</div>
		</nav>
	{/if}

	<!-- Main Content -->
	<main class="cyber-main {showNav ? 'with-nav' : ''}">
		<slot />
	</main>

	<!-- Command Palette -->
	{#if commandPaletteOpen}
		<!-- svelte-ignore a11y-click-events-have-key-events -->
		<!-- svelte-ignore a11y-no-static-element-interactions -->
		<div
			class="command-palette-overlay"
			on:click={() => (commandPaletteOpen = false)}
			role="dialog"
			aria-modal="true"
			aria-labelledby="command-palette-title"
			tabindex="0"
		>
			<div class="command-palette" on:click|stopPropagation>
				<div class="command-header">
					<div id="command-palette-title" class="terminal-text mb-4 text-lg text-cyber-mint">
						[COMMAND_PALETTE]
					</div>
					<input type="text" placeholder="Type command..." class="cyber-input w-full" />
				</div>

				<div class="command-list">
					<div class="command-section">
						<div class="command-section-title">NAVIGATION:</div>
						{#each routes as route}
							<button class="command-item" on:click={() => navigateTo(route.path)}>
								<span class="command-label">{route.label}</span>
								<span class="command-shortcut">⌘⇧{route.shortcut}</span>
							</button>
						{/each}
					</div>

					<div class="command-section">
						<div class="command-section-title">ACTIONS:</div>
						<button class="command-item" on:click={() => navigateTo('/pay')}>
							<span class="command-label">NEW_PAYMENT</span>
							<span class="command-shortcut">⌘⇧N</span>
						</button>
						<button class="command-item" on:click={() => (commandPaletteOpen = false)}>
							<span class="command-label">CLOSE_PALETTE</span>
							<span class="command-shortcut">ESC</span>
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}

	<!-- Accessibility Announcements -->
	<div class="sr-only" aria-live="polite" id="announcements"></div>
</div>

<style>
	.cyber-layout {
		min-height: 100vh;
		background: var(--cyber-bg-primary);
		position: relative;
	}

	.cyber-layout.fullscreen {
		height: 100vh;
		overflow: hidden;
	}

	/* Navigation */
	.cyber-nav {
		border-bottom: 1px solid var(--cyber-border-mint);
		background: var(--cyber-surface-1);
		position: sticky;
		top: 0;
		z-index: 100;
	}

	.nav-container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 1rem;
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 60px;
	}

	.nav-logo {
		text-decoration: none;
		transition: all 0.2s;
	}

	.nav-links {
		display: flex;
		gap: 2rem;
	}

	.nav-link {
		text-decoration: none;
		color: var(--cyber-text-secondary);
		font-family: 'Courier New', monospace;
		font-size: 0.875rem;
		padding: 0.5rem 1rem;
		border: 1px solid transparent;
		transition: all 0.2s;
		position: relative;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.nav-link:hover {
		color: var(--cyber-mint);
		border-color: var(--cyber-border-mint);
	}

	.nav-link.active {
		color: var(--cyber-mint);
		border-color: var(--cyber-border-mint);
		background: rgba(159, 239, 223, 0.05);
	}

	.nav-shortcut {
		font-size: 0.75rem;
		color: var(--cyber-text-tertiary);
		opacity: 0;
		transition: opacity 0.2s;
	}

	.nav-link:hover .nav-shortcut {
		opacity: 1;
	}

	.nav-command {
		padding: 0.5rem;
		min-height: auto;
	}

	/* Main Content */
	.cyber-main {
		flex: 1;
		position: relative;
		z-index: 10;
	}

	.cyber-main.with-nav {
		min-height: calc(100vh - 60px);
	}

	/* Command Palette */
	.command-palette-overlay {
		position: fixed;
		inset: 0;
		background: rgba(10, 10, 11, 0.8);
		backdrop-filter: blur(4px);
		display: flex;
		align-items: flex-start;
		justify-content: center;
		padding-top: 20vh;
		z-index: 1000;
	}

	.command-palette {
		background: var(--cyber-surface-1);
		border: 1px solid var(--cyber-border-mint);
		width: 100%;
		max-width: 500px;
		margin: 0 1rem;
		position: relative;
	}

	.command-palette::before {
		content: '';
		position: absolute;
		inset: 0;
		background: repeating-linear-gradient(
			0deg,
			transparent,
			transparent 2px,
			rgba(159, 239, 223, 0.03) 2px,
			rgba(159, 239, 223, 0.03) 4px
		);
		opacity: 0.5;
		pointer-events: none;
	}

	.command-header {
		padding: 1.5rem;
		border-bottom: 1px solid var(--cyber-border-mint);
		position: relative;
		z-index: 1;
	}

	.command-list {
		max-height: 400px;
		overflow-y: auto;
		position: relative;
		z-index: 1;
	}

	.command-section {
		border-bottom: 1px solid var(--cyber-text-tertiary);
		border-bottom-width: 1px;
		border-bottom-style: solid;
		border-bottom-color: rgba(112, 112, 112, 0.2);
	}

	.command-section:last-child {
		border-bottom: none;
	}

	.command-section-title {
		padding: 1rem 1.5rem 0.5rem;
		color: var(--cyber-text-tertiary);
		font-family: 'Courier New', monospace;
		font-size: 0.75rem;
		font-weight: 600;
	}

	.command-item {
		width: 100%;
		padding: 0.75rem 1.5rem;
		background: transparent;
		border: none;
		color: var(--cyber-text-primary);
		text-align: left;
		cursor: pointer;
		transition: all 0.2s;
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-family: 'Courier New', monospace;
		font-size: 0.875rem;
	}

	.command-item:hover {
		background: rgba(159, 239, 223, 0.1);
		color: var(--cyber-mint);
	}

	.command-label {
		flex: 1;
	}

	.command-shortcut {
		color: var(--cyber-text-tertiary);
		font-size: 0.75rem;
	}

	/* Mobile Responsive */
	@media (max-width: 768px) {
		.nav-links {
			display: none;
		}

		.nav-container {
			padding: 0 0.5rem;
		}

		.command-palette-overlay {
			padding-top: 10vh;
		}
	}

	/* Screen Reader Only */
	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border: 0;
	}

	/* High Contrast Mode */
	@media (prefers-contrast: high) {
		.cyber-nav {
			border-bottom-width: 2px;
		}

		.nav-link {
			border-width: 2px;
		}

		.command-palette {
			border-width: 2px;
		}
	}

	/* Reduced Motion */
	@media (prefers-reduced-motion: reduce) {
		.nav-link,
		.command-item,
		.glitch-hover {
			transition: none;
		}
	}
</style>
