<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { inventoryService } from '$lib/services/inventory';
	import { reportsService } from '$lib/services/reports';
	import { salesService } from '$lib/services/sales';
	import { productsService } from '$lib/services/products';
	import { notify } from '$lib/stores/notification.svelte';
	import type { DashboardStats, DailySalesRow, TopProductRow, Sale, Product } from '$lib/types';
	import {
		TrendingUp, TrendingDown, ShoppingCart, Package, AlertTriangle,
		Users, DollarSign, RefreshCw, Banknote, Smartphone, CreditCard, BarChart2, Clock
	} from '@lucide/svelte';

	let stats = $state<DashboardStats | null>(null);
	let dailySales = $state<DailySalesRow[]>([]);
	let topProducts = $state<TopProductRow[]>([]);
	let recentSales = $state<Sale[]>([]);
	let lowStockProducts = $state<Product[]>([]);
	let loading = $state(true);
	let refreshing = $state(false);
	let lastRefreshed = $state<Date | null>(null);

	const MAX_BAR = 140;
	const MAX_REV = $derived(dailySales.length > 0 ? Math.max(...dailySales.map(d => d.total)) : 0);
	const MAX_PROD = $derived(topProducts.length > 0 ? Math.max(...topProducts.map(p => p.revenue)) : 0);

	function fmt(n: number) {
		if (n >= 1_000_000) return `KES ${(n/1_000_000).toFixed(1)}M`;
		if (n >= 1_000) return `KES ${(n/1_000).toFixed(1)}K`;
		return `KES ${n.toFixed(0)}`;
	}
	function fmtFull(n: number) {
		return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}
	function fmtDate(s: string) {
		return new Date(s).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
	}
	function fmtTime(s: string) {
		return new Date(s).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}
	function pct(a: number, b: number) {
		if (!b) return 0;
		return Math.round(((a - b) / b) * 100);
	}

	async function loadData() {
		refreshing = true;
		try {
			const [statsRes, salesRes, topRes, recentRes, lowRes] = await Promise.all([
				inventoryService.dashboard(),
				reportsService.dailySales(7),
				reportsService.topProducts(8),
				salesService.list({ status: 'completed', page: 1, limit: 8 }),
				productsService.list({ low_stock: true, limit: 10 })
			]);
			stats = statsRes.data ?? null;
			dailySales = salesRes.data ?? [];
			topProducts = topRes.data ?? [];
			recentSales = recentRes.data ?? [];
			lowStockProducts = lowRes.data ?? [];
			lastRefreshed = new Date();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to load dashboard');
		} finally {
			loading = false;
			refreshing = false;
		}
	}

	let interval: ReturnType<typeof setInterval>;
	onMount(async () => {
		await loadData();
		interval = setInterval(loadData, 30_000);
	});
	onDestroy(() => clearInterval(interval));

	const salesTrend = $derived(stats ? pct(stats.today_sales, stats.yesterday_sales) : 0);
	const payTotal = $derived(stats ? stats.today_cash_sales + stats.today_mpesa + stats.today_card : 0);
</script>

