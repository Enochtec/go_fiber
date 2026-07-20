<script lang="ts">
	import { X, FileText, Table, Download, Loader } from '@lucide/svelte';

	interface Props {
		open: boolean;
		title?: string;
		hasFiltered?: boolean;
		hasSelected?: boolean;
		onclose: () => void;
		onexport: (format: 'csv', scope: 'all' | 'filtered' | 'current' | 'selected') => Promise<void>;
	}

	let {
		open,
		title       = 'Export Data',
		hasFiltered = false,
		hasSelected = false,
		onclose,
		onexport
	}: Props = $props();

	let format    = $state<'csv'>('csv');
	let scope     = $state<'all' | 'filtered' | 'current' | 'selected'>('all');
	let exporting = $state(false);
	let error     = $state('');

	async function doExport() {
		exporting = true;
		error     = '';
		try {
			await onexport(format, scope);
			onclose();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Export failed';
		} finally {
			exporting = false;
		}
	}
</script>

{#if open}
<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
	role="dialog"
	aria-modal="true"
>
	<div class="relative w-full max-w-md bg-white shadow-xl overflow-hidden">
		<!-- Header -->
		<div class="flex items-center justify-between px-5 py-4 border-b border-slate-100">
			<h2 class="text-base font-bold text-slate-900">{title}</h2>
			<button onclick={onclose} class="h-7 w-7 flex items-center justify-center text-slate-400 hover:text-slate-600 transition-colors">
				<X size={16} />
			</button>
		</div>

		<div class="p-5 space-y-5">
			<!-- Format selection -->
			<div>
				<p class="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">File Format</p>
				<div class="flex gap-2">
					<button
						onclick={() => format = 'csv'}
						class="flex-1 flex flex-col items-center gap-1.5 border-2 px-3 py-3 text-sm font-medium transition-all {format === 'csv' ? 'border-teal-500 bg-teal-50 text-teal-700' : 'border-slate-200 text-slate-500 hover:border-slate-300'}"
					>
						<Table size={20} />
						<span>CSV</span>
						<span class="text-[10px] font-normal opacity-70">Excel compatible</span>
					</button>
					<button
						disabled
						class="flex-1 flex flex-col items-center gap-1.5 border-2 border-dashed border-slate-200 px-3 py-3 text-sm font-medium text-slate-300 cursor-not-allowed"
						title="Coming soon"
					>
						<FileText size={20} />
						<span>PDF</span>
						<span class="text-[10px] font-normal">Coming soon</span>
					</button>
				</div>
			</div>

			<!-- Scope selection -->
			<div>
				<p class="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">Export Scope</p>
				<div class="space-y-1.5">
					{#each [
						{ value: 'all',      label: 'All Records',         desc: 'Export every record in this module' },
						{ value: 'filtered', label: 'Filtered Results',    desc: 'Only records matching active filters', disabled: !hasFiltered },
						{ value: 'selected', label: 'Selected Rows',       desc: 'Only checked/highlighted rows', disabled: !hasSelected },
						{ value: 'current',  label: 'Current Page',        desc: 'Only the rows visible on screen' },
					] as opt}
						<button
							onclick={() => !opt.disabled && (scope = opt.value as typeof scope)}
							disabled={opt.disabled}
							class="w-full flex items-center gap-3 border px-3 py-2.5 text-left transition-all {scope === opt.value ? 'border-teal-500 bg-teal-50' : 'border-slate-200 hover:border-slate-300'} {opt.disabled ? 'opacity-40 cursor-not-allowed' : ''}"
						>
							<span class="h-4 w-4 flex-shrink-0 border-2 flex items-center justify-center {scope === opt.value ? 'border-teal-500 bg-teal-500' : 'border-slate-300'}">
								{#if scope === opt.value}
									<span class="w-1.5 h-1.5 bg-white block"></span>
								{/if}
							</span>
							<span class="flex-1 min-w-0">
								<span class="block text-sm font-medium text-slate-900">{opt.label}</span>
								<span class="block text-xs text-slate-400">{opt.desc}</span>
							</span>
						</button>
					{/each}
				</div>
			</div>

			{#if error}
				<p class="text-sm text-red-600 bg-red-50 px-3 py-2">{error}</p>
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex items-center justify-between gap-3 px-5 py-4 border-t border-slate-100 bg-slate-50">
			<button onclick={onclose} class="px-4 py-2 text-sm font-medium text-slate-600 border border-slate-200 hover:bg-white transition-colors">
				Cancel
			</button>
			<button
				onclick={doExport}
				disabled={exporting}
				class="flex items-center gap-2 px-5 py-2 text-sm font-semibold text-white bg-teal-600 hover:bg-teal-700 disabled:opacity-60 transition-colors"
			>
				{#if exporting}
					<Loader size={14} class="animate-spin" />
					Exporting…
				{:else}
					<Download size={14} />
					Export {format.toUpperCase()}
				{/if}
			</button>
		</div>
	</div>
</div>
{/if}
