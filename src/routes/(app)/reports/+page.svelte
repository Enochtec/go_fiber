<script lang="ts">
	import { onMount } from 'svelte';
	import { reportsService } from '$lib/services/reports';
	import { notify } from '$lib/stores/notification.svelte';
	import type { DailySalesRow, TopProductRow, InventoryValueRow } from '$lib/types';
	import { BarChart2 } from '@lucide/svelte';

	let dailySales = $state<DailySalesRow[]>([]);
	let topProducts = $state<TopProductRow[]>([]);
	let inventoryValue = $state<InventoryValueRow[]>([]);
	let loading = $state(true);

	let activeTab = $state<'sales' | 'products' | 'inventory'>('sales');

	function fmt(n: number) {
		return new Intl.NumberFormat('en-KE', { style: 'currency', currency: 'KES' }).format(n);
	}

	function fmtDate(s: string) {
		const d = new Date(s + 'T00:00:00');
		return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	const maxSales = $derived(Math.max(...dailySales.map((r) => r.total), 1));
	const totalRevenue = $derived(dailySales.reduce((s, r) => s + r.total, 0));
	const totalOrders = $derived(dailySales.reduce((s, r) => s + r.orders, 0));

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

<div class="p-6 space-y-5">
	<h1 class="text-xl font-semibold text-gray-900">Reports</h1>

	<div class="flex gap-2 border-b">
		{#each [['sales', 'Daily Sales'], ['products', 'Top Products'], ['inventory', 'Inventory Value']] as [tab, label]}
			<button
				onclick={() => activeTab = tab as typeof activeTab}
				class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px"
				class:border-blue-600={activeTab === tab}
				class:text-blue-600={activeTab === tab}
				class:border-transparent={activeTab !== tab}
				class:text-gray-500={activeTab !== tab}
			>
				{label}
			</button>
		{/each}
	</div>

	{#if loading}
		<div class="rounded-xl border bg-white p-8">
			<div class="space-y-3">
				{#each Array(8) as _}
					<div class="h-5 bg-gray-100 rounded animate-pulse"></div>
				{/each}
			</div>
		</div>
	{:else if activeTab === 'sales'}
		<div class="grid grid-cols-3 gap-4">
			<div class="rounded-xl border bg-white p-5">
				<p class="text-sm text-gray-500">30-Day Revenue</p>
				<p class="text-2xl font-bold text-gray-900 mt-1">{fmt(totalRevenue)}</p>
			</div>
			<div class="rounded-xl border bg-white p-5">
				<p class="text-sm text-gray-500">30-Day Orders</p>
				<p class="text-2xl font-bold text-gray-900 mt-1">{totalOrders}</p>
			</div>
			<div class="rounded-xl border bg-white p-5">
				<p class="text-sm text-gray-500">Avg. Order Value</p>
				<p class="text-2xl font-bold text-gray-900 mt-1">{fmt(totalOrders > 0 ? totalRevenue / totalOrders : 0)}</p>
			</div>
		</div>
		<div class="rounded-xl border bg-white p-6">
			<h2 class="text-sm font-semibold text-gray-900 mb-4">Daily Sales — Last 30 Days</h2>
			{#if dailySales.length === 0}
				<p class="text-sm text-gray-400 text-center py-8">No sales data</p>
			{:else}
				<div class="space-y-2">
					{#each [...dailySales].reverse() as row}
						<div class="flex items-center gap-4 text-sm">
							<span class="w-20 text-gray-500 shrink-0">{fmtDate(row.date)}</span>
							<div class="flex-1 bg-gray-100 rounded-full h-2">
								<div class="bg-blue-500 h-2 rounded-full" style="width: {(row.total / maxSales) * 100}%"></div>
							</div>
							<span class="w-24 text-right font-medium">{fmt(row.total)}</span>
							<span class="w-12 text-right text-gray-400">{row.orders}x</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>

	{:else if activeTab === 'products'}
		<div class="rounded-xl border bg-white overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b bg-gray-50 text-left">
						<th class="px-4 py-3 font-medium text-gray-600">#</th>
						<th class="px-4 py-3 font-medium text-gray-600">Product</th>
						<th class="px-4 py-3 font-medium text-gray-600 text-right">Qty Sold</th>
						<th class="px-4 py-3 font-medium text-gray-600 text-right">Revenue</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-100">
					{#if topProducts.length === 0}
						<tr><td colspan="4" class="px-4 py-12 text-center text-gray-400">No data</td></tr>
					{:else}
						{#each topProducts as p, i}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-3 text-gray-400 font-medium">{i + 1}</td>
								<td class="px-4 py-3 font-medium text-gray-900">{p.product_name}</td>
								<td class="px-4 py-3 text-right text-gray-700">{p.quantity}</td>
								<td class="px-4 py-3 text-right font-semibold">{fmt(p.revenue)}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>

	{:else}
		<div class="rounded-xl border bg-white overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b bg-gray-50 text-left">
						<th class="px-4 py-3 font-medium text-gray-600">Category</th>
						<th class="px-4 py-3 font-medium text-gray-600 text-right">Products</th>
						<th class="px-4 py-3 font-medium text-gray-600 text-right">Cost Value</th>
						<th class="px-4 py-3 font-medium text-gray-600 text-right">Retail Value</th>
						<th class="px-4 py-3 font-medium text-gray-600 text-right">Potential Profit</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-100">
					{#if inventoryValue.length === 0}
						<tr><td colspan="5" class="px-4 py-12 text-center text-gray-400">No data</td></tr>
					{:else}
						{#each inventoryValue as row}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-3 font-medium text-gray-900">{row.category_name}</td>
								<td class="px-4 py-3 text-right text-gray-600">{row.product_count}</td>
								<td class="px-4 py-3 text-right text-gray-600">{fmt(row.total_cost)}</td>
								<td class="px-4 py-3 text-right font-semibold">{fmt(row.total_value)}</td>
								<td class="px-4 py-3 text-right text-green-600 font-medium">{fmt(row.total_value - row.total_cost)}</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	{/if}
</div>
