<script lang="ts">
	import { notify } from '$lib/stores/notification.svelte';
	import { CheckCircle, XCircle, Info, X } from '@lucide/svelte';
</script>

<svelte:head>
	<style>
		@keyframes notif-slide-in {
			from { transform: translateX(100%); opacity: 0; }
			to { transform: translateX(0); opacity: 1; }
		}
		.animate-slide-in {
			animation: notif-slide-in 0.25s ease-out;
		}
	</style>
</svelte:head>

<div class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 w-80 pointer-events-none">
	{#each notify.items as n (n.id)}
		<div
			class="flex items-start gap-3 rounded-lg px-4 py-3 shadow-2xl text-sm font-medium pointer-events-auto animate-slide-in {n.type === 'success' ? 'bg-emerald-600 text-white' : n.type === 'error' ? 'bg-red-600 text-white' : 'bg-blue-600 text-white'}"
		>
			{#if n.type === 'success'}
				<CheckCircle class="shrink-0 mt-0.5" size={18} />
			{:else if n.type === 'error'}
				<XCircle class="shrink-0 mt-0.5" size={18} />
			{:else}
				<Info class="shrink-0 mt-0.5" size={18} />
			{/if}
			<p class="flex-1">{n.message}</p>
			<button onclick={() => notify.dismiss(n.id)} class="shrink-0 opacity-70 hover:opacity-100 transition-opacity">
				<X size={14} />
			</button>
		</div>
	{/each}
</div>
