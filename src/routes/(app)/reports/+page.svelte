<script lang="ts">
	import { onMount } from 'svelte';
	import { reportsService } from '$lib/services/reports';
	import { notify } from '$lib/stores/notification.svelte';
	import type { DailySalesRow, TopProductRow, InventoryValueRow } from '$lib/types';
	import { BarChart2, TrendingUp, Package, DollarSign, ShoppingCart } from '@lucide/svelte';

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

<svelte:head><title>Reports — Maestro POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-5 min-h-full dark:bg-slate-950">

	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pl-3 border-l-4" style="border-color:#3F00FF;">
		<div>
			<h1 class="text-lg font-bold text-slate-900 dark:text-slate-100">Reports</h1>
			<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Business performance overview</p>
		</div>
	</div>

	<!-- Tabs -->
	<div class="flex gap-1 bg-slate-100 dark:bg-slate-800 p-1 w-fit">
		{#each [['sales', 'Daily Sales'], ['products', 'Top Products'], ['inventory', 'Inventory']] as [tab, label]}
			<button
				onclick={() => activeTab = tab as typeof activeTab}
				class="px-4 py-1.5 text-xs font-semibold transition-all
					{activeTab === tab
						? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm'
						: 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200'}"
			>{label}</button>
		{/each}
	</div>

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
			{#each Array(3) as _}
				<div class="bg-slate-200 dark:bg-slate-700 h-20 animate-pulse"></div>
			{/each}
		</div>
		<div class="bg-slate-200 dark:bg-slate-700 h-64 animate-pulse"></div>

	{:else if activeTab === 'sales'}
		<!-- KPI row -->
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
			{#each [
				{ label: '30-Day Revenue', value: fmt(totalRevenue),                              icon: DollarSign, grad: 'linear-gradient(135deg,#3F00FF,#3200CC)' },
				{ label: '30-Day Orders',  value: String(totalOrders),                            icon: ShoppingCart, grad: 'linear-gradient(135deg,#7B68EE,#6A5ACD)' },
				{ label: 'Avg. Order',     value: fmt(totalOrders > 0 ? totalRevenue/totalOrders : 0), icon: TrendingUp, grad: 'linear-gradient(135deg,#FFD700,#DAA520)' },
			] as kpi}
				<div class="relative overflow-hidden rounded-[1px] p-5 text-white" style="background:{kpi.grad};">
					<div class="absolute -top-6 -right-6 h-24 w-24 bg-white/10"></div>
					<div class="absolute -bottom-6 -left-6 h-20 w-20 bg-white/5"></div>
					<div class="flex items-start justify-between relative">
						<div>
							<p class="text-xs font-semibold uppercase tracking-wider text-white/80">{kpi.label}</p>
							<p class="text-xl font-bold mt-1 tracking-tight tabular-nums">{kpi.value}</p>
						</div>
						<div class="flex h-10 w-10 items-center justify-center bg-white/20">
							<kpi.icon size={20} class="text-white" />
						</div>
					</div>
				</div>
			{/each}
		</div>

		<!-- Bar chart -->
		<div class="rounded-[1px] bg-white dark:bg-slate-800 p-5">
			<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100 mb-5">Daily Sales — Last 30 Days</h2>
			{#if dailySales.length === 0}
				<p class="text-sm text-slate-400 py-12 text-center">No sales data</p>
			{:else}
				<div class="flex items-end gap-1 h-36 overflow-x-auto">
					{#each dailySales as row}
						<div class="group flex-1 min-w-[10px] flex flex-col items-center gap-1 h-full justify-end">
							<div class="relative w-full">
								<div class="absolute -top-6 left-1/2 -translate-x-1/2 whitespace-nowrap rounded bg-slate-900 px-1.5 py-0.5 text-[10px] font-semibold text-white opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
									{fmt(row.total)}
								</div>
								<div class="w-full bg-blue-600 hover:bg-blue-700 transition-colors"
									style="height:{Math.max((row.total/MAX_SALES)*MAX_BAR_HEIGHT, 2)}px;"
								></div>
							</div>
							<p class="text-[8px] text-slate-400 leading-none whitespace-nowrap">{fmtDate(row.date)}</p>
						</div>
					{/each}
				</div>
			{/if}
		</div>

	{:else if activeTab === 'products'}
		<div class="rounded-[1px] bg-white dark:bg-slate-800 overflow-hidden">
			<div class="px-5 py-3.5 border-b border-slate-100 dark:border-slate-700">
				<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Top Products — 30 Days</h2>
			</div>
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
							<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide">#</th>
							<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide">Product</th>
							<th class="px-4 py-2.5 text-right text-xs font-semibold text-white uppercase tracking-wide">Units Sold</th>
							<th class="px-4 py-2.5 text-right text-xs font-semibold text-white uppercase tracking-wide">Revenue</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
						{#if topProducts.length === 0}
							<tr><td colspan="4" class="px-5 py-16 text-center text-slate-400 dark:text-slate-500">No data yet</td></tr>
						{:else}
							{#each topProducts as p, i}
								<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-colors">
									<td class="px-5 py-3 text-xs font-bold text-slate-300 dark:text-slate-600 tabular-nums">{i+1}</td>
									<td class="px-5 py-3 font-medium text-slate-900 dark:text-slate-100">{p.product_name}</td>
									<td class="px-5 py-3 text-right text-slate-500 dark:text-slate-400 tabular-nums">{p.quantity_sold}</td>
									<td class="px-5 py-3 text-right font-semibold text-slate-900 dark:text-slate-100 tabular-nums">{fmt(p.revenue)}</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</div>

	{:else}
		<div class="rounded-[1px] bg-white dark:bg-slate-800 overflow-hidden">
			<div class="px-5 py-3.5 border-b border-slate-100 dark:border-slate-700 flex items-center gap-2">
				<Package size={14} class="text-slate-400" />
				<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Inventory Value by Category</h2>
			</div>
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
							{#each ['Category', 'Products', 'Cost Value', 'Retail Value', 'Gross Profit'] as h, i}
								<th class="px-4 py-2.5 text-xs font-semibold text-white uppercase tracking-wide {i > 0 ? 'text-right' : 'text-left'}">{h}</th>
							{/each}
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
						{#if inventoryValue.length === 0}
							<tr><td colspan="5" class="px-5 py-16 text-center text-slate-400 dark:text-slate-500">No data</td></tr>
						{:else}
							{#each inventoryValue as row}
								{@const profit = row.total_value - row.total_cost}
								{@const margin = row.total_value > 0 ? (profit / row.total_value * 100) : 0}
								<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-colors">
									<td class="px-5 py-3 font-medium text-slate-900 dark:text-slate-100">{row.category_name}</td>
									<td class="px-5 py-3 text-right text-slate-500 dark:text-slate-400 tabular-nums">{row.product_count}</td>
									<td class="px-5 py-3 text-right text-slate-500 dark:text-slate-400 tabular-nums">{fmt(row.total_cost)}</td>
									<td class="px-5 py-3 text-right font-semibold text-slate-900 dark:text-slate-100 tabular-nums">{fmt(row.total_value)}</td>
									<td class="px-5 py-3 text-right">
										<span class="font-semibold text-emerald-600 dark:text-emerald-400 tabular-nums">{fmt(profit)}</span>
										<span class="text-[10px] text-slate-400 ml-1">({margin.toFixed(1)}%)</span>
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
