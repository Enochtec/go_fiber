<script lang="ts">
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth.svelte';
	import { onMount } from 'svelte';

	onMount(() => {
		goto(authStore.isAuthenticated ? '/dashboard' : '/login', { replaceState: true });
	});
</script>
