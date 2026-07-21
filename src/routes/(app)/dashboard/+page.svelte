<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { inventoryService } from '$lib/services/inventory';
	import { reportsService } from '$lib/services/reports';
	import { offlineSales as salesService } from '$lib/services/offline';
	import { offlineProducts as productsService } from '$lib/services/offline';
	import { notify } from '$lib/stores/notification.svelte';
	import type { DashboardStats, DailySalesRow, TopProductRow, Product, InventoryValueRow } from '$lib/types';
	import {
		TrendingUp, TrendingDown, ShoppingCart, Package, AlertTriangle,
		Users, DollarSign, RefreshCw,
		Activity, Layers, Target, Zap
	} from '@lucide/svelte';
	import LineAreaChart from '$lib/components/charts/LineAreaChart.svelte';
	import BarChart      from '$lib/components/charts/BarChart.svelte';
	import DonutChart    from '$lib/components/charts/DonutChart.svelte';
	import HBarChart     from '$lib/components/charts/HBarChart.svelte';

	let stats = $state<DashboardStats | null>(null);
	let dailySales = $state<DailySalesRow[]>([]);
	let monthlySales = $state<DailySalesRow[]>([]);
	let topProducts = $state<TopProductRow[]>([]);
	let inventoryValue = $state<InventoryValueRow[]>([]);
	let lowStockProducts = $state<Product[]>([]);
	let loading = $state(true);
	let refreshing = $state(false);
	let lastRefreshed = $state<Date | null>(null);



	function fmt(n: number) {
		if (n >= 1_000_000) return `KES ${(n/1_000_000).toFixed(1)}M`;
		if (n >= 1_000) return `KES ${(n/1_000).toFixed(1)}K`;
		return `KES ${n.toFixed(0)}`;
	}
	function fmtCompact(n: number) {
		if (n >= 1_000_000) return `${(n/1_000_000).toFixed(1)}M`;
		if (n >= 1_000) return `${(n/1_000).toFixed(1)}K`;
		return n.toFixed(0);
	}
	function fmtFull(n: number) {
		return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}
	function fmtDate(s: string) {
		return new Date(s).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
	}
	function fmtMonth(s: string) {
		const d = new Date(s + '-01');
		return d.toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
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
			const [statsRes, salesRes, monthlyRes, topRes, invValRes, lowRes] = await Promise.all([
				inventoryService.dashboard(),
				reportsService.dailySales(7),
				reportsService.monthlySales(12),
				reportsService.topProducts(8),
				reportsService.inventoryValue(),
				productsService.list({ low_stock: true, limit: 10 })
			]);
			stats = statsRes.data ?? null;
			dailySales = salesRes.data ?? [];
			monthlySales = monthlyRes.data ?? [];
			topProducts = topRes.data ?? [];
			inventoryValue = invValRes.data ?? [];
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

	// ── Core derived values ─────────────────────────────────────────────────────
	const salesTrend      = $derived(stats ? pct(stats.today_sales, stats.yesterday_sales) : 0);
	const payTotal        = $derived(stats ? stats.today_cash_sales + stats.today_mpesa + stats.today_card : 0);
	const healthyProducts = $derived(stats ? stats.total_products - stats.low_stock_count - stats.out_of_stock : 0);
	const healthPct       = $derived(stats && stats.total_products > 0 ? (healthyProducts / stats.total_products) * 100 : 0);
	const totalInvValue   = $derived(inventoryValue.reduce((s, r) => s + r.total_value, 0));
	const totalInvCost    = $derived(inventoryValue.reduce((s, r) => s + r.total_cost, 0));
	const invMargin       = $derived(totalInvValue > 0 ? ((totalInvValue - totalInvCost) / totalInvValue) * 100 : 0);
	const monthlyGrowth   = $derived(monthlySales.length >= 2 ? pct(monthlySales[0].total, monthlySales[1].total) : 0);
	const avgDaily        = $derived(dailySales.length > 0 ? dailySales.reduce((s, d) => s + d.total, 0) / dailySales.length : 0);
	const thisWeekTotal   = $derived(dailySales.reduce((s, d) => s + d.total, 0));
	const weekPace        = $derived(stats?.month_sales ? Math.min((thisWeekTotal / (stats.month_sales / 4)) * 100, 100) : 0);

	// ── Chart datasets ───────────────────────────────────────────────────────────
	const lineLabels   = $derived(dailySales.map(d => fmtDate(d.date).split(' ').slice(0, 2).join(' ')));
	const lineDatasets = $derived([{ label: 'Revenue', data: dailySales.map(d => d.total), color: '#6366f1' }]);

	const barLabels  = $derived([...monthlySales].reverse().map(m => fmtMonth(m.date)));
	const barData    = $derived([...monthlySales].reverse().map(m => m.total));

	const payLabels  = ['Cash', 'M-Pesa', 'Card'];
	const payColors  = ['#6366f1', '#10b981', '#8b5cf6'];
	const payData    = $derived(stats ? [stats.today_cash_sales, stats.today_mpesa, stats.today_card] : [0, 0, 0]);
	const payCenter  = $derived(fmt(payTotal));

	const stockLabels = ['Healthy', 'Low Stock', 'Out of Stock'];
	const stockColors = ['#10b981', '#f59e0b', '#ef4444'];
	const stockData   = $derived(stats ? [healthyProducts, stats.low_stock_count, stats.out_of_stock] : [0, 0, 0]);
	const stockCenter = $derived(`${healthPct.toFixed(0)}%`);

	const invLabels  = $derived(inventoryValue.map(v => v.category_name));
	const invColors  = ['#6366f1', '#10b981', '#8b5cf6', '#f59e0b', '#ef4444', '#0ea5e9', '#ec4899'];
	const invData    = $derived(inventoryValue.map(v => v.total_value));

	const hBarLabels = $derived(topProducts.slice(0, 7).map(p => p.product_name));
	const hBarData   = $derived(topProducts.slice(0, 7).map(p => p.revenue));
	const hBarSub    = $derived(topProducts.slice(0, 7).map(p => p.quantity_sold));

	// ── Mini SVG sparklines ──────────────────────────────────────────────────────
	function sparkPath(values: number[], w = 80, h = 28): string {
		if (values.length < 2) return '';
		const max = Math.max(...values, 1);
		return values.map((v, i) => {
			const x = (i / (values.length - 1)) * w;
			const y = h - 4 - ((v / max) * (h - 8));
			return `${x},${y}`;
		}).join(' ');
	}
	const sparkRev    = $derived(sparkPath(dailySales.map(d => d.total)));
	const sparkOrders = $derived(sparkPath(dailySales.map(d => d.orders)));
</script>

<svelte:head><title>Dashboard — Maestro POS</title></svelte:head>

<div class="min-h-full bg-slate-50 dark:bg-slate-950 p-4 md:p-6 space-y-6">

	<!-- ── Header ─────────────────────────────────────────────────────────── -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">Business Overview</h1>
			<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
				{new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}
			</p>
		</div>
		<div class="flex items-center gap-3">
			{#if lastRefreshed}
				<span class="text-xs text-slate-400 hidden sm:block">
					Updated {lastRefreshed.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
				</span>
			{/if}
			<button
				onclick={loadData}
				disabled={refreshing}
				class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors disabled:opacity-50"
			>
				<RefreshCw size={12} class={refreshing ? 'animate-spin' : ''} />
				Refresh
			</button>
		</div>
	</div>

	{#if loading}
		<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
			{#each Array(4) as _}
				<div class="h-28 bg-slate-200 dark:bg-slate-700 animate-pulse rounded"></div>
			{/each}
		</div>
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
			<div class="h-80 bg-slate-200 dark:bg-slate-700 animate-pulse rounded lg:col-span-2"></div>
			<div class="h-80 bg-slate-200 dark:bg-slate-700 animate-pulse rounded"></div>
		</div>
	{:else if stats}

		<!-- ── KPI Cards ─────────────────────────────────────────────────── -->
		<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">

			<!-- Today Revenue -->
			<div class="relative overflow-hidden bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
				<div class="absolute inset-y-0 left-0 w-1 bg-indigo-500"></div>
				<div class="flex items-start justify-between pl-2">
					<div class="flex-1 min-w-0">
						<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-widest">Today Revenue</p>
						<p class="text-xl font-bold text-slate-900 dark:text-slate-100 mt-1.5 tracking-tight tabular-nums truncate">{fmt(stats.today_sales)}</p>
						<div class="flex items-center gap-1 mt-2">
							{#if salesTrend >= 0}
								<span class="inline-flex items-center text-[10px] font-bold text-emerald-700 bg-emerald-50 dark:bg-emerald-900/30 dark:text-emerald-400 px-1.5 py-0.5 rounded-full">▲ +{salesTrend}%</span>
							{:else}
								<span class="inline-flex items-center text-[10px] font-bold text-red-600 bg-red-50 dark:bg-red-900/30 dark:text-red-400 px-1.5 py-0.5 rounded-full">▼ {salesTrend}%</span>
							{/if}
							<span class="text-[10px] text-slate-400">vs yesterday</span>
						</div>
					</div>
					<div class="flex flex-col items-end gap-2 shrink-0 ml-2">
						<div class="h-9 w-9 flex items-center justify-center bg-indigo-50 dark:bg-indigo-900/30 rounded">
							<DollarSign size={17} class="text-indigo-600 dark:text-indigo-400" />
						</div>
						{#if sparkRev}
							<svg viewBox="0 0 80 28" class="w-20 h-7 text-indigo-400 opacity-80" preserveAspectRatio="none">
								<polyline points={sparkRev} fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
							</svg>
						{/if}
					</div>
				</div>
			</div>

			<!-- Orders Today -->
			<div class="relative overflow-hidden bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
				<div class="absolute inset-y-0 left-0 w-1 bg-emerald-500"></div>
				<div class="flex items-start justify-between pl-2">
					<div class="flex-1 min-w-0">
						<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-widest">Orders Today</p>
						<p class="text-xl font-bold text-slate-900 dark:text-slate-100 mt-1.5 tracking-tight tabular-nums">{stats.today_orders}</p>
						<p class="text-[10px] text-slate-400 mt-2">{fmt(stats.today_avg_sale)} avg value</p>
					</div>
					<div class="flex flex-col items-end gap-2 shrink-0 ml-2">
						<div class="h-9 w-9 flex items-center justify-center bg-emerald-50 dark:bg-emerald-900/30 rounded">
							<ShoppingCart size={17} class="text-emerald-600 dark:text-emerald-400" />
						</div>
						{#if sparkOrders}
							<svg viewBox="0 0 80 28" class="w-20 h-7 text-emerald-400 opacity-80" preserveAspectRatio="none">
								<polyline points={sparkOrders} fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
							</svg>
						{/if}
					</div>
				</div>
			</div>

			<!-- Month Revenue -->
			<div class="relative overflow-hidden bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
				<div class="absolute inset-y-0 left-0 w-1 bg-violet-500"></div>
				<div class="flex items-start justify-between pl-2">
					<div class="flex-1 min-w-0">
						<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-widest">Month Revenue</p>
						<p class="text-xl font-bold text-slate-900 dark:text-slate-100 mt-1.5 tracking-tight tabular-nums truncate">{fmt(stats.month_sales)}</p>
						<div class="flex items-center gap-1 mt-2">
							{#if monthlyGrowth >= 0}
								<span class="inline-flex items-center text-[10px] font-bold text-emerald-700 bg-emerald-50 dark:bg-emerald-900/30 dark:text-emerald-400 px-1.5 py-0.5 rounded-full">▲ +{monthlyGrowth}%</span>
							{:else}
								<span class="inline-flex items-center text-[10px] font-bold text-red-600 bg-red-50 dark:bg-red-900/30 dark:text-red-400 px-1.5 py-0.5 rounded-full">▼ {monthlyGrowth}%</span>
							{/if}
							<span class="text-[10px] text-slate-400">MoM</span>
						</div>
					</div>
					<div class="h-9 w-9 shrink-0 ml-2 flex items-center justify-center bg-violet-50 dark:bg-violet-900/30 rounded">
						<TrendingUp size={17} class="text-violet-600 dark:text-violet-400" />
					</div>
				</div>
			</div>

			<!-- Customers + Alerts -->
			<div class="relative overflow-hidden bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
				<div class="absolute inset-y-0 left-0 w-1 bg-amber-500"></div>
				<div class="flex items-start justify-between pl-2">
					<div class="flex-1 min-w-0">
						<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-widest">Customers</p>
						<p class="text-xl font-bold text-slate-900 dark:text-slate-100 mt-1.5 tracking-tight tabular-nums">{stats.total_customers}</p>
						<div class="flex items-center gap-1.5 mt-2">
							<AlertTriangle size={10} class="text-red-500 shrink-0" />
							<span class="text-[10px] font-semibold text-red-500">{stats.low_stock_count + stats.out_of_stock}</span>
							<span class="text-[10px] text-slate-400">stock alerts</span>
						</div>
					</div>
					<div class="h-9 w-9 shrink-0 ml-2 flex items-center justify-center bg-amber-50 dark:bg-amber-900/30 rounded">
						<Users size={17} class="text-amber-600 dark:text-amber-400" />
					</div>
				</div>
			</div>
		</div>

		<!-- ── Row 2: 7-Day Line Chart + Payment Donut ─────────────────── -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">

			<!-- 7-Day Area Line Chart -->
			<div class="lg:col-span-2 bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
				<div class="flex items-start justify-between mb-5">
					<div>
						<h2 class="text-sm font-bold text-slate-800 dark:text-slate-100">Revenue Trend — Last 7 Days</h2>
						<p class="text-xs text-slate-400 mt-0.5">Daily revenue with smooth area fill</p>
					</div>
					<div class="flex items-center gap-1.5 text-xs text-slate-400 shrink-0">
						<span class="w-4 h-0.5 bg-indigo-500 rounded inline-block"></span> Revenue
					</div>
				</div>
				{#if dailySales.length === 0}
					<div class="flex items-center justify-center h-56 text-slate-400 text-sm">No sales data yet</div>
				{:else}
					<LineAreaChart labels={lineLabels} datasets={lineDatasets} height={220} />
					{@const bestDay = dailySales.reduce((a, b) => a.total > b.total ? a : b)}
					<div class="grid grid-cols-3 gap-3 mt-5 pt-4 border-t border-slate-100 dark:border-slate-700">
						<div>
							<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide">Best Day</p>
							<p class="text-xs font-bold text-slate-800 dark:text-slate-100 mt-0.5">{fmtDate(bestDay.date)}</p>
							<p class="text-[10px] font-bold text-indigo-600 mt-0.5">{fmt(bestDay.total)}</p>
						</div>
						<div>
							<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide">Daily Average</p>
							<p class="text-xs font-bold text-slate-800 dark:text-slate-100 mt-0.5">{fmt(avgDaily)}</p>
						</div>
						<div>
							<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide">Week Total</p>
							<p class="text-xs font-bold text-slate-800 dark:text-slate-100 mt-0.5">{fmt(thisWeekTotal)}</p>
						</div>
					</div>
				{/if}
			</div>

			<!-- Payment Method Donut -->
			<div class="bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
				<h2 class="text-sm font-bold text-slate-800 dark:text-slate-100">Payment Split</h2>
				<p class="text-xs text-slate-400 mt-0.5 mb-4">Today's revenue by method</p>
				{#if payTotal === 0}
					<div class="flex items-center justify-center h-56 text-slate-400 text-sm">No payments today</div>
				{:else}
					<DonutChart labels={payLabels} data={payData} colors={payColors} cutout="0" height={190} />
					<div class="mt-5 space-y-3">
						{#each [
							{ label: 'Cash',   value: stats.today_cash_sales, color: '#6366f1' },
							{ label: 'M-Pesa', value: stats.today_mpesa,       color: '#10b981' },
							{ label: 'Card',   value: stats.today_card,        color: '#8b5cf6' },
						] as pm}
							{@const p = payTotal > 0 ? ((pm.value / payTotal) * 100).toFixed(1) : '0'}
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<span class="h-2.5 w-2.5 rounded-full shrink-0" style="background:{pm.color}"></span>
									<span class="text-xs text-slate-600 dark:text-slate-300">{pm.label}</span>
								</div>
								<div class="text-right">
									<span class="text-xs font-bold text-slate-800 dark:text-slate-100 tabular-nums">{fmt(pm.value)}</span>
									<span class="text-[10px] text-slate-400 ml-1">{p}%</span>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- ── Row 3: 12-Month Bar Chart + Performance ─────────────────── -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">

			<!-- 12-Month Bar Chart -->
			<div class="lg:col-span-2 bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
				<div class="flex items-start justify-between mb-5">
					<div>
						<h2 class="text-sm font-bold text-slate-800 dark:text-slate-100">Monthly Revenue — Last 12 Months</h2>
						<p class="text-xs text-slate-400 mt-0.5">Peak month highlighted in solid indigo</p>
					</div>
					{#if monthlySales.length >= 2}
						<span class="inline-flex items-center gap-1 text-xs font-bold px-2.5 py-1 rounded-full shrink-0 {monthlyGrowth >= 0 ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-red-50 text-red-600 dark:bg-red-900/30 dark:text-red-400'}">
							{monthlyGrowth >= 0 ? `▲ +${monthlyGrowth}%` : `▼ ${monthlyGrowth}%`}
						</span>
					{/if}
				</div>
				{#if monthlySales.length === 0}
					<div class="flex items-center justify-center h-56 text-slate-400 text-sm">No historical data yet</div>
				{:else}
					<BarChart labels={barLabels} data={barData} height={235} />
					{@const highest = monthlySales.reduce((a, b) => a.total > b.total ? a : b)}
					<div class="grid grid-cols-3 gap-3 mt-5 pt-4 border-t border-slate-100 dark:border-slate-700">
						<div>
							<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide">Best Month</p>
							<p class="text-xs font-bold text-slate-800 dark:text-slate-100 mt-0.5">{fmtMonth(highest.date)}</p>
							<p class="text-[10px] font-bold text-indigo-600 mt-0.5">{fmt(highest.total)}</p>
						</div>
						<div>
							<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide">Monthly Avg</p>
							<p class="text-xs font-bold text-slate-800 dark:text-slate-100 mt-0.5">{fmt(monthlySales.reduce((s, m) => s + m.total, 0) / monthlySales.length)}</p>
						</div>
						<div>
							<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide">12-Month Total</p>
							<p class="text-xs font-bold text-slate-800 dark:text-slate-100 mt-0.5">{fmt(monthlySales.reduce((s, m) => s + m.total, 0))}</p>
						</div>
					</div>
				{/if}
			</div>

			<!-- Performance Panel -->
			<div class="bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
				<div class="flex items-center gap-2 mb-5">
					<Activity size={14} class="text-indigo-600" />
					<h2 class="text-sm font-bold text-slate-800 dark:text-slate-100">Performance</h2>
				</div>
				<div class="space-y-5">
					<div>
						<div class="flex items-center justify-between mb-2">
							<div class="flex items-center gap-1.5"><Zap size={11} class="text-indigo-500"/><span class="text-xs font-semibold text-slate-600 dark:text-slate-300">MoM Growth</span></div>
							<span class="text-sm font-bold {monthlyGrowth >= 0 ? 'text-emerald-600' : 'text-red-500'}">{monthlyGrowth >= 0 ? '+' : ''}{monthlyGrowth}%</span>
						</div>
						<div class="h-2 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
							<div class="h-full rounded-full {monthlyGrowth >= 0 ? 'bg-emerald-500' : 'bg-red-500'}" style="width:{Math.min(Math.abs(monthlyGrowth), 100)}%"></div>
						</div>
					</div>
					<div>
						<div class="flex items-center justify-between mb-2">
							<div class="flex items-center gap-1.5"><Layers size={11} class="text-violet-500"/><span class="text-xs font-semibold text-slate-600 dark:text-slate-300">Gross Margin</span></div>
							<span class="text-sm font-bold text-violet-600">{invMargin.toFixed(1)}%</span>
						</div>
						<div class="h-2 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
							<div class="h-full bg-violet-500 rounded-full" style="width:{Math.min(invMargin, 100)}%"></div>
						</div>
				</div>
				<div>
					<div class="flex items-center justify-between mb-2">
						<div class="flex items-center gap-1.5"><Target size={11} class="text-amber-500"/><span class="text-xs font-semibold text-slate-600 dark:text-slate-300">Weekly Pace</span></div>
							<span class="text-sm font-bold text-amber-600">{weekPace.toFixed(0)}%</span>
						</div>
						<div class="h-2 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
							<div class="h-full bg-amber-500 rounded-full" style="width:{weekPace}%"></div>
						</div>
						<p class="text-[10px] text-slate-400 mt-1">of expected monthly pace</p>
					</div>
					<div class="border-t border-slate-100 dark:border-slate-700 pt-4 grid grid-cols-2 gap-2.5">
						{#each [
							{ label: 'Avg Daily',    value: fmt(avgDaily),                  cls: 'text-slate-800 dark:text-slate-100' },
							{ label: 'Products',     value: String(stats.total_products),    cls: 'text-slate-800 dark:text-slate-100' },
							{ label: 'Low Stock',    value: String(stats.low_stock_count),   cls: 'text-amber-600' },
							{ label: 'Out of Stock', value: String(stats.out_of_stock),      cls: 'text-red-500' },
						] as s}
							<div class="bg-slate-50 dark:bg-slate-900/40 p-3">
								<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide">{s.label}</p>
								<p class="text-sm font-bold {s.cls} mt-0.5">{s.value}</p>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>

		<!-- ── Row 4: Top Products HBar + Stock Donut + Inventory Donut ── -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">

			<!-- Top Products Horizontal Bar Chart -->
			<div class="lg:col-span-2 bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
				<div class="flex items-start justify-between mb-5">
					<div>
						<h2 class="text-sm font-bold text-slate-800 dark:text-slate-100">Top Products — Last 30 Days</h2>
						<p class="text-xs text-slate-400 mt-0.5">Revenue &amp; units sold side-by-side</p>
					</div>
					{#if topProducts.length > 0}
						<span class="text-xs font-semibold text-slate-500 shrink-0">{topProducts.reduce((s, p) => s + p.quantity_sold, 0)} units total</span>
					{/if}
				</div>
				{#if topProducts.length === 0}
					<div class="flex items-center justify-center h-48 text-slate-400 text-sm">No sales data yet</div>
				{:else}
					<HBarChart labels={hBarLabels} data={hBarData} subData={hBarSub} label="Revenue (KES)" subLabel="Units Sold" height={260} />
				{/if}
			</div>

			<!-- Right column -->
			<div class="space-y-4">

				<!-- Stock Health Donut -->
				<div class="bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
					<h2 class="text-sm font-bold text-slate-800 dark:text-slate-100 mb-1">Stock Health</h2>
					<p class="text-xs text-slate-400 mb-3">Product availability breakdown</p>
					{#if stats.total_products === 0}
						<p class="text-xs text-slate-400 py-4 text-center">No products</p>
					{:else}
						<DonutChart labels={stockLabels} data={stockData} colors={stockColors} centerLabel="Healthy" centerValue={stockCenter} height={160} />
						<div class="mt-4 space-y-2">
							{#each [
								{ label: 'Healthy',      value: healthyProducts,       color: '#10b981' },
								{ label: 'Low Stock',    value: stats.low_stock_count,  color: '#f59e0b' },
								{ label: 'Out of Stock', value: stats.out_of_stock,     color: '#ef4444' },
							] as row}
								<div class="flex items-center justify-between">
									<div class="flex items-center gap-2">
										<span class="h-2.5 w-2.5 rounded-full shrink-0" style="background:{row.color}"></span>
										<span class="text-xs text-slate-600 dark:text-slate-300">{row.label}</span>
									</div>
									<span class="text-xs font-bold text-slate-800 dark:text-slate-100">{row.value}</span>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Inventory Value Donut -->
				{#if inventoryValue.length > 0}
				<div class="bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-5 shadow-sm">
					<h2 class="text-sm font-bold text-slate-800 dark:text-slate-100 mb-1">Inventory by Category</h2>
					<p class="text-xs text-slate-400 mb-3">Retail value distribution</p>
					<DonutChart labels={invLabels} data={invData} colors={invColors} centerLabel="Total" centerValue={fmt(totalInvValue)} height={160} />
					<div class="mt-4 pt-3 border-t border-slate-100 dark:border-slate-700 flex items-center justify-between">
						<div>
							<p class="text-[10px] text-slate-400 uppercase tracking-wide">Cost</p>
							<p class="text-xs font-bold text-slate-800 dark:text-slate-100">{fmt(totalInvCost)}</p>
						</div>
						<div class="text-right">
							<p class="text-[10px] text-slate-400 uppercase tracking-wide">Margin</p>
							<p class="text-xs font-bold text-emerald-600">{invMargin.toFixed(1)}%</p>
						</div>
					</div>
				</div>
				{/if}
			</div>
		</div>

	{/if}
</div>
