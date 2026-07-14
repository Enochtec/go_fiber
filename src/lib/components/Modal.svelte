<script lang="ts">
	import { X } from '@lucide/svelte';

	interface Props {
		open: boolean;
		title: string;
		onclose: () => void;
		children: import('svelte').Snippet;
		footer?: import('svelte').Snippet;
		size?: 'sm' | 'md' | 'lg';
	}

	let { open, title, onclose, children, footer, size = 'md' }: Props = $props();

	const widths = { sm: 'max-w-sm', md: 'max-w-lg', lg: 'max-w-2xl' };
</script>

{#if open}
	<div
		class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 backdrop-blur-sm p-0 sm:p-4"
		role="dialog"
		aria-modal="true"
		onclick={onclose}
		onkeydown={(e) => e.key === 'Escape' && onclose()}
	>
		<div
			class="relative w-full {widths[size]} max-h-[92vh] flex flex-col
				rounded-t-3xl sm:rounded-2xl bg-white shadow-2xl overflow-hidden"
			onclick={(e) => e.stopPropagation()}
			role="presentation"
		>
			<!-- Mobile handle -->
			<div class="flex justify-center pt-3 pb-1 sm:hidden">
				<div class="h-1 w-10 rounded-full bg-slate-300"></div>
			</div>

			<div class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
				<h2 class="text-base font-semibold text-slate-900">{title}</h2>
				<button
					onclick={onclose}
					class="rounded-xl p-1.5 text-slate-400 hover:bg-slate-100 hover:text-slate-700 transition-colors"
				>
					<X size={18} />
				</button>
			</div>

			<div class="px-5 py-4 overflow-y-auto flex-1">
				{@render children()}
			</div>

			{#if footer}
				<div class="flex justify-end gap-3 border-t border-slate-100 px-5 py-4">
					{@render footer()}
				</div>
			{/if}
		</div>
	</div>
{/if}
