<script lang="ts">
	import { Download, X } from '@lucide/svelte';
	import { onMount } from 'svelte';

	let deferredPrompt = $state<Event | null>(null);
	let dismissed = $state(false);

	onMount(() => {
		if (typeof window === 'undefined') return;
		const handler = (e: Event) => {
			e.preventDefault();
			deferredPrompt = e;
		};
		window.addEventListener('beforeinstallprompt', handler);
		return () => window.removeEventListener('beforeinstallprompt', handler);
	});

	async function install() {
		if (!deferredPrompt) return;
		const prompt = deferredPrompt as unknown as { prompt: () => Promise<void>; userChoice: Promise<{ outcome: string }> };
		prompt.prompt();
		const { outcome } = await prompt.userChoice;
		if (outcome === 'accepted') {
			deferredPrompt = null;
		}
	}

	function dismiss() {
		dismissed = true;
	}
</script>

{#if deferredPrompt && !dismissed}
	<div class="fixed bottom-4 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 px-4 py-3 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-lg max-sm:left-4 max-sm:right-4 max-sm:-translate-x-0 max-sm:bottom-[4.5rem]">
		<button onclick={install} class="flex items-center gap-1.5 px-4 py-2 text-xs font-bold text-white transition-all active:scale-[.97] shrink-0" style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
			<Download size={14} />
			Install App
		</button>
		<span class="text-xs text-slate-600 dark:text-slate-400">Install Maestro POS for the best experience</span>
		<button onclick={dismiss} class="h-7 w-7 flex items-center justify-center text-slate-400 hover:text-slate-600 transition-colors shrink-0">
			<X size={14} />
		</button>
	</div>
{/if}
