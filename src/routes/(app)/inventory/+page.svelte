<script lang="ts">
	import { onMount } from 'svelte';
	import { inventoryService, type AdjustInput } from '$lib/services/inventory';
	import { productsService } from '$lib/services/products';
	import { reportsService } from '$lib/services/reports';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import type { StockAdjustment, Product, InventoryValueRow, TopProductRow } from '$lib/types';
	import { Plus, AlertTriangle, Search, Package, TrendingUp, DollarSign, BarChart2, ArrowUp, ArrowDown } from '@lucide/svelte';

	// ─── State ───────────────────────────────────────────────────
	let adjustments = $state<StockAdjustment[]>([]);
	let adjTotal = $state(0);
	let adjPage = $state(1);
	const adjLimit = 15;
	let loading = $state(true);

	let stockValue = $state<InventoryValueRow[]>([]);
	let lowStockItems = $state<Product[]>([]);
	let fastMovers = $state<TopProductRow[]>([]);
	let overviewLoading = $state(true);

	// ─── Modal state ────────────────────────────────────────────
	let showModal = $state(false);
	let submitting = $state(false);
	let productSearch = $state('');
	let searchResults = $state<Product[]>([]);
	let selectedProduct = $state<Product | null>(null);
	let form = $state<AdjustInput>({ product_id: '', quantity: 0, reason: '' });

	// ─── Derived ────────────────────────────────────────────────
	const totalRetailValue = $derived(stockValue.reduce((s, r) => s + r.total_value, 0));
	const totalCostValue = $derived(stockValue.reduce((s, r) => s + r.total_cost, 0));
	const totalProducts = $derived(stockValue.reduce((s, r) => s + r.product_count, 0));

	// ─── Fetching ────────────────────────────────────────────────
	async function fetchAdjustments() {
		loading = true;
		try {
			const res = await inventoryService.listAdjustments(adjPage, adjLimit);
			adjustments = res.data ?? [];
			adjTotal = res.total ?? 0;
		} finally {
			loading = false;
		}
	}

	async function fetchOverview() {
		overviewLoading = true;
		try {
			const [valRes, lowRes, fastRes] = await Promise.all([
				reportsService.inventoryValue(),
				productsService.list({ low_stock: true, limit: 20 }),
				reportsService.topProducts(8)
			]);
			stockValue = valRes.data ?? [];
			lowStockItems = lowRes.data ?? [];
			fastMovers = fastRes.data ?? [];
		} finally {
			overviewLoading = false;
		}
	}

	async function searchProducts() {
		if (!productSearch.trim()) { searchResults = []; return; }
		const res = await productsService.list({ search: productSearch, limit: 8 });
		searchResults = res.data ?? [];
	}

	function selectProduct(p: Product) {
		selectedProduct = p;
		form.product_id = p.id;
		productSearch = p.name;
		searchResults = [];
	}

	async function save() {
		if (!form.product_id) { notify.error('Select a product'); return; }
		if (form.quantity === 0) { notify.error('Quantity cannot be zero'); return; }
		if (!form.reason.trim()) { notify.error('Reason is required'); return; }
		submitting = true;
		try {
			await inventoryService.adjust(form);
			notify.success('Stock adjusted');
			showModal = false;
			form = { product_id: '', quantity: 0, reason: '' };
			selectedProduct = null;
			productSearch = '';
			fetchAdjustments();
			fetchOverview();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Adjustment failed');
		} finally {
			submitting = false;
		}
	}

	function fmt(n: number) {
		return new Intl.NumberFormat('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(n);
	}
	function fmtDate(s: string) {
		return new Date(s).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	onMount(() => { fetchAdjustments(); fetchOverview(); });
</script>

<svelte:head><title>Inventory — POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-5 bg-slate-50 dark:bg-slate-950 min-h-full">

	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-slate-900 dark:text-slate-100">Inventory</h1>
			<p class="text-xs text-slate-400 mt-0.5">Stock overview and adjustments</p>
		</div>
		<button
			onclick={() => showModal = true}
			class="flex items-center gap-2 rounded-xl px-4 py-2.5 text-sm font-semibold text-white transition-all active:scale-95"
			style="background-color:#008B8B;"
		>
			<Plus size={15} /> Adjust Stock
		</button>
	</div>

	{#if overviewLoading}
		<!-- Skeleton overview -->
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
			{#each Array(3) as _}
				<div class="rounded-2xl bg-slate-200 dark:bg-slate-700 h-24 animate-pulse"></div>
			{/each}
		</div>
	{:else}
		<!-- ── Stock value cards ─────────────────────────────────── -->
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
			<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4 shadow-sm">
				<div class="flex items-center justify-between mb-2">
					<span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Retail Value</span>
					<div class="h-7 w-7 rounded-lg flex items-center justify-center" style="background-color:#008B8B15;">
						<DollarSign size={14} style="color:#008B8B;" />
					</div>
				</div>
				<p class="text-2xl font-bold text-slate-900 dark:text-slate-100 tabular-nums">KES {fmt(totalRetailValue)}</p>
				<p class="text-xs text-slate-400 mt-1">across {totalProducts} products</p>
			</div>
			<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4 shadow-sm">
				<div class="flex items-center justify-between mb-2">
					<span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Cost Value</span>
					<div class="h-7 w-7 rounded-lg flex items-center justify-center bg-blue-50 dark:bg-blue-900/20">
						<BarChart2 size={14} class="text-blue-500" />
					</div>
				</div>
				<p class="text-2xl font-bold text-slate-900 dark:text-slate-100 tabular-nums">KES {fmt(totalCostValue)}</p>
				<p class="text-xs text-emerald-600 mt-1">Margin: KES {fmt(totalRetailValue - totalCostValue)}</p>
			</div>
			<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4 shadow-sm">
				<div class="flex items-center justify-between mb-2">
					<span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Stock Alerts</span>
					<div class="h-7 w-7 rounded-lg flex items-center justify-center bg-amber-50 dark:bg-amber-900/20">
						<AlertTriangle size={14} class="text-amber-500" />
					</div>
				</div>
				<p class="text-2xl font-bold {lowStockItems.length > 0 ? 'text-amber-600' : 'text-slate-900 dark:text-slate-100'} tabular-nums">
					{lowStockItems.length}
				</p>
				<p class="text-xs text-slate-400 mt-1">
					{lowStockItems.filter(p => p.stock_qty === 0).length} out of stock
				</p>
			</div>
		</div>

		<!-- ── Middle row: Low stock + Fast movers ────────────── -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

			<!-- Low stock -->
			<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 shadow-sm overflow-hidden">
				<div class="flex items-center gap-2 px-5 py-3.5 border-b border-slate-100 dark:border-slate-700">
					<AlertTriangle size={14} class="text-amber-500 shrink-0" />
					<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Low Stock Items</h2>
					{#if lowStockItems.length > 0}
						<span class="ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 text-[10px] font-bold px-1.5">{lowStockItems.length}</span>
					{/if}
				</div>
				{#if lowStockItems.length === 0}
					<div class="flex flex-col items-center gap-2 py-10 text-slate-300 dark:text-slate-600">
						<Package size={28} class="opacity-50" />
						<p class="text-xs text-slate-400">All products well-stocked</p>
					</div>
				{:else}
					<ul class="divide-y divide-slate-50 dark:divide-slate-700/50 max-h-64 overflow-y-auto">
						{#each lowStockItems as p}
							<li class="flex items-center justify-between px-5 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
								<div class="min-w-0 flex-1">
									<p class="text-sm font-medium text-slate-700 dark:text-slate-200 truncate">{p.name}</p>
									{#if p.category_name}
										<p class="text-[10px] text-slate-400">{p.category_name}</p>
									{/if}
								</div>
								<div class="flex items-center gap-2 shrink-0 ml-3">
									<span class="text-xs text-slate-400">Reorder: {p.reorder_level}</span>
									<span class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-bold {p.stock_qty === 0 ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'}">
										{p.stock_qty === 0 ? 'Out' : p.stock_qty}
									</span>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<!-- Fast movers -->
			<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 shadow-sm overflow-hidden">
				<div class="flex items-center gap-2 px-5 py-3.5 border-b border-slate-100 dark:border-slate-700">
					<TrendingUp size={14} style="color:#008B8B;" class="shrink-0" />
					<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Fast Movers — 30 Days</h2>
				</div>
				{#if fastMovers.length === 0}
					<div class="flex items-center justify-center py-10 text-slate-400 text-sm">No sales data yet</div>
				{:else}
					{@const maxQty = Math.max(...fastMovers.map(f => f.quantity_sold))}
					<ul class="divide-y divide-slate-50 dark:divide-slate-700/50 max-h-64 overflow-y-auto">
						{#each fastMovers as p, i}
							<li class="px-5 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
								<div class="flex items-center justify-between mb-1.5">
									<div class="flex items-center gap-2 min-w-0">
										<span class="text-xs font-bold text-slate-300 dark:text-slate-600 shrink-0 w-4">{i+1}</span>
										<span class="text-sm font-medium text-slate-700 dark:text-slate-200 truncate">{p.product_name}</span>
									</div>
									<div class="flex items-center gap-2 shrink-0 ml-2">
										<span class="text-xs text-slate-400">{p.quantity_sold} sold</span>
										<span class="text-xs font-bold text-slate-700 dark:text-slate-200 tabular-nums">KES {fmt(p.revenue)}</span>
									</div>
								</div>
								<div class="h-1.5 rounded-full bg-slate-100 dark:bg-slate-700 overflow-hidden">
									<div class="h-full rounded-full" style="width:{maxQty > 0 ? (p.quantity_sold / maxQty) * 100 : 0}%; background-color:#008B8B;"></div>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		</div>

		<!-- ── Stock value by category ───────────────────────── -->
		{#if stockValue.length > 0}
			<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 shadow-sm overflow-hidden">
				<div class="px-5 py-3.5 border-b border-slate-100 dark:border-slate-700">
					<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Stock Value by Category</h2>
				</div>
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead>
							<tr class="bg-slate-50 dark:bg-slate-900/50">
								<th class="px-5 py-2.5 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Category</th>
								<th class="px-5 py-2.5 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Products</th>
								<th class="px-5 py-2.5 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Cost Value</th>
								<th class="px-5 py-2.5 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Retail Value</th>
								<th class="px-5 py-2.5 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Margin</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
							{#each stockValue as row}
								{@const margin = row.total_cost > 0 ? ((row.total_value - row.total_cost) / row.total_cost * 100) : 0}
								<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
									<td class="px-5 py-3 font-medium text-slate-800 dark:text-slate-100">{row.category_name}</td>
									<td class="px-5 py-3 text-right text-slate-500 dark:text-slate-400 tabular-nums">{row.product_count}</td>
									<td class="px-5 py-3 text-right text-slate-600 dark:text-slate-300 tabular-nums">KES {fmt(row.total_cost)}</td>
									<td class="px-5 py-3 text-right font-semibold text-slate-800 dark:text-slate-100 tabular-nums">KES {fmt(row.total_value)}</td>
									<td class="px-5 py-3 text-right">
										<span class="inline-flex items-center gap-1 text-xs font-semibold {margin >= 0 ? 'text-emerald-600' : 'text-red-500'}">
											{#if margin >= 0}<ArrowUp size={10} />{:else}<ArrowDown size={10} />{/if}
											{Math.abs(margin).toFixed(1)}%
										</span>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	{/if}

	<!-- ── Adjustments log ───────────────────────────────────── -->
	<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 shadow-sm overflow-hidden">
		<div class="px-5 py-3.5 border-b border-slate-100 dark:border-slate-700">
			<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Adjustment History</h2>
		</div>
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead>
					<tr class="bg-slate-50 dark:bg-slate-900/50">
						{#each ['Product', 'Qty', 'Reason', 'By', 'Date'] as h}
							<th class="px-4 py-2.5 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide {h === 'Qty' ? 'text-center' : ''}">{h}</th>
						{/each}
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
					{#if loading}
						{#each Array(5) as _}
							<tr>{#each Array(5) as _}<td class="px-4 py-3"><div class="h-4 rounded bg-slate-100 dark:bg-slate-700 animate-pulse"></div></td>{/each}</tr>
						{/each}
					{:else if adjustments.length === 0}
						<tr><td colspan="5" class="px-4 py-10 text-center text-sm text-slate-400">No adjustments yet</td></tr>
					{:else}
						{#each adjustments as a}
							<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
								<td class="px-4 py-3 font-medium text-slate-800 dark:text-slate-100">{a.product_name ?? '—'}</td>
								<td class="px-4 py-3 text-center">
									<span class="inline-flex items-center gap-1 font-bold text-xs {a.quantity > 0 ? 'text-emerald-600' : 'text-red-500'}">
										{#if a.quantity > 0}<ArrowUp size={11} />{:else}<ArrowDown size={11} />{/if}
										{Math.abs(a.quantity)}
									</span>
								</td>
								<td class="px-4 py-3 text-slate-600 dark:text-slate-300 max-w-48 truncate">{a.reason}</td>
								<td class="px-4 py-3 text-slate-500 dark:text-slate-400">{a.user_name ?? '—'}</td>
								<td class="px-4 py-3 text-slate-400 text-xs">{fmtDate(a.created_at)}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
	<Pagination page={adjPage} total={adjTotal} limit={adjLimit} onchange={(p) => { adjPage = p; fetchAdjustments(); }} />
</div>

<!-- ─── Stock Adjustment Modal ───────────────────────────────── -->
<Modal open={showModal} title="Stock Adjustment" onclose={() => showModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-4">
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Product *</label>
				<div class="relative">
					<Search size={13} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
					<input
						bind:value={productSearch}
						oninput={searchProducts}
						placeholder="Search product…"
						class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 py-2.5 pl-9 pr-4 text-sm focus:outline-none"
					/>
				</div>
				{#if searchResults.length > 0}
					<ul class="mt-1.5 rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 shadow-lg max-h-44 overflow-y-auto divide-y divide-slate-50 dark:divide-slate-700">
						{#each searchResults as p}
							<li>
								<button onclick={() => selectProduct(p)} class="w-full flex items-center justify-between px-4 py-2.5 text-left text-sm hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
									<span class="font-medium text-slate-800 dark:text-slate-100 truncate">{p.name}</span>
									<span class="text-xs text-slate-400 shrink-0 ml-2">Stock: {p.stock_qty}</span>
								</button>
							</li>
						{/each}
					</ul>
				{/if}
				{#if selectedProduct}
					<p class="mt-1.5 text-xs font-medium" style="color:#008B8B;">Current stock: {selectedProduct.stock_qty} units</p>
				{/if}
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Quantity <span class="normal-case font-normal">(negative to deduct)</span></label>
				<input type="number" bind:value={form.quantity} class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Reason *</label>
				<input bind:value={form.reason} placeholder="e.g. Damaged, Stock count, Return…" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-xl border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-xl px-5 py-2.5 text-sm font-semibold text-white disabled:opacity-50 transition-all active:scale-95" style="background-color:#008B8B;">
			{submitting ? 'Saving…' : 'Apply'}
		</button>
	{/snippet}
</Modal>
