<script lang="ts">
	import { onMount } from 'svelte';
	import { salesService } from '$lib/services/sales';
	import { notify } from '$lib/stores/notification.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import type { Sale } from '$lib/types';
	import { Eye, Download } from '@lucide/svelte';
	import ExportModal from '$lib/components/ExportModal.svelte';
	import { shopService } from '$lib/services/shop';
	import { authStore } from '$lib/stores/auth.svelte';
	import { exportSales } from '$lib/services/export';

	let sales = $state<Sale[]>([]);
	let total = $state(0);
	let page = $state(1);
	const limit = 20;
	let loading = $state(true);
	let showExport = $state(false);

	let statusFilter = $state('');
	let dateFrom = $state('');
	let dateTo = $state('');

	let selectedSale = $state<Sale | null>(null);
	let showDetail = $state(false);

	function fmt(n: number) {
		return new Intl.NumberFormat('en-KE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}

	function fmtDate(s: string) {
		return new Date(s).toLocaleString('en-US', {
			year: 'numeric', month: 'short', day: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}

	async function fetch() {
		loading = true;
		try {
			const res = await salesService.list({
				status: statusFilter as 'completed' | 'held' | 'voided' | undefined || undefined,
				date_from: dateFrom || undefined,
				date_to: dateTo || undefined,
				page,
				limit
			});
			sales = res.data ?? [];
			total = res.total ?? 0;
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to load sales');
		} finally {
			loading = false;
		}
	}

	async function viewDetail(sale: Sale) {
		try {
			const res = await salesService.getById(sale.id);
			selectedSale = res.data ?? null;
			showDetail = true;
		} catch {
			notify.error('Failed to load sale details');
		}
	}

	async function voidSale(id: string) {
		if (!confirm('Void this sale? Stock will be returned.')) return;
		try {
			await salesService.void(id);
			notify.success('Sale voided');
			fetch();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to void sale');
		}
	}

	const statusColors: Record<string, string> = {
		completed: 'badge-green',
		held:      'badge-amber',
		voided:    'badge-red'
	};

	async function handleExport(_fmt: 'csv', scope: 'all' | 'filtered' | 'current' | 'selected') {
		const info = await shopService.getInfo();
		const shopName = info?.shop?.name ?? 'Export';
		const userName = authStore.user?.name ?? 'System';
		let data: Sale[];
		if (scope === 'current') {
			data = sales;
		} else {
			const res = await salesService.list({
				status: scope === 'filtered' ? (statusFilter as 'completed' | 'held' | 'voided' | undefined || undefined) : undefined,
				date_from: scope === 'filtered' ? (dateFrom || undefined) : undefined,
				date_to: scope === 'filtered' ? (dateTo || undefined) : undefined,
				page: 1, limit: 10000
			});
			data = res.data ?? [];
		}
		await exportSales(data, shopName, userName);
	}

	onMount(fetch);
</script>

<svelte:head><title>Sales History — Maestro POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-5 min-h-full dark:bg-slate-950">
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
		<div>
			<h1 class="text-lg font-bold text-slate-900 dark:text-slate-100">Sales History</h1>
			<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
				{total} transaction{total !== 1 ? 's' : ''} found
			</p>
		</div>
		<button onclick={() => showExport = true} class="flex items-center gap-2 px-3 py-2 text-sm font-semibold text-slate-600 border border-slate-200 hover:bg-slate-50 transition-all">
			<Download size={14} /> Export
		</button>
	</div>

	<!-- Filters -->
	<div class="flex gap-2 flex-wrap">
		<select
			bind:value={statusFilter}
			onchange={() => { page = 1; fetch(); }}
			class="rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 px-3 py-1.5 text-sm text-slate-700 dark:text-slate-200 focus:border-blue-500 focus:outline-none transition-colors"
		>
			<option value="">All Status</option>
			<option value="completed">Completed</option>
			<option value="held">Held</option>
			<option value="voided">Voided</option>
		</select>
		<input
			type="date"
			bind:value={dateFrom}
			onchange={() => { page = 1; fetch(); }}
			class="rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 px-3 py-1.5 text-sm text-slate-700 dark:text-slate-200 focus:border-blue-500 focus:outline-none transition-colors"
		/>
		<input
			type="date"
			bind:value={dateTo}
			onchange={() => { page = 1; fetch(); }}
			class="rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 px-3 py-1.5 text-sm text-slate-700 dark:text-slate-200 focus:border-blue-500 focus:outline-none transition-colors"
		/>
	</div>

	<div class="rounded-[1px] bg-white dark:bg-slate-800 overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead>
					<tr style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
						<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide whitespace-nowrap">Date</th>
						<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide hidden sm:table-cell">Cashier</th>
						<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide hidden md:table-cell">Customer</th>
						<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide hidden lg:table-cell">Payment</th>
						<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide">Status</th>
						<th class="px-4 py-2.5 text-right text-xs font-semibold text-white uppercase tracking-wide">Total</th>
						<th class="px-4 py-2.5 text-right text-xs font-semibold text-white uppercase tracking-wide">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
					{#if loading}
						{#each Array(8) as _}
							<tr>{#each Array(7) as _}<td class="px-4 py-2.5"><div class="h-3.5 bg-slate-100 dark:bg-slate-700 rounded-[1px] animate-pulse"></div></td>{/each}</tr>
						{/each}
					{:else if sales.length === 0}
						<tr><td colspan="7" class="px-5 py-12 text-center text-sm text-slate-400 dark:text-slate-500">No sales found for the selected filters</td></tr>
					{:else}
						{#each sales as s}
							<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-colors">
								<td class="px-4 py-2.5 text-slate-600 dark:text-slate-300 whitespace-nowrap">{fmtDate(s.created_at)}</td>
								<td class="px-4 py-2.5 text-slate-700 dark:text-slate-200 hidden sm:table-cell">{s.cashier_name ?? '—'}</td>
								<td class="px-4 py-2.5 text-slate-500 dark:text-slate-400 hidden md:table-cell">{s.customer_name ?? 'Walk-in'}</td>
								<td class="px-4 py-2.5 text-slate-500 dark:text-slate-400 capitalize hidden lg:table-cell">{s.payment_method}</td>
								<td class="px-4 py-2.5">
									<span class="badge {statusColors[s.status] ?? 'badge-slate'} capitalize">{s.status}</span>
								</td>
								<td class="px-4 py-2.5 text-right font-semibold text-slate-900 dark:text-slate-100 tabular-nums">KES {fmt(s.total)}</td>
								<td class="px-4 py-2.5 text-right">
									<div class="flex items-center justify-end gap-1">
										<button onclick={() => viewDetail(s)} class="h-7 w-7 flex items-center justify-center rounded-[1px] text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 hover:text-blue-600 transition-colors">
											<Eye size={13} />
										</button>
										{#if s.status !== 'voided'}
											<button onclick={() => voidSale(s.id)} class="rounded-[1px] px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors">
												Void
											</button>
										{/if}
									</div>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>

	<Pagination {page} {total} {limit} onchange={(p) => { page = p; fetch(); }} />
</div>

<Modal open={showDetail} title="Sale Details" onclose={() => showDetail = false} size="md">
	{#snippet children()}
		{#if selectedSale}
			<div class="space-y-4">
				<!-- Meta grid -->
				<div class="grid grid-cols-2 gap-2 bg-slate-50 dark:bg-slate-900/40 p-3 text-xs">
					{#each [
						['Date',     fmtDate(selectedSale.created_at)],
						['Status',   selectedSale.status],
						['Cashier',  selectedSale.cashier_name ?? '—'],
						['Customer', selectedSale.customer_name ?? 'Walk-in'],
						['Payment',  selectedSale.payment_method],
					] as [label, value]}
						<div>
							<p class="text-slate-400 dark:text-slate-500 uppercase tracking-wide mb-0.5">{label}</p>
							<p class="font-semibold text-slate-800 dark:text-slate-200 capitalize">{value}</p>
						</div>
					{/each}
				</div>

				{#if selectedSale.items && selectedSale.items.length > 0}
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b border-slate-100 dark:border-slate-700 text-left">
								<th class="pb-2 text-xs font-semibold text-white uppercase tracking-wide">Item</th>
								<th class="pb-2 text-xs font-semibold text-white uppercase tracking-wide text-right">Qty</th>
								<th class="pb-2 text-xs font-semibold text-white uppercase tracking-wide text-right">Price</th>
								<th class="pb-2 text-xs font-semibold text-white uppercase tracking-wide text-right">Total</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
							{#each selectedSale.items as item}
								<tr>
									<td class="py-2 text-slate-800 dark:text-slate-200">{item.product_name ?? '—'}</td>
									<td class="py-2 text-right text-slate-500 dark:text-slate-400 tabular-nums">{item.quantity}</td>
									<td class="py-2 text-right text-slate-500 dark:text-slate-400 tabular-nums">KES {fmt(item.unit_price)}</td>
									<td class="py-2 text-right font-semibold text-slate-800 dark:text-slate-200 tabular-nums">KES {fmt(item.total)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				{/if}

				<div class="border-t border-slate-100 dark:border-slate-700 pt-3 space-y-1.5 text-sm">
					<div class="flex justify-between text-slate-500 dark:text-slate-400"><span>Subtotal</span><span class="tabular-nums">KES {fmt(selectedSale.subtotal)}</span></div>
					{#if selectedSale.discount > 0}
						<div class="flex justify-between text-emerald-600 dark:text-emerald-400"><span>Discount</span><span class="tabular-nums">−KES {fmt(selectedSale.discount)}</span></div>
					{/if}
					{#if selectedSale.tax > 0}
						<div class="flex justify-between text-slate-500 dark:text-slate-400"><span>Tax</span><span class="tabular-nums">KES {fmt(selectedSale.tax)}</span></div>
					{/if}
					<div class="flex justify-between font-bold text-slate-900 dark:text-slate-100 text-base border-t border-slate-100 dark:border-slate-700 pt-2">
						<span>Total</span><span class="tabular-nums">KES {fmt(selectedSale.total)}</span>
					</div>
				</div>
			</div>
		{/if}
	{/snippet}
</Modal>

<ExportModal
	open={showExport}
	title="Export Sales History"
	hasFiltered={!!(statusFilter || dateFrom || dateTo)}
	onclose={() => showExport = false}
	onexport={handleExport}
/>
