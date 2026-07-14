<script lang="ts">
	import { notify } from '$lib/stores/notification.svelte';
	import { CheckCircle, XCircle, Info, X } from '@lucide/svelte';
</script>

<div class="fixed top-4 right-4 z-50 flex flex-col gap-2 w-80">
	{#each notify.items as n (n.id)}
		<div
			class="flex items-start gap-3 rounded-lg border px-4 py-3 shadow-lg bg-white text-sm"
			class:border-green-200={n.type === 'success'}
			class:border-red-200={n.type === 'error'}
			class:border-blue-200={n.type === 'info'}
		>
			{#if n.type === 'success'}
				<CheckCircle class="shrink-0 mt-0.5 text-green-500" size={16} />
			{:else if n.type === 'error'}
				<XCircle class="shrink-0 mt-0.5 text-red-500" size={16} />
			{:else}
				<Info class="shrink-0 mt-0.5 text-blue-500" size={16} />
			{/if}
			<p class="flex-1 text-gray-800">{n.message}</p>
			<button onclick={() => notify.dismiss(n.id)} class="shrink-0 text-gray-400 hover:text-gray-600">
				<X size={14} />
			</button>
		</div>
	{/each}
</div>
