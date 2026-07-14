<script lang="ts">
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';

	interface Props {
		page: number;
		total: number;
		limit: number;
		onchange: (page: number) => void;
	}

	let { page, total, limit, onchange }: Props = $props();

	const pages = $derived(Math.ceil(total / limit));
</script>

{#if pages > 1}
	<div class="flex items-center justify-between text-sm text-gray-600">
		<span>{total} total</span>
		<div class="flex items-center gap-1">
			<button
				onclick={() => onchange(page - 1)}
				disabled={page <= 1}
				class="rounded p-1 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed"
			>
				<ChevronLeft size={16} />
			</button>
			<span class="px-3">Page {page} of {pages}</span>
			<button
				onclick={() => onchange(page + 1)}
				disabled={page >= pages}
				class="rounded p-1 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed"
			>
				<ChevronRight size={16} />
			</button>
		</div>
	</div>
{/if}
