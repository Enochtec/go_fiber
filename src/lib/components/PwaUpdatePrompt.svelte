<script lang="ts">
	import { onMount } from 'svelte';
	import { RefreshCw, X } from '@lucide/svelte';

	let needRefresh = $state(false);
	let registration: ServiceWorkerRegistration | null = null;
	let dismissed = $state(false);

	onMount(() => {
		if (typeof window === 'undefined' || !('serviceWorker' in navigator)) return;

		navigator.serviceWorker.register('/sw.js').then((reg) => {
			registration = reg;

			reg.addEventListener('updatefound', () => {
				const newSW = reg.installing;
				if (!newSW) return;
				newSW.addEventListener('statechange', () => {
					if (newSW.state === 'installed' && navigator.serviceWorker.controller) {
						needRefresh = true;
					}
				});
			});
		});

		navigator.serviceWorker.addEventListener('controllerchange', () => {
			window.location.reload();
		});
	});

	async function update() {
		if (!registration || !registration.waiting) return;
		registration.waiting.postMessage({ type: 'SKIP_WAITING' });
		needRefresh = false;
	}

	function dismiss() {
		dismissed = true;
	}
</script>

{#if needRefresh && !dismissed}
	<div class="fixed top-4 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 px-4 py-3 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-lg max-sm:left-4 max-sm:right-4 max-sm:-translate-x-0">
		<p class="text-xs text-slate-600 dark:text-slate-400">A new version of Maestro POS is available.</p>
		<div class="flex items-center gap-2 shrink-0">
			<button onclick={update} class="flex items-center gap-1 px-3 py-1.5 text-xs font-bold text-white transition-colors" style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
				<RefreshCw size={12} />
				Update
			</button>
			<button onclick={dismiss} class="text-xs font-medium text-slate-500 hover:text-slate-700 transition-colors">Later</button>
			<button onclick={dismiss} class="h-7 w-7 flex items-center justify-center text-slate-400 hover:text-slate-600 transition-colors">
				<X size={14} />
			</button>
		</div>
	</div>
{/if}
