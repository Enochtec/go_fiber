import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		SvelteKitPWA({
			registerType: 'prompt',
			includeAssets: [
				'favicon.svg',
				'apple-touch-icon.svg',
				'pinned-tab.svg',
				'icons/*.svg'
			],
			manifest: {
				name: 'Maestro POS',
				short_name: 'Maestro POS',
				description: 'Modern Point of Sale System — Manage sales, inventory, customers and more.',
				theme_color: '#3F00FF',
				background_color: '#0f172a',
				display: 'standalone',
				display_override: ['window-controls-overlay', 'standalone'],
				orientation: 'portrait-primary',
				start_url: '/',
				scope: '/',
				id: '/',
				categories: ['business', 'finance'],
				lang: 'en',
				dir: 'ltr',
				icons: [
					{ src: '/icons/icon.svg', sizes: '48x48', type: 'image/svg+xml' },
					{ src: '/icons/icon.svg', sizes: '72x72', type: 'image/svg+xml' },
					{ src: '/icons/icon.svg', sizes: '96x96', type: 'image/svg+xml' },
					{ src: '/icons/icon.svg', sizes: '128x128', type: 'image/svg+xml' },
					{ src: '/icons/icon.svg', sizes: '192x192', type: 'image/svg+xml' },
					{ src: '/icons/icon.svg', sizes: '256x256', type: 'image/svg+xml' },
					{ src: '/icons/icon.svg', sizes: '384x384', type: 'image/svg+xml' },
					{ src: '/icons/icon.svg', sizes: '512x512', type: 'image/svg+xml' },
					{ src: '/icons/icon-maskable.svg', sizes: '192x192', type: 'image/svg+xml', purpose: 'maskable' },
					{ src: '/icons/icon-maskable.svg', sizes: '512x512', type: 'image/svg+xml', purpose: 'maskable' },
				],
				shortcuts: [
					{
						name: 'New Sale',
						short_name: 'Sale',
						description: 'Start a new point of sale transaction',
						url: '/sales',
						icons: [{ src: '/icons/icon.svg', sizes: '96x96', type: 'image/svg+xml' }]
					},
					{
						name: 'Dashboard',
						short_name: 'Dashboard',
						description: 'View business overview and analytics',
						url: '/dashboard',
						icons: [{ src: '/icons/icon.svg', sizes: '96x96', type: 'image/svg+xml' }]
					},
					{
						name: 'Products',
						short_name: 'Products',
						description: 'Manage product catalog',
						url: '/products',
						icons: [{ src: '/icons/icon.svg', sizes: '96x96', type: 'image/svg+xml' }]
					},
					{
						name: 'Inventory',
						short_name: 'Stock',
						description: 'Check inventory levels and adjustments',
						url: '/inventory',
						icons: [{ src: '/icons/icon.svg', sizes: '96x96', type: 'image/svg+xml' }]
					},
				],
				screenshots: [],
			},
			workbox: {
				globPatterns: ['**/*.{js,css,html,svg,png,ico,woff,woff2}'],
				runtimeCaching: [
					{
						urlPattern: /^\/api\/.*/i,
						handler: 'NetworkFirst',
						method: 'GET',
						options: {
							cacheName: 'api-cache',
							expiration: {
								maxEntries: 100,
								maxAgeSeconds: 60 * 60,
							},
							cacheableResponse: {
								statuses: [0, 200],
							},
						},
					},
				],
			},
		}),
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
