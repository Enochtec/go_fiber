<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { WifiOff } from '@lucide/svelte';

	let offline = $state(false);

	function update() {
		offline = !navigator.onLine;
	}

	onMount(() => {
		update();
		window.addEventListener('online', update);
		window.addEventListener('offline', update);
	});

	onDestroy(() => {
		window.removeEventListener('online', update);
		window.removeEventListener('offline', update);
	});
</script>

{#if offline}
	<div class="fixed top-0 left-0 right-0 z-[9999] flex items-center justify-center gap-2 bg-amber-500 px-4 py-2 text-xs font-semibold text-white">
		<WifiOff size={14} />
		<span>You are offline. Data will sync when connection resumes.</span>
	</div>
{/if}
