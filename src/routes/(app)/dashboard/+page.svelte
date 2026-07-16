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
	const healthyProducts = $derived(stats ? stats.total_products - stats.low_stock_count - stats.out_of_stock : 0);
	const healthPct = $derived(stats && stats.total_products > 0 ? (healthyProducts / stats.total_products) * 100 : 0);
	const lowPct = $derived(stats && stats.total_products > 0 ? (stats.low_stock_count / stats.total_products) * 100 : 0);
	const outPct = $derived(stats && stats.total_products > 0 ? (stats.out_of_stock / stats.total_products) * 100 : 0);
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
					{@const bestDay = dailySales.reduce((a,b) => a.total > b.total ? a : b)}
					{@const dailyAvg = dailySales.reduce((s,d) => s + d.total, 0) / dailySales.length}
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
					<!-- Mini stats under chart -->
					<div class="grid grid-cols-3 gap-3 mt-4 pt-4 border-t border-slate-100 dark:border-slate-700">
						<div class="text-center">
							<p class="text-[10px] text-slate-400 uppercase tracking-wide">Best Day</p>
							<p class="text-xs font-bold text-slate-700 dark:text-slate-200 mt-0.5">{fmtDate(bestDay.date)}</p>
							<p class="text-[10px] font-semibold text-emerald-600">{fmt(bestDay.total)}</p>
						</div>
						<div class="text-center">
							<p class="text-[10px] text-slate-400 uppercase tracking-wide">Daily Avg</p>
							<p class="text-xs font-bold text-slate-700 dark:text-slate-200 mt-0.5">{fmt(dailyAvg)}</p>
						</div>
						<div class="text-center">
							<p class="text-[10px] text-slate-400 uppercase tracking-wide">vs Yesterday</p>
							<p class="text-xs font-bold {salesTrend >= 0 ? 'text-emerald-600' : 'text-red-500'} mt-0.5">{salesTrend >= 0 ? '+' : ''}{salesTrend}%</p>
						</div>
					</div>
				{/if}
			</div>

			<!-- Payment method breakdown - Donut chart -->
			<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-5 shadow-sm">
				<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100 mb-4">Payments — Today</h2>
				{#if payTotal === 0}
					<div class="flex items-center justify-center h-36 text-slate-400 text-sm">No payments yet</div>
				{:else}
					{@const cashPct = (stats.today_cash_sales / payTotal) * 100}
					{@const mpesaPct = (stats.today_mpesa / payTotal) * 100}
					{@const cardPct = (stats.today_card / payTotal) * 100}
					{@const grad = `conic-gradient(#10b981 0% ${cashPct}%, #008B8B ${cashPct}% ${cashPct + mpesaPct}%, #6366f1 ${cashPct + mpesaPct}% 100%)`}
					<div class="flex items-center gap-5">
						<div class="relative h-28 w-28 shrink-0 rounded-full" style="background:{grad};">
							<div class="absolute inset-[10px] rounded-full bg-white dark:bg-slate-800 flex flex-col items-center justify-center">
								<p class="text-lg font-bold text-slate-800 dark:text-slate-100">{fmt(payTotal)}</p>
								<p class="text-[9px] text-slate-400 uppercase tracking-wide">total</p>
							</div>
						</div>
						<div class="flex-1 space-y-2.5">
							{#each [
								{ label: 'Cash', value: stats.today_cash_sales, color: '#10b981', pct: cashPct },
								{ label: 'M-Pesa', value: stats.today_mpesa, color: '#008B8B', pct: mpesaPct },
								{ label: 'Card', value: stats.today_card, color: '#6366f1', pct: cardPct }
							] as pm}
								<div class="flex items-center justify-between gap-2">
									<div class="flex items-center gap-1.5 min-w-0">
										<span class="h-2.5 w-2.5 shrink-0 rounded-full" style="background:{pm.color};"></span>
										<span class="text-xs text-slate-600 dark:text-slate-300 truncate">{pm.label}</span>
									</div>
									<div class="flex items-center gap-2 shrink-0">
										<span class="text-xs font-bold text-slate-700 dark:text-slate-200">{fmt(pm.value)}</span>
										<span class="text-[10px] text-slate-400 w-9 text-right tabular-nums">{pm.pct.toFixed(0)}%</span>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>

		<!-- ── Bottom row: Product Analysis ──────────────────── -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">

			<!-- Top products with revenue bars -->
			<div class="lg:col-span-2 rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-5 shadow-sm">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Top Products — 30 Days</h2>
					{#if topProducts.length > 0}
						{@const totalSold = topProducts.reduce((s,p) => s + p.quantity_sold, 0)}
						<span class="text-xs text-slate-400">{totalSold} total sold</span>
					{/if}
				</div>
				{#if topProducts.length === 0}
					<div class="flex items-center justify-center h-32 text-slate-400 text-sm">No sales data yet</div>
				{:else}
					<div class="space-y-3">
						{#each topProducts.slice(0, 6) as product, i}
							{@const revW = MAX_PROD > 0 ? (product.revenue / MAX_PROD) * 100 : 0}
							{@const qtyW = topProducts[0].quantity_sold > 0 ? (product.quantity_sold / topProducts[0].quantity_sold) * 100 : 0}
							<div class="flex items-center gap-3">
								<span class="text-xs font-bold text-slate-300 dark:text-slate-600 w-4 shrink-0 text-right">{i+1}</span>
								<div class="flex-1 min-w-0">
									<div class="flex items-center justify-between mb-0.5">
										<span class="text-xs font-semibold text-slate-700 dark:text-slate-200 truncate">{product.product_name}</span>
										<div class="flex items-center gap-2 shrink-0 ml-2">
											<span class="text-xs text-slate-400">{product.quantity_sold} sold</span>
											<span class="text-xs font-bold text-slate-800 dark:text-slate-100 tabular-nums">{fmt(product.revenue)}</span>
										</div>
									</div>
									<div class="flex items-center gap-2">
										<div class="flex-1 h-1.5 rounded-full bg-slate-100 dark:bg-slate-700 overflow-hidden">
											<div class="h-full rounded-full" style="width:{revW}%; background-color:#008B8B;"></div>
										</div>
										<span class="text-[9px] text-slate-400 w-8 text-right tabular-nums">{revW.toFixed(0)}%</span>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Right column: Product Stats + Stock Health -->
			<div class="space-y-4">

				<!-- Product performance stats -->
				<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4 shadow-sm">
					<div class="flex items-center gap-2 mb-3">
						<Package size={14} class="text-slate-400 shrink-0" />
						<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Product Stats</h2>
					</div>
					<div class="grid grid-cols-2 gap-3">
						<div class="rounded-xl bg-emerald-50 dark:bg-emerald-900/20 p-3 text-center">
							<p class="text-lg font-bold text-emerald-700 dark:text-emerald-400">{stats.total_products}</p>
							<p class="text-[10px] text-emerald-600/70 dark:text-emerald-400/70">Active Products</p>
						</div>
						<div class="rounded-xl bg-blue-50 dark:bg-blue-900/20 p-3 text-center">
							<p class="text-lg font-bold text-blue-700 dark:text-blue-400">{stats.total_customers}</p>
							<p class="text-[10px] text-blue-600/70 dark:text-blue-400/70">Customers</p>
						</div>
						<div class="rounded-xl bg-amber-50 dark:bg-amber-900/20 p-3 text-center">
							<p class="text-lg font-bold text-amber-700 dark:text-amber-400">{stats.low_stock_count}</p>
							<p class="text-[10px] text-amber-600/70 dark:text-amber-400/70">Low Stock</p>
						</div>
						<div class="rounded-xl bg-red-50 dark:bg-red-900/20 p-3 text-center">
							<p class="text-lg font-bold text-red-700 dark:text-red-400">{stats.out_of_stock}</p>
							<p class="text-[10px] text-red-600/70 dark:text-red-400/70">Out of Stock</p>
						</div>
					</div>
				</div>

				<!-- Stock health gauge -->
				<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4 shadow-sm">
					<div class="flex items-center gap-2 mb-3">
						<BarChart2 size={14} class="text-slate-400 shrink-0" />
						<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Stock Health</h2>
					</div>
					<div class="flex items-center gap-4">
						<div class="relative h-24 w-24 shrink-0">
							<svg class="h-full w-full -rotate-90" viewBox="0 0 36 36">
								<circle cx="18" cy="18" r="15" fill="none" stroke="#e2e8f0" stroke-width="3" class="dark:stroke-slate-700"/>
								<circle cx="18" cy="18" r="15" fill="none" stroke="#10b981" stroke-width="3" stroke-dasharray="{healthPct} {100 - healthPct}" stroke-linecap="round"/>
								<circle cx="18" cy="18" r="15" fill="none" stroke="#f59e0b" stroke-width="3" stroke-dasharray="{lowPct} {100 - lowPct}" stroke-dashoffset="{-healthPct}" stroke-linecap="round"/>
								<circle cx="18" cy="18" r="15" fill="none" stroke="#ef4444" stroke-width="3" stroke-dasharray="{outPct} {100 - outPct}" stroke-dashoffset="{-(healthPct + lowPct)}" stroke-linecap="round"/>
							</svg>
							<div class="absolute inset-0 flex flex-col items-center justify-center">
								<p class="text-lg font-bold text-slate-800 dark:text-slate-100">{healthPct.toFixed(0)}%</p>
								<p class="text-[8px] text-slate-400 uppercase tracking-wide">healthy</p>
							</div>
						</div>
						<div class="flex-1 space-y-1.5">
							<div class="flex items-center justify-between text-xs">
								<div class="flex items-center gap-1.5">
									<span class="h-2 w-2 rounded-full bg-emerald-500"></span>
									<span class="text-slate-600 dark:text-slate-300">Healthy</span>
								</div>
								<span class="font-bold text-slate-700 dark:text-slate-200">{healthyProducts}</span>
							</div>
							<div class="flex items-center justify-between text-xs">
								<div class="flex items-center gap-1.5">
									<span class="h-2 w-2 rounded-full bg-amber-500"></span>
									<span class="text-slate-600 dark:text-slate-300">Low Stock</span>
								</div>
								<span class="font-bold text-slate-700 dark:text-slate-200">{stats.low_stock_count}</span>
							</div>
							<div class="flex items-center justify-between text-xs">
								<div class="flex items-center gap-1.5">
									<span class="h-2 w-2 rounded-full bg-red-500"></span>
									<span class="text-slate-600 dark:text-slate-300">Out of Stock</span>
								</div>
								<span class="font-bold text-slate-700 dark:text-slate-200">{stats.out_of_stock}</span>
							</div>
						</div>
					</div>
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
