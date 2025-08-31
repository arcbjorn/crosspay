import adapter from '@sveltejs/adapter-auto';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),
	kit: { 
		adapter: adapter(),
		alias: {
			'@': 'src',
			'@components': 'src/lib/components',
			'@services': 'src/lib/services',
			'@stores': 'src/lib/stores',
			'@config': 'src/lib/config',
			'@utils': 'src/lib/utils',
			'@types': 'src/lib/types',
			'@contracts': 'src/lib/contracts',
			'@routes': 'src/routes',
			'@packages': '../packages'
		}
	}
};

export default config;
