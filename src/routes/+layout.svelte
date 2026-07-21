<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import OfflineIndicator from '$lib/components/OfflineIndicator.svelte';
	import PwaInstallPrompt from '$lib/components/PwaInstallPrompt.svelte';
	import PwaUpdatePrompt from '$lib/components/PwaUpdatePrompt.svelte';
	import { startConnectionMonitor } from '$lib/services/connection';
	import { startSyncEngine, processSyncQueue } from '$lib/services/sync';
	import { refreshLocalCache } from '$lib/services/offline';

	let { children } = $props();

	onMount(() => {
		startConnectionMonitor();
		startSyncEngine();
		refreshLocalCache();
	});
</script>

<svelte:head>
	<title>Maestro POS</title>
	<meta name="description" content="Modern Point of Sale System — Manage sales, inventory, customers and more." />
</svelte:head>

<OfflineIndicator />
<PwaInstallPrompt />
<PwaUpdatePrompt />

{@render children()}
