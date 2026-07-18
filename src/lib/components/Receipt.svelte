<script lang="ts">
	import { onMount } from 'svelte';
	import { Printer, Download, X, Phone } from '@lucide/svelte';
	import { shopService, type ShopInfo } from '$lib/services/shop';
	import type { Sale } from '$lib/types';

	let {
		sale,
		amountTendered,
		change,
		pointsPhone = '',
		onclose,
		onnewsale,
	}: {
		sale: Sale;
		amountTendered: number;
		change: number;
		pointsPhone?: string;
		onclose: () => void;
		onnewsale: () => void;
	} = $props();

	let shopInfo = $state<ShopInfo | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			shopInfo = await shopService.getInfo();
		} catch {}
		loading = false;
	});

	function fmt(n: number) {
		return new Intl.NumberFormat('en-KE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}

	function formatPhone(num: string) {
		const d = num.replace(/\D/g, '');
		if (d.startsWith('0')) return '254' + d.slice(1);
		if (d.startsWith('254')) return d;
		if (d.startsWith('+')) return d.slice(1);
		return d;
	}

	function sendWhatsApp() {
		const num = pointsPhone.trim();
		if (!num) return;
		const shop = shopInfo?.shop;
		const items = (sale.items ?? []).map(i =>
			`${i.product_name ?? 'Item'} x${i.quantity}  KES ${fmt(i.total)}`
		).join('\n');
		const msg = [
			`*${shop?.name || 'Maestro POS'}*`,
			`Receipt: #${sale.id.slice(0, 8).toUpperCase()}`,
			`Date: ${new Date(sale.created_at).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })}`,
			``,
			`${items}`,
			``,
			`Total: KES ${fmt(sale.total)}`,
			`Payment: ${sale.payment_method.toUpperCase()}`,
			``,
			`${shopInfo?.settings?.receipt_footer || 'Thank you for your business!'}`
		].join('\n');
		window.open(`https://wa.me/${formatPhone(num)}?text=${encodeURIComponent(msg)}`, '_blank');
	}

	function print() {
		window.print();
	}

	function downloadPDF() {
		window.print();
	}
</script>

<!-- Main overlay -->
<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
	<div class="relative w-full max-w-lg max-h-[96vh] flex flex-col rounded-2xl bg-white dark:bg-slate-800 shadow-2xl overflow-hidden">
		<!-- Receipt body (scrollable, printable) -->
		<div class="flex-1 overflow-y-auto receipt-print">
			<div class="bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100">
				<!-- Receipt content -->
				<div class="px-6 py-5">
					<!-- Header -->
					<div class="text-center mb-5">
						{#if shopInfo?.shop?.logo}
							<img src={shopInfo.shop.logo} alt="Logo" class="h-12 w-12 object-contain mx-auto mb-2 rounded" />
						{/if}
						<h1 class="text-lg font-bold tracking-tight text-slate-900 dark:text-slate-100">{shopInfo?.shop?.name || 'Maestro POS'}</h1>
						{#if shopInfo?.shop?.address}
							<p class="text-[10px] text-slate-500 mt-0.5">{shopInfo.shop.address}</p>
						{/if}
						<div class="flex items-center justify-center gap-3 text-[10px] text-slate-400 mt-1">
							{#if shopInfo?.shop?.phone}<span>{shopInfo.shop.phone}</span>{/if}
							{#if shopInfo?.shop?.email}<span>{shopInfo.shop.email}</span>{/if}
						</div>
					</div>

					<!-- Receipt header info -->
					<div class="border-y border-slate-200 dark:border-slate-600 py-2.5 mb-4 text-[10px]">
						<div class="flex justify-between text-slate-500">
							<span>Receipt</span>
							<span class="font-mono font-bold text-slate-700 dark:text-slate-300">#{sale.id.slice(0, 10).toUpperCase()}</span>
						</div>
						<div class="flex justify-between text-slate-500 mt-1">
							<span>Date</span>
							<span class="text-slate-700 dark:text-slate-300">{new Date(sale.created_at).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })}</span>
						</div>
						<div class="flex justify-between text-slate-500 mt-1">
							<span>Time</span>
							<span class="text-slate-700 dark:text-slate-300">{new Date(sale.created_at).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })}</span>
						</div>
						<div class="flex justify-between text-slate-500 mt-1">
							<span>Cashier</span>
							<span class="text-slate-700 dark:text-slate-300">{sale.cashier_name || '—'}</span>
						</div>
						{#if sale.customer_name}
							<div class="flex justify-between text-slate-500 mt-1">
								<span>Customer</span>
								<span class="text-slate-700 dark:text-slate-300">{sale.customer_name}</span>
							</div>
						{/if}
					</div>

					<!-- Items table -->
					<table class="w-full text-xs mb-4">
						<thead>
							<tr class="border-b border-slate-200 dark:border-slate-600">
								<th class="pb-1.5 text-left font-semibold text-slate-400 w-1/2">Item</th>
								<th class="pb-1.5 text-center font-semibold text-slate-400 w-[15%]">Qty</th>
								<th class="pb-1.5 text-right font-semibold text-slate-400 w-[15%]">Price</th>
								<th class="pb-1.5 text-right font-semibold text-slate-400 w-[20%]">Total</th>
							</tr>
						</thead>
						<tbody>
							{#each sale.items || [] as item}
								<tr class="border-b border-slate-50 dark:border-slate-700/50">
									<td class="py-1.5 pr-2 text-slate-800 dark:text-slate-200 font-medium leading-tight">{item.product_name ?? '—'}</td>
									<td class="py-1.5 text-center text-slate-500 tabular-nums">{item.quantity}</td>
									<td class="py-1.5 text-right text-slate-500 tabular-nums">{fmt(item.unit_price)}</td>
									<td class="py-1.5 text-right font-semibold text-slate-800 dark:text-slate-200 tabular-nums">{fmt(item.total)}</td>
								</tr>
							{/each}
						</tbody>
					</table>

					<!-- Totals -->
					<div class="text-xs space-y-1">
						<div class="flex justify-between text-slate-500">
							<span>Subtotal</span>
							<span class="tabular-nums text-slate-700 dark:text-slate-300">{fmt(sale.subtotal)}</span>
						</div>
						{#if sale.discount > 0}
							<div class="flex justify-between text-emerald-600">
								<span>Discount</span>
								<span class="tabular-nums">-{fmt(sale.discount)}</span>
							</div>
						{/if}
						{#if sale.tax > 0}
							<div class="flex justify-between text-slate-500">
								<span>Tax</span>
								<span class="tabular-nums text-slate-700 dark:text-slate-300">{fmt(sale.tax)}</span>
							</div>
						{/if}
						<div class="flex justify-between font-bold text-sm text-slate-900 dark:text-slate-100 pt-1.5 border-t border-slate-300 dark:border-slate-500 mt-1.5">
							<span>TOTAL</span>
							<span class="tabular-nums">{fmt(sale.total)}</span>
						</div>
					</div>

					<!-- Payment -->
					<div class="mt-3 pt-3 border-t border-slate-200 dark:border-slate-600 text-[10px] space-y-1">
						<div class="flex justify-between text-slate-500">
							<span>Payment Method</span>
							<span class="font-semibold uppercase text-slate-700 dark:text-slate-300">{sale.payment_method}</span>
						</div>
						{#if sale.payment_method === 'cash'}
							<div class="flex justify-between text-slate-500">
								<span>Tendered</span>
								<span class="tabular-nums text-slate-700 dark:text-slate-300">{fmt(amountTendered)}</span>
							</div>
							<div class="flex justify-between text-slate-500">
								<span>Change</span>
								<span class="tabular-nums font-semibold text-slate-700 dark:text-slate-300">{fmt(change)}</span>
							</div>
						{/if}
					</div>

					<!-- Footer -->
					<div class="text-center mt-5 pt-3 border-t border-slate-100 dark:border-slate-700">
						<p class="text-[10px] text-slate-400">{shopInfo?.settings?.receipt_footer || 'Thank you for your business!'}</p>
						<p class="text-[8px] text-slate-300 mt-2">Powered by Maestro POS</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Action buttons -->
		<div class="flex items-center gap-2 border-t border-slate-200 dark:border-slate-600 px-5 py-4 shrink-0 bg-white dark:bg-slate-800">
			<button onclick={print} class="flex items-center justify-center gap-1.5 rounded-xl border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-xs font-semibold text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors min-w-[70px]">
				<Printer size={14} /> Print
			</button>
			{#if pointsPhone}
				<button onclick={sendWhatsApp} class="flex items-center justify-center gap-1.5 rounded-xl border border-emerald-200 dark:border-emerald-800 bg-emerald-50 dark:bg-emerald-900/20 px-4 py-2.5 text-xs font-semibold text-emerald-700 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/40 transition-colors min-w-[70px]">
					<Phone size={14} /> WhatsApp
				</button>
			{/if}
			<button onclick={onnewsale} class="flex-1 rounded-xl py-2.5 text-sm font-bold text-white bg-blue-600 hover:bg-blue-700 transition-all active:scale-95">
				New Sale
			</button>
		</div>

		<!-- Close button (top-right, not printable) -->
		<button onclick={onclose} class="absolute top-3 right-3 z-10 h-7 w-7 flex items-center justify-center rounded-full bg-white/90 dark:bg-slate-700/90 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 shadow-sm transition-colors">
			<X size={14} />
		</button>
	</div>
</div>

<style>
	@media print {
		:global(body *) { visibility: hidden; }
		:global(.receipt-print), :global(.receipt-print *) { visibility: visible; }
		:global(.receipt-print) { position: absolute; left: 0; top: 0; width: 80mm; padding: 0; }
		:global(.receipt-print) button, :global(.receipt-print) .shrink-0 { display: none; }
	}
</style>
