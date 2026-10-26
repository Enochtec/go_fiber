<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import OfflineIndicator from '$lib/components/OfflineIndicator.svelte';
	import PwaInstallPrompt from '$lib/components/PwaInstallPrompt.svelte';
	import PwaUpdatePrompt from '$lib/components/PwaUpdatePrompt.svelte';
	import { connectionStatus, startConnectionMonitor } from '$lib/services/connection';
	import { startSyncEngine, processSyncQueue } from '$lib/services/sync.svelte';
	import { refreshLocalCache } from '$lib/services/offline';
	import { authStore } from '$lib/stores/auth.svelte';

	let { children } = $props();

	onMount(() => {
		const unsubscribeConnection = connectionStatus.subscribe((status) => {
			if (status === 'online' && authStore.isAuthenticated) {
				void processSyncQueue();
			}
		});
		startConnectionMonitor();
		if (authStore.isAuthenticated) {
			startSyncEngine();
			refreshLocalCache();
		}

		return unsubscribeConnection;
	});
</script>

<svelte:head>
	<title>One Point POS</title>
	<meta name="description" content="Modern Point of Sale System — Manage sales, inventory, customers and more." />
</svelte:head>

<OfflineIndicator />
<PwaInstallPrompt />
<PwaUpdatePrompt />

{@render children()}
