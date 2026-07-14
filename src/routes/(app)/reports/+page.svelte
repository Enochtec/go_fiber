<script lang="ts">
	import { onMount } from 'svelte';
	import { reportsService } from '$lib/services/reports';
	import { notify } from '$lib/stores/notification.svelte';
	import type { DailySalesRow, TopProductRow, InventoryValueRow } from '$lib/types';
	import { BarChart2, TrendingUp, Package, DollarSign } from '@lucide/svelte';

	let dailySales = $state<DailySalesRow[]>([]);
	let topProducts = $state<TopProductRow[]>([]);
	let inventoryValue = $state<InventoryValueRow[]>([]);
	let loading = $state(true);

	let activeTab = $state<'sales' | 'products' | 'inventory'>('sales');

	const MAX_BAR_HEIGHT = 120;
	const MAX_SALES = $derived(Math.max(...dailySales.map((r) => r.total), 1));
	const totalRevenue = $derived(dailySales.reduce((s, r) => s + r.total, 0));
	const totalOrders = $derived(dailySales.reduce((s, r) => s + r.orders, 0));

	function fmt(n: number) {
		return new Intl.NumberFormat('en-KE', { style: 'currency', currency: 'KES' }).format(n);
	}

	function fmtDate(s: string) {
		const d = new Date(s + 'T00:00:00');
		return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	onMount(async () => {
		try {
			const [s, p, iv] = await Promise.all([
				reportsService.dailySales(30),
				reportsService.topProducts(10),
				reportsService.inventoryValue()
			]);
			dailySales = s.data ?? [];
			topProducts = p.data ?? [];
			inventoryValue = iv.data ?? [];
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to load reports');
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head><title>Reports — POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-slate-900">Reports</h1>
			<p class="text-sm text-slate-500 mt-0.5">Business performance overview</p>
		</div>
	</div>

	<div class="flex gap-1 bg-slate-100 p-1 rounded-xl w-fit">
		{#each [['sales', 'Daily Sales'], ['products', 'Top Products'], ['inventory', 'Inventory Value']] as [tab, label]}
			<button
				onclick={() => activeTab = tab as typeof activeTab}
				class="px-4 py-2 text-sm font-medium rounded-lg transition-all"
				class:bg-white={activeTab === tab}
				class:text-slate-900={activeTab === tab}
				class:shadow-sm={activeTab === tab}
				class:text-slate-500={activeTab !== tab}
				class:hover:text-slate-700={activeTab !== tab}
			>
				{label}
			</button>
		{/each}
	</div>

	{#if loading}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
			{#each Array(2) as _}
				<div class="rounded-2xl bg-slate-200 h-72 animate-pulse"></div>
			{/each}
		</div>
	{:else if activeTab === 'sales'}
		<div class="grid grid-cols-3 gap-4">
			<div class="rounded-2xl bg-white p-5 shadow-sm border border-slate-100">
				<div class="flex items-center gap-3 mb-3">
					<div class="rounded-xl bg-blue-100 p-2.5"><DollarSign size={16} class="text-blue-600" /></div>
					<p class="text-sm font-medium text-slate-500">30-Day Revenue</p>
				</div>
				<p class="text-2xl font-bold text-slate-900">{fmt(totalRevenue)}</p>
			</div>
			<div class="rounded-2xl bg-white p-5 shadow-sm border border-slate-100">
				<div class="flex items-center gap-3 mb-3">
					<div class="rounded-xl bg-emerald-100 p-2.5"><BarChart2 size={16} class="text-emerald-600" /></div>
					<p class="text-sm font-medium text-slate-500">30-Day Orders</p>
				</div>
				<p class="text-2xl font-bold text-slate-900">{totalOrders}</p>
			</div>
			<div class="rounded-2xl bg-white p-5 shadow-sm border border-slate-100">
				<div class="flex items-center gap-3 mb-3">
					<div class="rounded-xl bg-violet-100 p-2.5"><TrendingUp size={16} class="text-violet-600" /></div>
					<p class="text-sm font-medium text-slate-500">Avg. Order Value</p>
				</div>
				<p class="text-2xl font-bold text-slate-900">{fmt(totalOrders > 0 ? totalRevenue / totalOrders : 0)}</p>
			</div>
		</div>

		<div class="rounded-2xl bg-white p-5 shadow-sm border border-slate-100">
			<h2 class="text-sm font-semibold text-slate-800 mb-4">Daily Sales — Last 30 Days</h2>
			{#if dailySales.length === 0}
				<p class="text-sm text-slate-400 py-12 text-center">No sales data</p>
			{:else}
				<div class="flex items-end gap-1.5 h-32">
					{#each dailySales as row}
						<div class="flex-1 flex flex-col items-center gap-1 h-full justify-end">
							<p class="text-[10px] font-medium text-blue-600 leading-none mb-0.5">{fmt(row.total)}</p>
							<div
								class="w-full rounded-t-md bg-blue-500 hover:bg-blue-600 transition-colors"
								style="height: {(row.total / MAX_SALES) * MAX_BAR_HEIGHT}px; min-height: 3px"
							></div>
							<p class="text-[10px] text-slate-400 mt-1 leading-none">{fmtDate(row.date)}</p>
						</div>
					{/each}
				</div>
			{/if}
		</div>

	{:else if activeTab === 'products'}
		<div class="rounded-2xl bg-white shadow-sm border border-slate-100 overflow-hidden">
			<div class="px-5 py-4 border-b border-slate-100">
				<h2 class="text-sm font-semibold text-slate-800">Top Products — 30 Days</h2>
			</div>
			<table class="w-full text-sm">
				<thead>
					<tr class="text-left">
						<th class="px-5 py-3 font-medium text-slate-400 text-xs uppercase tracking-wider">#</th>
						<th class="px-5 py-3 font-medium text-slate-400 text-xs uppercase tracking-wider">Product</th>
						<th class="px-5 py-3 font-medium text-slate-400 text-xs uppercase tracking-wider text-right">Qty Sold</th>
						<th class="px-5 py-3 font-medium text-slate-400 text-xs uppercase tracking-wider text-right">Revenue</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-50">
					{#if topProducts.length === 0}
						<tr><td colspan="4" class="px-5 py-16 text-center text-slate-400">No data</td></tr>
					{:else}
						{#each topProducts as p, i}
							<tr class="hover:bg-slate-50 transition-colors">
								<td class="px-5 py-3.5 text-slate-400 font-medium">{i + 1}</td>
								<td class="px-5 py-3.5 font-medium text-slate-900">{p.product_name}</td>
								<td class="px-5 py-3.5 text-right text-slate-600">{p.quantity}</td>
								<td class="px-5 py-3.5 text-right font-semibold text-slate-900">{fmt(p.revenue)}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>

	{:else}
		<div class="rounded-2xl bg-white shadow-sm border border-slate-100 overflow-hidden">
			<div class="px-5 py-4 border-b border-slate-100">
				<h2 class="text-sm font-semibold text-slate-800 flex items-center gap-2">
					<Package size={15} class="text-slate-400" />
					Inventory Value by Category
				</h2>
			</div>
			<table class="w-full text-sm">
				<thead>
					<tr class="text-left">
						<th class="px-5 py-3 font-medium text-slate-400 text-xs uppercase tracking-wider">Category</th>
						<th class="px-5 py-3 font-medium text-slate-400 text-xs uppercase tracking-wider text-right">Products</th>
						<th class="px-5 py-3 font-medium text-slate-400 text-xs uppercase tracking-wider text-right">Cost Value</th>
						<th class="px-5 py-3 font-medium text-slate-400 text-xs uppercase tracking-wider text-right">Retail Value</th>
						<th class="px-5 py-3 font-medium text-slate-400 text-xs uppercase tracking-wider text-right">Profit</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-50">
					{#if inventoryValue.length === 0}
						<tr><td colspan="5" class="px-5 py-16 text-center text-slate-400">No data</td></tr>
					{:else}
						{#each inventoryValue as row}
							<tr class="hover:bg-slate-50 transition-colors">
								<td class="px-5 py-3.5 font-medium text-slate-900">{row.category_name}</td>
								<td class="px-5 py-3.5 text-right text-slate-500">{row.product_count}</td>
								<td class="px-5 py-3.5 text-right text-slate-600">{fmt(row.total_cost)}</td>
								<td class="px-5 py-3.5 text-right font-semibold text-slate-900">{fmt(row.total_value)}</td>
								<td class="px-5 py-3.5 text-right font-medium text-emerald-600">{fmt(row.total_value - row.total_cost)}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	{/if}
</div>
