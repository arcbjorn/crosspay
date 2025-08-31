import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
  plugins: [sveltekit()],
  test: {
    include: ['src/**/*.{test,spec}.{js,ts}'],
    globals: true,
    environment: 'jsdom',
    setupFiles: ['src/setupTests.ts'],
  },
  define: {
    // Eliminate in-source test code in production builds
    'import.meta.vitest': 'undefined',
  },
});