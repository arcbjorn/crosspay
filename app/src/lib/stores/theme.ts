import { writable } from 'svelte/store';
import { browser } from '$app/environment';

type Theme = 'light' | 'dark' | 'auto';

// Get initial theme from localStorage or default to auto
const getInitialTheme = (): Theme => {
	if (!browser) return 'auto';
	return (localStorage.getItem('theme') as Theme) || 'auto';
};

export const theme = writable<Theme>(getInitialTheme());

// Apply theme to document
export function applyTheme(selectedTheme: Theme) {
	if (!browser) return;

	let actualTheme: 'light' | 'dark';

	if (selectedTheme === 'auto') {
		actualTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
	} else {
		actualTheme = selectedTheme;
	}

	document.documentElement.setAttribute('data-theme', actualTheme);
	localStorage.setItem('theme', selectedTheme);
}

// Subscribe to theme changes
theme.subscribe(applyTheme);

// Listen for system theme changes when in auto mode
if (browser) {
	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
		theme.update((currentTheme) => {
			if (currentTheme === 'auto') {
				applyTheme('auto');
			}
			return currentTheme;
		});
	});
}
