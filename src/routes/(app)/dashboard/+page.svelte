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
		interval = setInterval(loadData, 60_000);
	});
	onDestroy(() => clearInterval(interval));

	const salesTrend = $derived(stats ? pct(stats.today_sales, stats.yesterday_sales) : 0);
	const payTotal = $derived(stats ? stats.today_cash_sales + stats.today_mpesa + stats.today_card : 0);
	const healthyProducts = $derived(stats ? stats.total_products - stats.low_stock_count - stats.out_of_stock : 0);
	const healthPct = $derived(stats && stats.total_products > 0 ? (healthyProducts / stats.total_products) * 100 : 0);
	const lowPct = $derived(stats && stats.total_products > 0 ? (stats.low_stock_count / stats.total_products) * 100 : 0);
	const outPct = $derived(stats && stats.total_products > 0 ? (stats.out_of_stock / stats.total_products) * 100 : 0);
</script>

<svelte:head><title>Dashboard — Maestro POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-5 min-h-full bg-slate-50 dark:bg-slate-950">

	<!-- Page header -->
	<div class="flex items-center justify-between gap-3 pl-3 border-l-4 border-blue-500">
		<div>
			<h1 class="text-lg font-bold text-slate-900 dark:text-slate-100">Dashboard</h1>
			<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
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
				class="flex h-8 w-8 items-center justify-center bg-white dark:bg-slate-800 text-slate-500 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors disabled:opacity-50"
			>
				<RefreshCw size={13} class={refreshing ? 'animate-spin' : ''} />
			</button>
		</div>
	</div>

	{#if loading}
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
			{#each Array(6) as _}
				<div class="bg-slate-200 dark:bg-slate-700 h-20 animate-pulse"></div>
			{/each}
		</div>
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
			{#each [2, 1] as span}
				<div class="bg-slate-200 dark:bg-slate-700 h-56 animate-pulse lg:col-span-{span}"></div>
			{/each}
		</div>
	{:else if stats}

		<!-- ── KPI tiles ─── -->
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">

			<!-- Revenue (wider) — Primary -->
			<div class="lg:col-span-2 relative overflow-hidden p-4 text-white" style="background:linear-gradient(135deg,#3F00FF,#3200CC);">
				<div class="absolute -top-6 -right-6 h-28 w-28 bg-white/10"></div>
				<div class="absolute -bottom-8 -left-8 h-32 w-32 bg-white/5"></div>
				<div class="flex items-start justify-between relative">
					<div>
						<span class="text-xs font-semibold uppercase tracking-wider text-white/80">Today's Revenue</span>
						<p class="text-2xl font-bold mt-1 tracking-tight">{fmt(stats.today_sales)}</p>
					</div>
					<div class="flex h-10 w-10 items-center justify-center bg-white/20">
						<TrendingUp size={20} class="text-white" />
					</div>
				</div>
				<div class="flex items-center gap-1 mt-2 relative">
					{#if salesTrend >= 0}
						<span class="text-xs font-medium text-white/90">▲ +{salesTrend}% vs yesterday</span>
					{:else}
						<span class="text-xs font-medium text-white/90">▼ {salesTrend}% vs yesterday</span>
					{/if}
				</div>
			</div>

			<!-- Orders — Secondary -->
			<div class="relative overflow-hidden p-4 text-white" style="background:linear-gradient(135deg,#7B68EE,#6A5ACD);">
				<div class="absolute -top-6 -right-6 h-24 w-24 bg-white/10"></div>
				<div class="absolute -bottom-6 -left-6 h-20 w-20 bg-white/5"></div>
				<div class="flex items-start justify-between relative">
					<div>
						<span class="text-xs font-semibold uppercase tracking-wider text-white/80">Orders</span>
						<p class="text-2xl font-bold mt-1 tracking-tight">{stats.today_orders}</p>
					</div>
					<div class="flex h-10 w-10 items-center justify-center bg-white/20">
						<ShoppingCart size={20} class="text-white" />
					</div>
				</div>
				<p class="text-xs text-white/70 mt-2 relative">transactions today</p>
			</div>

			<!-- Avg Sale — Gold -->
			<div class="relative overflow-hidden p-4 text-white" style="background:linear-gradient(135deg,#FFD700,#DAA520);">
				<div class="absolute -top-6 -right-6 h-24 w-24 bg-white/10"></div>
				<div class="absolute -bottom-6 -left-6 h-20 w-20 bg-white/5"></div>
				<div class="flex items-start justify-between relative">
					<div>
						<span class="text-xs font-semibold uppercase tracking-wider text-white/80">Avg Sale</span>
						<p class="text-2xl font-bold mt-1 tracking-tight">{fmt(stats.today_avg_sale)}</p>
					</div>
					<div class="flex h-10 w-10 items-center justify-center bg-white/20">
						<BarChart2 size={20} class="text-white" />
					</div>
				</div>
				<p class="text-xs text-white/70 mt-2 relative">per transaction</p>
			</div>

			<!-- Products — Neon Green -->
			<div class="relative overflow-hidden p-4 text-white" style="background:linear-gradient(135deg,#39FF14,#2ECC40);">
				<div class="absolute -top-6 -right-6 h-24 w-24 bg-white/10"></div>
				<div class="absolute -bottom-6 -left-6 h-20 w-20 bg-white/5"></div>
				<div class="flex items-start justify-between relative">
					<div>
						<span class="text-xs font-semibold uppercase tracking-wider text-white/80">Products</span>
						<p class="text-2xl font-bold mt-1 tracking-tight">{stats.total_products}</p>
					</div>
					<div class="flex h-10 w-10 items-center justify-center bg-white/20">
						<Package size={20} class="text-white" />
					</div>
				</div>
				<p class="text-xs text-white/70 mt-2 relative">active items</p>
			</div>

			<!-- Stock Alerts — Red -->
			<div class="relative overflow-hidden p-4 text-white" style="background:linear-gradient(135deg,#FF2400,#CC0000);">
				<div class="absolute -top-6 -right-6 h-24 w-24 bg-white/10"></div>
				<div class="absolute -bottom-6 -left-6 h-20 w-20 bg-white/5"></div>
				<div class="flex items-start justify-between relative">
					<div>
						<span class="text-xs font-semibold uppercase tracking-wider text-white/80">Stock Alerts</span>
						<p class="text-2xl font-bold mt-1 tracking-tight">{stats.low_stock_count + stats.out_of_stock}</p>
					</div>
					<div class="flex h-10 w-10 items-center justify-center bg-white/20">
						<AlertTriangle size={20} class="text-white" />
					</div>
				</div>
				<p class="text-xs text-white/70 mt-2 relative">{stats.out_of_stock} out of stock</p>
			</div>
		</div>

		<!-- ── Revenue chart + Payment breakdown ── -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">

			<!-- 7-day bar chart -->
			<div class="lg:col-span-2 bg-white dark:bg-slate-800 p-5">
				<div class="flex items-center justify-between mb-5">
					<div>
						<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Revenue — Last 7 Days</h2>
						<p class="text-xs text-slate-400 mt-0.5">{fmt(stats.month_sales)} this month</p>
					</div>
				</div>
				{#if dailySales.length === 0}
					<div class="flex items-center justify-center h-36 text-slate-400 text-sm">No sales data yet</div>
				{:else}
					{@const bestDay = dailySales.reduce((a,b) => a.total > b.total ? a : b)}
					{@const dailyAvg = dailySales.reduce((s,d) => s + d.total, 0) / dailySales.length}
					<div class="flex items-end gap-2 h-36">
						{#each dailySales as day}
							{@const h = MAX_REV > 0 ? Math.max((day.total / MAX_REV) * MAX_BAR, 3) : 3}
							<div class="group flex-1 flex flex-col items-center gap-1 h-full justify-end">
								<div class="relative w-full">
									<div class="absolute -top-6 left-1/2 -translate-x-1/2 whitespace-nowrap rounded bg-slate-900 px-1.5 py-0.5 text-[10px] font-semibold text-white opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
										{fmt(day.total)}
									</div>
									<div class="w-full transition-all bg-blue-600 hover:bg-blue-700"
										style="height:{h}px; opacity:{day.total > 0 ? 1 : 0.12};"
									></div>
								</div>
								<p class="text-[9px] text-slate-400 text-center leading-tight">{fmtDate(day.date).split(' ').slice(0,2).join(' ')}</p>
							</div>
						{/each}
					</div>
					<div class="grid grid-cols-3 gap-3 mt-4 pt-4 border-t border-slate-100 dark:border-slate-700">
						<div>
							<p class="text-[10px] text-slate-400 uppercase tracking-wide">Best Day</p>
							<p class="text-xs font-bold text-slate-700 dark:text-slate-200 mt-0.5">{fmtDate(bestDay.date)}</p>
							<p class="text-[10px] font-semibold text-emerald-600">{fmt(bestDay.total)}</p>
						</div>
						<div>
							<p class="text-[10px] text-slate-400 uppercase tracking-wide">Daily Avg</p>
							<p class="text-xs font-bold text-slate-700 dark:text-slate-200 mt-0.5">{fmt(dailyAvg)}</p>
						</div>
						<div>
							<p class="text-[10px] text-slate-400 uppercase tracking-wide">vs Yesterday</p>
							<p class="text-xs font-bold {salesTrend >= 0 ? 'text-emerald-600' : 'text-red-500'} mt-0.5">{salesTrend >= 0 ? '+' : ''}{salesTrend}%</p>
						</div>
					</div>
				{/if}
			</div>

			<!-- Payment breakdown -->
			<div class="bg-white dark:bg-slate-800 p-5">
				<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100 mb-4">Payments Today</h2>
				{#if payTotal === 0}
					<div class="flex items-center justify-center h-36 text-slate-400 text-sm">No payments yet</div>
				{:else}
					{@const cashPct  = (stats.today_cash_sales / payTotal) * 100}
					{@const mpesaPct = (stats.today_mpesa     / payTotal) * 100}
					{@const cardPct  = (stats.today_card      / payTotal) * 100}
					{@const grad = `conic-gradient(#3F00FF 0% ${cashPct}%, #10b981 ${cashPct}% ${cashPct+mpesaPct}%, #8b5cf6 ${cashPct+mpesaPct}% 100%)`}
					<div class="flex flex-col gap-5">
						<div class="flex justify-center">
							<div class="relative h-28 w-28 rounded-full" style="background:{grad};">
								<div class="absolute inset-[10px] rounded-full bg-white dark:bg-slate-800 flex flex-col items-center justify-center">
									<p class="text-base font-bold text-slate-800 dark:text-slate-100">{fmt(payTotal)}</p>
									<p class="text-[9px] text-slate-400 uppercase tracking-wide">total</p>
								</div>
							</div>
						</div>
						<div class="space-y-2.5">
							{#each [
								{ label: 'Cash',   value: stats.today_cash_sales, color: '#3F00FF', icon: Banknote   },
								{ label: 'M-Pesa', value: stats.today_mpesa,       color: '#10b981', icon: Smartphone },
								{ label: 'Card',   value: stats.today_card,        color: '#8b5cf6', icon: CreditCard },
							] as pm}
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-2 min-w-0">
										<span class="h-2 w-2 shrink-0 rounded-full" style="background:{pm.color};"></span>
										<span class="text-xs text-slate-600 dark:text-slate-300">{pm.label}</span>
									</div>
									<span class="text-xs font-semibold text-slate-800 dark:text-slate-100 tabular-nums">{fmt(pm.value)}</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>

		<!-- ── Top Products + Side panels ── -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">

			<!-- Top products -->
			<div class="lg:col-span-2 bg-white dark:bg-slate-800 p-5">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Top Products — 30 Days</h2>
					{#if topProducts.length > 0}
						<span class="text-xs text-slate-400">{topProducts.reduce((s,p)=>s+p.quantity_sold,0)} units sold</span>
					{/if}
				</div>
				{#if topProducts.length === 0}
					<div class="flex items-center justify-center h-32 text-slate-400 text-sm">No sales data yet</div>
				{:else}
					<div class="space-y-3">
						{#each topProducts.slice(0,6) as product, i}
							{@const revW = MAX_PROD > 0 ? (product.revenue / MAX_PROD) * 100 : 0}
							<div class="flex items-center gap-3">
								<span class="text-xs font-bold text-slate-300 dark:text-slate-600 w-4 shrink-0 text-right tabular-nums">{i+1}</span>
								<div class="flex-1 min-w-0">
									<div class="flex items-center justify-between mb-1">
										<span class="text-xs font-semibold text-slate-700 dark:text-slate-200 truncate">{product.product_name}</span>
										<div class="flex items-center gap-2 shrink-0 ml-2">
											<span class="text-[10px] text-slate-400">{product.quantity_sold} sold</span>
											<span class="text-xs font-bold text-slate-900 dark:text-slate-100 tabular-nums">{fmt(product.revenue)}</span>
										</div>
									</div>
									<div class="h-1.5 bg-slate-100 dark:bg-slate-700 overflow-hidden">
										<div class="h-full bg-blue-600" style="width:{revW}%;"></div>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Right column -->
			<div class="space-y-4">

				<!-- Quick stats grid -->
				<div class="bg-white dark:bg-slate-800 p-4">
					<h2 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-3">Catalogue</h2>
					<div class="grid grid-cols-2 gap-2">
						{#each [
							{ label: 'Active Products', value: stats.total_products, cls: 'text-blue-600 dark:text-blue-400' },
							{ label: 'Customers',        value: stats.total_customers, cls: 'text-violet-600 dark:text-violet-400' },
							{ label: 'Low Stock',        value: stats.low_stock_count,  cls: 'text-amber-600 dark:text-amber-400' },
							{ label: 'Out of Stock',     value: stats.out_of_stock,     cls: 'text-red-600 dark:text-red-400' },
						] as s}
							<div class="bg-slate-50 dark:bg-slate-900/40 p-3 text-center">
								<p class="text-lg font-bold {s.cls}">{s.value}</p>
								<p class="text-[10px] text-slate-500 mt-0.5 leading-tight">{s.label}</p>
							</div>
						{/each}
					</div>
				</div>

				<!-- Stock health -->
				<div class="bg-white dark:bg-slate-800 p-4">
					<h2 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-3">Stock Health</h2>
					<div class="flex items-center gap-4">
						<div class="relative h-20 w-20 shrink-0">
							<svg class="h-full w-full -rotate-90" viewBox="0 0 36 36">
								<circle cx="18" cy="18" r="15" fill="none" stroke="#e2e8f0" stroke-width="3" class="dark:stroke-slate-700"/>
								<circle cx="18" cy="18" r="15" fill="none" stroke="#22c55e" stroke-width="3" stroke-dasharray="{healthPct} {100-healthPct}" stroke-linecap="round"/>
								<circle cx="18" cy="18" r="15" fill="none" stroke="#f59e0b" stroke-width="3" stroke-dasharray="{lowPct} {100-lowPct}" stroke-dashoffset="{-healthPct}" stroke-linecap="round"/>
								<circle cx="18" cy="18" r="15" fill="none" stroke="#ef4444" stroke-width="3" stroke-dasharray="{outPct} {100-outPct}" stroke-dashoffset="{-(healthPct+lowPct)}" stroke-linecap="round"/>
							</svg>
							<div class="absolute inset-0 flex flex-col items-center justify-center">
								<p class="text-sm font-bold text-slate-800 dark:text-slate-100">{healthPct.toFixed(0)}%</p>
								<p class="text-[8px] text-slate-400 uppercase">OK</p>
							</div>
						</div>
						<div class="flex-1 space-y-1.5 text-xs">
							{#each [
								{ label: 'Healthy',      value: healthyProducts,      dot: 'bg-emerald-500' },
								{ label: 'Low Stock',    value: stats.low_stock_count, dot: 'bg-amber-500' },
								{ label: 'Out of Stock', value: stats.out_of_stock,    dot: 'bg-red-500' },
							] as row}
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-1.5">
										<span class="h-2 w-2 rounded-full {row.dot}"></span>
										<span class="text-slate-600 dark:text-slate-300">{row.label}</span>
									</div>
									<span class="font-bold text-slate-700 dark:text-slate-200">{row.value}</span>
								</div>
							{/each}
						</div>
					</div>
				</div>

				<!-- Recent sales -->
				<div class="bg-white dark:bg-slate-800 p-4">
					<div class="flex items-center gap-2 mb-3">
						<Clock size={13} class="text-slate-400" />
						<h2 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Recent Sales</h2>
					</div>
					{#if recentSales.length === 0}
						<p class="text-xs text-slate-400 py-4 text-center">No sales today yet</p>
					{:else}
						<ul class="space-y-2.5">
							{#each recentSales.slice(0,5) as sale}
								<li class="flex items-center justify-between gap-2">
									<div class="min-w-0">
										<p class="text-xs font-semibold text-slate-700 dark:text-slate-200">#{sale.id.slice(0,6).toUpperCase()}</p>
										<p class="text-[10px] text-slate-400">{fmtTime(sale.created_at)}</p>
									</div>
									<div class="text-right shrink-0">
										<p class="text-xs font-bold text-slate-900 dark:text-slate-100 tabular-nums">KES {fmtFull(sale.total)}</p>
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
