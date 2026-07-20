import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit()
	],
	build: {
		target: 'es2020',
		cssMinify: 'esbuild',
		rollupOptions: {
			output: {
				manualChunks(id: string) {
					if (id.includes('node_modules/.pnpm') || id.includes('node_modules/')) {
						if (id.includes('@lucide/svelte')) return 'lucide';
						if (id.includes('svelte')) return 'vendor';
					}
				}
			}
		}
	},
	server: {
		proxy: {
			'/api': {
				target: 'http://localhost:8080',
				changeOrigin: true
			}
		}
	}
});
