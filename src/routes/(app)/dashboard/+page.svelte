<script lang="ts">
	import { onMount } from 'svelte';
	import { inventoryService } from '$lib/services/inventory';
	import { reportsService } from '$lib/services/reports';
	import { notify } from '$lib/stores/notification.svelte';
	import type { DashboardStats, DailySalesRow, TopProductRow } from '$lib/types';
	import { TrendingUp, ShoppingCart, Package, AlertTriangle, Users, DollarSign } from '@lucide/svelte';

	let stats = $state<DashboardStats | null>(null);
	let dailySales = $state<DailySalesRow[]>([]);
	let topProducts = $state<TopProductRow[]>([]);
	let loading = $state(true);

	const MAX_BAR_HEIGHT = 140;
	const MAX_REV = $derived(dailySales.length > 0 ? Math.max(...dailySales.map(d => d.total)) : 0);
	const MAX_PROD = $derived(topProducts.length > 0 ? Math.max(...topProducts.map(p => p.revenue)) : 0);

	function fmt(n: number) {
		return new Intl.NumberFormat('en-KE', { style: 'currency', currency: 'KES' }).format(n);
	}

	function fmtDate(s: string) {
		return new Date(s).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	onMount(async () => {
		try {
			const [statsRes, salesRes, topRes] = await Promise.all([
				inventoryService.dashboard(),
				reportsService.dailySales(7),
				reportsService.topProducts(5)
			]);
			stats = statsRes.data ?? null;
			dailySales = salesRes.data ?? [];
			topProducts = topRes.data ?? [];
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to load dashboard');
		} finally {
			loading = false;
		}
	});

	const statCards = $derived(stats ? [
		{ label: "Today's Sales", value: fmt(stats.today_sales), sub: `${stats.today_orders} orders today`, icon: DollarSign, grad: 'from-blue-500 to-blue-700', iconBg: 'bg-blue-400/30' },
		{ label: 'This Month', value: fmt(stats.month_sales), sub: 'Monthly revenue', icon: TrendingUp, grad: 'from-emerald-500 to-emerald-700', iconBg: 'bg-emerald-400/30' },
		{ label: 'Products', value: stats.total_products.toString(), sub: 'Active items', icon: Package, grad: 'from-violet-500 to-violet-700', iconBg: 'bg-violet-400/30' },
		{ label: 'Low Stock', value: stats.low_stock_count.toString(), sub: 'Need reorder', icon: AlertTriangle, grad: 'bg-amber-500', iconBg: 'bg-amber-400/30' },
		{ label: 'Customers', value: stats.total_customers.toString(), sub: 'Registered', icon: Users, grad: 'from-cyan-500 to-cyan-700', iconBg: 'bg-cyan-400/30' }
	] : []);
</script>

<svelte:head><title>Dashboard — POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-slate-900 dark:text-slate-100">Dashboard</h1>
			<p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
				{new Date().toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
			</p>
		</div>
	</div>

	{#if loading}
		<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3 md:gap-4">
			{#each Array(5) as _}
				<div class="rounded-2xl bg-slate-200 h-28 animate-pulse"></div>
			{/each}
		</div>
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
			{#each Array(2) as _}
				<div class="rounded-2xl bg-slate-200 h-64 animate-pulse"></div>
			{/each}
		</div>
	{:else}
		<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3 md:gap-4">
			{#each statCards as card}
				<div class="relative overflow-hidden rounded-2xl {card.grad.includes('bg-') ? card.grad : 'bg-gradient-to-br ' + card.grad} p-5 text-white shadow-lg">
					<div class="flex items-start justify-between mb-4">
						<p class="text-xs font-medium text-white/80 leading-tight">{card.label}</p>
						<div class="rounded-xl p-2 {card.iconBg}"><card.icon size={15} class="text-white" /></div>
					</div>
					<p class="text-2xl font-bold tracking-tight">{card.value}</p>
					<p class="text-xs text-white/70 mt-1">{card.sub}</p>
					<div class="absolute -right-4 -bottom-4 h-20 w-20 rounded-full bg-white/10"></div>
				</div>
			{/each}
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
			<!-- Daily Sales - CSS bars -->
			<div class="rounded-2xl bg-white dark:bg-slate-800 p-5 shadow-sm border border-slate-100 dark:border-slate-700">
				<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-4">Daily Sales — 7 Days</h2>
				{#if dailySales.length === 0}
					<p class="text-sm text-slate-400 py-8 text-center">No data yet</p>
				{:else}
					<div class="flex items-end gap-2 h-40">
						{#each dailySales as day}
							<div class="flex-1 flex flex-col items-center gap-1 h-full justify-end">
								<p class="text-xs font-medium text-blue-600">{fmt(day.total)}</p>
								<div
									class="w-full rounded-t-md bg-blue-500 hover:bg-blue-600 transition-colors"
									style="height: {MAX_REV > 0 ? (day.total / MAX_REV) * MAX_BAR_HEIGHT : 0}px; min-height: 4px"
								></div>
								<p class="text-xs text-slate-400 mt-1">{fmtDate(day.date)}</p>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Monthly Revenue - CSS bars -->
			<div class="rounded-2xl bg-white dark:bg-slate-800 p-5 shadow-sm border border-slate-100 dark:border-slate-700">
				<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-4">Top Products — 30 Days</h2>
				{#if topProducts.length === 0}
					<p class="text-sm text-slate-400 py-8 text-center">No data yet</p>
				{:else}
					<div class="space-y-3">
						{#each topProducts as product}
							<div>
								<div class="flex justify-between text-sm mb-1">
									<span class="text-slate-700 font-medium truncate">{product.product_name}</span>
									<span class="text-slate-500 shrink-0 ml-2">{fmt(product.revenue)}</span>
								</div>
								<div class="w-full bg-slate-100 rounded-full h-2.5 overflow-hidden">
									<div
										class="h-full rounded-full bg-blue-500 transition-all"
										style="width: {MAX_PROD > 0 ? (product.revenue / MAX_PROD) * 100 : 0}%"
									></div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