<svelte:head><title>Dashboard — POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-5 min-h-full bg-slate-50 dark:bg-slate-950">

	<!-- Header -->
	<div class="flex items-center justify-between gap-3">
		<div>
			<h1 class="text-xl font-bold text-slate-900 dark:text-slate-100">Dashboard</h1>
			<p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
				{new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}
			</p>
		</div>
		<div class="flex items-center gap-2">
			{#if lastRefreshed}
				<span class="text-xs text-slate-400 hidden sm:block">Updated {lastRefreshed.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
			{/if}
			<button
				onclick={loadData}
				disabled={refreshing}
				class="flex h-8 w-8 items-center justify-center rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-500 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors disabled:opacity-50"
			>
				<RefreshCw size={14} class={refreshing ? 'animate-spin' : ''} />
			</button>
		</div>
	</div>

	{#if loading}
		<!-- Skeletons -->
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
			{#each Array(6) as _}
				<div class="rounded-2xl bg-slate-200 dark:bg-slate-700 h-24 animate-pulse"></div>
			{/each}
		</div>
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
			{#each Array(3) as _}
				<div class="rounded-2xl bg-slate-200 dark:bg-slate-700 h-64 animate-pulse {_ === 0 ? 'lg:col-span-2' : ''}"></div>
			{/each}
		</div>
	{:else if stats}

		<!-- ── Stat cards ────────────────────────────────────── -->
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">

			<!-- Today's Revenue -->
			<div class="lg:col-span-2 relative overflow-hidden rounded-2xl p-4 shadow-sm flex flex-col justify-between text-white" style="background:linear-gradient(135deg,#0ea5e9,#2563eb);">
				<div class="absolute -top-4 -right-4 h-24 w-24 rounded-full bg-white/10"></div>
				<div class="absolute -bottom-6 -left-6 h-32 w-32 rounded-full bg-white/5"></div>
				<div class="flex items-center justify-between relative">
					<span class="text-xs font-semibold uppercase tracking-wide text-white/80">Today's Revenue</span>
					<div class="h-7 w-7 rounded-lg flex items-center justify-center bg-white/20">
						<DollarSign size={14} class="text-white" />
					</div>
				</div>
				<div class="mt-2 relative">
					<p class="text-2xl font-bold tabular-nums">{fmt(stats.today_sales)}</p>
					<div class="flex items-center gap-1.5 mt-1">
						{#if salesTrend >= 0}
							<TrendingUp size={12} class="text-white/80" />
							<span class="text-xs font-semibold text-white/80">+{salesTrend}% vs yesterday</span>
						{:else}
							<TrendingDown size={12} class="text-white/80" />
							<span class="text-xs font-semibold text-white/80">{salesTrend}% vs yesterday</span>
						{/if}
					</div>
				</div>
			</div>

			<!-- Orders -->
			<div class="relative overflow-hidden rounded-2xl p-4 shadow-sm text-white" style="background:linear-gradient(135deg,#8b5cf6,#6366f1);">
				<div class="absolute -top-3 -right-3 h-16 w-16 rounded-full bg-white/10"></div>
				<div class="absolute -bottom-4 -left-4 h-20 w-20 rounded-full bg-white/5"></div>
				<div class="flex items-center justify-between relative">
					<span class="text-xs font-semibold uppercase tracking-wide text-white/80">Orders</span>
					<div class="h-7 w-7 rounded-lg flex items-center justify-center bg-white/20">
						<ShoppingCart size={14} class="text-white" />
					</div>
				</div>
				<div class="mt-2 relative">
					<p class="text-2xl font-bold tabular-nums">{stats.today_orders}</p>
					<p class="text-xs text-white/70 mt-1">transactions today</p>
				</div>
			</div>

			<!-- Avg Sale -->
			<div class="relative overflow-hidden rounded-2xl p-4 shadow-sm text-white" style="background:linear-gradient(135deg,#f59e0b,#d97706);">
				<div class="absolute -top-3 -right-3 h-16 w-16 rounded-full bg-white/10"></div>
				<div class="absolute -bottom-4 -left-4 h-20 w-20 rounded-full bg-white/5"></div>
				<div class="flex items-center justify-between relative">
					<span class="text-xs font-semibold uppercase tracking-wide text-white/80">Avg Sale</span>
					<div class="h-7 w-7 rounded-lg flex items-center justify-center bg-white/20">
						<BarChart2 size={14} class="text-white" />
					</div>
				</div>
				<div class="mt-2 relative">
					<p class="text-2xl font-bold tabular-nums">{fmt(stats.today_avg_sale)}</p>
					<p class="text-xs text-white/70 mt-1">per transaction</p>
				</div>
			</div>

			<!-- Products -->
			<div class="relative overflow-hidden rounded-2xl p-4 shadow-sm text-white" style="background:linear-gradient(135deg,#10b981,#059669);">
				<div class="absolute -top-3 -right-3 h-16 w-16 rounded-full bg-white/10"></div>
				<div class="absolute -bottom-4 -left-4 h-20 w-20 rounded-full bg-white/5"></div>
				<div class="flex items-center justify-between relative">
					<span class="text-xs font-semibold uppercase tracking-wide text-white/80">Products</span>
					<div class="h-7 w-7 rounded-lg flex items-center justify-center bg-white/20">
						<Package size={14} class="text-white" />
					</div>
				</div>
				<div class="mt-2 relative">
					<p class="text-2xl font-bold tabular-nums">{stats.total_products}</p>
					<p class="text-xs text-white/70 mt-1">active items</p>
				</div>
			</div>

			<!-- Stock alerts -->
			<div class="relative overflow-hidden rounded-2xl p-4 shadow-sm text-white" style="background:linear-gradient(135deg,#ef4444,#dc2626);">
				<div class="absolute -top-3 -right-3 h-16 w-16 rounded-full bg-white/10"></div>
				<div class="absolute -bottom-4 -left-4 h-20 w-20 rounded-full bg-white/5"></div>
				<div class="flex items-center justify-between relative">
					<span class="text-xs font-semibold uppercase tracking-wide text-white/80">Stock Alerts</span>
					<div class="h-7 w-7 rounded-lg flex items-center justify-center bg-white/20">
						<AlertTriangle size={14} class="text-white" />
					</div>
				</div>
				<div class="mt-2 relative">
					<p class="text-2xl font-bold tabular-nums">{stats.low_stock_count + stats.out_of_stock}</p>
					<p class="text-xs text-white/70 mt-1">{stats.out_of_stock} out of stock</p>
				</div>
			</div>
		</div>

		<!-- ── Charts row ────────────────────────────────────── -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">

			<!-- 7-day bar chart -->
			<div class="lg:col-span-2 rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-5 shadow-sm">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Revenue — Last 7 Days</h2>
					<span class="text-xs text-slate-400">{fmt(stats.month_sales)} this month</span>
				</div>
				{#if dailySales.length === 0}
					<div class="flex items-center justify-center h-36 text-slate-400 text-sm">No sales data yet</div>
				{:else}
					<div class="flex items-end gap-2 h-36 mt-2">
						{#each dailySales as day}
							{@const h = MAX_REV > 0 ? Math.max((day.total / MAX_REV) * MAX_BAR, 4) : 4}
							<div class="group flex-1 flex flex-col items-center gap-1 h-full justify-end">
								<div class="relative w-full">
									<div
										class="absolute -top-6 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-lg bg-slate-900 dark:bg-slate-600 px-2 py-0.5 text-[10px] font-semibold text-white opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10"
									>{fmt(day.total)}</div>
									<div
										class="w-full rounded-lg transition-all"
										style="height:{h}px; background-color:#008B8B; opacity:{day.total > 0 ? 0.8 : 0.15};"
									></div>
								</div>
								<p class="text-[9px] text-slate-400 mt-1 text-center leading-tight">{fmtDate(day.date).split(' ').slice(0,2).join(' ')}</p>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Payment method breakdown -->
			<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-5 shadow-sm">
				<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100 mb-4">Payments — Today</h2>
				{#if payTotal === 0}
					<div class="flex items-center justify-center h-36 text-slate-400 text-sm">No payments yet</div>
				{:else}
					<div class="space-y-3.5">
						{#each [
							{ label: 'Cash', value: stats.today_cash_sales, icon: Banknote, color: '#10b981' },
							{ label: 'M-Pesa', value: stats.today_mpesa, icon: Smartphone, color: '#008B8B' },
							{ label: 'Card', value: stats.today_card, icon: CreditCard, color: '#6366f1' }
						] as pm}
							{@const w = payTotal > 0 ? (pm.value / payTotal) * 100 : 0}
							<div>
								<div class="flex items-center justify-between mb-1.5">
									<div class="flex items-center gap-1.5">
										<svelte:component this={pm.icon} size={12} style="color:{pm.color};" />
										<span class="text-xs font-semibold text-slate-600 dark:text-slate-300">{pm.label}</span>
									</div>
									<span class="text-xs font-bold text-slate-700 dark:text-slate-200 tabular-nums">{fmt(pm.value)}</span>
								</div>
								<div class="h-2 rounded-full bg-slate-100 dark:bg-slate-700 overflow-hidden">
									<div class="h-full rounded-full transition-all" style="width:{w}%; background-color:{pm.color};"></div>
								</div>
								<p class="text-right text-[10px] text-slate-400 mt-0.5">{w.toFixed(0)}%</p>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- ── Bottom row ────────────────────────────────────── -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">

			<!-- Top products -->
			<div class="lg:col-span-2 rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-5 shadow-sm">
				<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100 mb-4">Top Products — 30 Days</h2>
				{#if topProducts.length === 0}
					<div class="flex items-center justify-center h-32 text-slate-400 text-sm">No sales data yet</div>
				{:else}
					<div class="space-y-2.5">
						{#each topProducts.slice(0, 6) as product, i}
							{@const w = MAX_PROD > 0 ? (product.revenue / MAX_PROD) * 100 : 0}
							<div class="flex items-center gap-3">
								<span class="text-xs font-bold text-slate-300 dark:text-slate-600 w-4 shrink-0 text-right">{i+1}</span>
								<div class="flex-1 min-w-0">
									<div class="flex items-center justify-between mb-1">
										<span class="text-xs font-semibold text-slate-700 dark:text-slate-200 truncate">{product.product_name}</span>
										<div class="flex items-center gap-2 shrink-0 ml-2">
											<span class="text-xs text-slate-400">{product.quantity_sold} sold</span>
											<span class="text-xs font-bold text-slate-800 dark:text-slate-100 tabular-nums">{fmt(product.revenue)}</span>
										</div>
									</div>
									<div class="h-1.5 rounded-full bg-slate-100 dark:bg-slate-700 overflow-hidden">
										<div class="h-full rounded-full" style="width:{w}%; background-color:#008B8B;"></div>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Low stock + recent sales -->
			<div class="space-y-4">

				<!-- Low stock alerts -->
				<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4 shadow-sm">
					<div class="flex items-center gap-2 mb-3">
						<AlertTriangle size={14} class="text-amber-500 shrink-0" />
						<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Low Stock</h2>
						{#if lowStockProducts.length > 0}
							<span class="ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-amber-100 text-amber-700 text-[10px] font-bold px-1">{lowStockProducts.length}</span>
						{/if}
					</div>
					{#if lowStockProducts.length === 0}
						<p class="text-xs text-slate-400 py-3 text-center">All products are well-stocked</p>
					{:else}
						<ul class="space-y-1.5">
							{#each lowStockProducts.slice(0, 5) as p}
								<li class="flex items-center justify-between">
									<span class="text-xs text-slate-600 dark:text-slate-300 truncate">{p.name}</span>
									<span class="shrink-0 ml-2 inline-flex rounded-full px-2 py-0.5 text-[10px] font-bold {p.stock_qty === 0 ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'}">{p.stock_qty === 0 ? 'Out' : p.stock_qty}</span>
								</li>
							{/each}
						</ul>
					{/if}
				</div>

				<!-- Recent transactions -->
				<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4 shadow-sm">
					<div class="flex items-center gap-2 mb-3">
						<Clock size={14} class="text-slate-400 shrink-0" />
						<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Recent Sales</h2>
					</div>
					{#if recentSales.length === 0}
						<p class="text-xs text-slate-400 py-3 text-center">No sales today yet</p>
					{:else}
						<ul class="space-y-2">
							{#each recentSales.slice(0, 5) as sale}
								<li class="flex items-center justify-between gap-2">
									<div class="min-w-0">
										<p class="text-xs font-semibold text-slate-700 dark:text-slate-200 truncate">#{sale.id.slice(0,6).toUpperCase()}</p>
										<p class="text-[10px] text-slate-400">{fmtTime(sale.created_at)}</p>
									</div>
									<div class="text-right shrink-0">
										<p class="text-xs font-bold text-slate-800 dark:text-slate-100 tabular-nums">KES {fmtFull(sale.total)}</p>
										<p class="text-[10px] text-slate-400 capitalize">{sale.payment_method}</p>
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			</div>
		</div>

	{/if}
</div>
