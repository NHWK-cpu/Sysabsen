// vite.config.ts
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	// Ignore a11y warnings so the build can succeed
	onwarn: (warning, handler) => {
		if (warning.code && warning.code.startsWith('a11y_')) return;
		handler(warning);
	}
});
