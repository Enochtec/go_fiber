<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { WifiOff } from '@lucide/svelte';
	import { connectionStatus } from '$lib/services/connection';

	let offline = $state(false);

	let unsubscribe: (() => void) | undefined;

	onMount(() => {
		unsubscribe = connectionStatus.subscribe((status) => offline = status === 'offline');
	});

	onDestroy(() => {
		unsubscribe?.();
	});
</script>

{#if offline}
	<div class="fixed top-0 left-0 right-0 z-[9999] flex items-center justify-center gap-2 bg-amber-500 px-4 py-2 text-xs font-semibold text-white">
		<WifiOff size={14} />
		<span>You are offline. Data will sync when connection resumes.</span>
	</div>
{/if}
