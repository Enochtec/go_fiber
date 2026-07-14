<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>POS System</title>
	<meta name="description" content="Modern Point of Sale System" />
</svelte:head>

{@render children()}
