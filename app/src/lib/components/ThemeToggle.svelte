<script lang="ts">
	import { theme } from '$lib/stores/theme';

	$: currentTheme = $theme;

	const themes = [
		{ value: 'light', label: 'Light', icon: '☀️' },
		{ value: 'dark', label: 'Dark', icon: '🌙' },
		{ value: 'auto', label: 'Auto', icon: '🔄' }
	] as const;

	function setTheme(newTheme: typeof currentTheme) {
		theme.set(newTheme);
	}
</script>

<div class="dropdown dropdown-end">
	<div tabindex="0" role="button" class="btn btn-ghost btn-circle">
		<span class="text-lg">
			{#if currentTheme === 'light'}☀️{:else if currentTheme === 'dark'}🌙{:else}🔄{/if}
		</span>
	</div>
	<ul role="menu" class="dropdown-content menu bg-base-100 rounded-box w-40 p-2 shadow">
		<li class="menu-title">Theme</li>
		{#each themes as themeOption}
			<li>
				<button
					class="flex items-center gap-2"
					class:active={currentTheme === themeOption.value}
					on:click={() => setTheme(themeOption.value)}
				>
					<span>{themeOption.icon}</span>
					<span>{themeOption.label}</span>
				</button>
			</li>
		{/each}
	</ul>
</div>
