<script lang="ts">
	import { onMount } from 'svelte';
	import { Printer, Mail, MessageCircle, X, CheckCircle, Download, Send, Loader, Store, Phone, User, FileText } from '@lucide/svelte';
	import { emailService, validateEmail } from '$lib/services/email';
	import { shopService, type ShopInfo } from '$lib/services/shop';
	import { api } from '$lib/services/api';
	import type { Sale } from '$lib/types';
	import { jsPDF } from 'jspdf';
	import html2canvas from 'html2canvas';

	let {
		sale,
		amountTendered,
		change,
		onclose,
		onnewsale,
	}: {
		sale: Sale;
		amountTendered: number;
		change: number;
		onclose: () => void;
		onnewsale: () => void;
	} = $props();

	let shopInfo     = $state<ShopInfo | null>(null);
	let waNumber     = $state('');
	let loyaltyPhone = $state('');
	let emailAddr    = $state(sale.customer_name ?? '');
	let emailSending = $state(false);
	let emailSent    = $state(false);
	let emailError   = $state('');
	let waError      = $state('');

	let receiptRef = $state<HTMLDivElement | null>(null);

	onMount(async () => {
		try {
			shopInfo = await shopService.getInfo();
		} catch {}
	});

	function fmt(n: number) {
		return new Intl.NumberFormat('en-KE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}
	function fmtDate(d: string) {
		return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
	}
	function fmtTime(d: string) {
		return new Date(d).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });
	}
	function toWa(num: string) {
		const d = num.replace(/\D/g, '');
		if (d.startsWith('0'))   return '254' + d.slice(1);
		if (d.startsWith('254')) return d;
		if (d.startsWith('+'))   return d.slice(1);
		return d;
	}

	const receiptId = $derived(sale.id.slice(0, 10).toUpperCase());
	const lineItems = $derived(sale.items ?? []);

	function buildReceiptHTML(): string {
		const shop = shopInfo?.shop;
		const s    = shopInfo?.settings;
		let itemsHtml = '';
		for (const item of lineItems) {
			itemsHtml += `<tr>
				<td style="padding:4px 8px;font-size:10px;">${item.product_name || 'Item'}</td>
				<td style="padding:4px 8px;font-size:10px;text-align:center;">${item.quantity}</td>
				<td style="padding:4px 8px;font-size:10px;text-align:right;">${fmt(item.unit_price)}</td>
				<td style="padding:4px 8px;font-size:10px;text-align:right;font-weight:bold;">${fmt(item.total)}</td>
			</tr>`;
		}

		const discountRow = sale.discount > 0
			? `<tr><td colspan="3" style="padding:3px 8px;font-size:10px;text-align:right;">Discount</td><td style="padding:3px 8px;font-size:10px;text-align:right;color:#d00;">-${fmt(sale.discount)}</td></tr>`
			: '';
		const taxRow = sale.tax > 0
			? `<tr><td colspan="3" style="padding:3px 8px;font-size:10px;text-align:right;">Tax</td><td style="padding:3px 8px;font-size:10px;text-align:right;">${fmt(sale.tax)}</td></tr>`
			: '';
		const cashRows = sale.payment_method === 'cash'
			? `<tr><td colspan="3" style="padding:4px 8px;font-size:10px;text-align:right;">Tendered</td><td style="padding:4px 8px;font-size:10px;text-align:right;">${fmt(amountTendered)}</td></tr>
			   <tr><td colspan="3" style="padding:4px 8px;font-size:10px;text-align:right;">Change</td><td style="padding:4px 8px;font-size:10px;text-align:right;">${fmt(change)}</td></tr>`
			: '';

		const contact = [shop?.phone, shop?.email].filter(Boolean).join(' | ');

		return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Receipt #${receiptId}</title>
<style>
  @page { margin:0; size:80mm auto; }
  * { margin:0; padding:0; box-sizing:border-box; }
  body { font-family:'Courier New',Courier,monospace; font-size:10px; color:#000; background:#fff; width:80mm; padding:12px 16px; }
  .hdr { text-align:center; border-bottom:2px solid #000; padding-bottom:6px; margin-bottom:6px; }
  .hdr .nm { font-size:13px; font-weight:bold; letter-spacing:2px; text-transform:uppercase; }
  .hdr .sm { font-size:8px; color:#555; }
  .ttl { text-align:center; font-size:9px; font-weight:bold; letter-spacing:3px; margin-bottom:6px; }
  .inf { font-size:8px; border-bottom:1px dashed #999; padding-bottom:4px; margin-bottom:6px; }
  .inf l { display:flex; justify-content:space-between; }
  .tbl { width:100%; border-collapse:collapse; }
  .tbl th { font-size:8px; text-align:left; border-bottom:1px solid #000; padding:3px 8px; }
  .tbl th.r { text-align:right; }
  .tbl th.c { text-align:center; }
  .tot { font-size:9px; margin-bottom:6px; }
  .tot l { display:flex; justify-content:space-between; padding:2px 8px; }
  .tot .gr { border-top:2px solid #000; margin-top:2px; padding-top:4px; font-size:11px; font-weight:bold; }
  .pay { border-top:1px dashed #999; padding-top:4px; font-size:9px; margin-bottom:8px; }
  .pay l { display:flex; justify-content:space-between; padding:1px 8px; }
  .ftr { border-top:2px solid #000; padding-top:6px; text-align:center; font-size:8px; }
  .bc { margin:6px auto 0; width:120px; height:24px; background:repeating-linear-gradient(90deg,#000 0,#000 2px,transparent 2px,transparent 4px); }
</style>
</head>
<body>
  <div class="hdr">
    <div class="nm">${shop?.name || 'MAESTRO POS'}</div>
    ${shop?.address ? `<div class="sm">${shop.address}</div>` : ''}
    ${contact ? `<div class="sm">${contact}</div>` : ''}
  </div>
  <div class="ttl">* * * R E C E I P T * * *</div>
  <div class="inf">
    <l><span>Receipt No.</span><span style="font-weight:bold;">#${receiptId}</span></l>
    <l><span>Date</span><span>${fmtDate(sale.created_at)}</span></l>
    <l><span>Time</span><span>${fmtTime(sale.created_at)}</span></l>
    <l><span>Cashier</span><span>${sale.cashier_name || '—'}</span></l>
    ${sale.customer_name ? `<l><span>Customer</span><span>${sale.customer_name}</span></l>` : ''}
    ${loyaltyPhone ? `<l><span>Phone</span><span>${loyaltyPhone}</span></l>` : ''}
  </div>
  <table class="tbl">
    <thead><tr><th>ITEM</th><th class="c">QTY</th><th class="r">PRICE</th><th class="r">AMOUNT</th></tr></thead>
    <tbody>${itemsHtml}</tbody>
  </table>
  <div class="tot">
    <l><span>Subtotal</span><span>${fmt(sale.subtotal)}</span></l>
    ${discountRow ? `<l style="color:#d00;"><span>Discount</span><span>-${fmt(sale.discount)}</span></l>` : ''}
    ${taxRow}
    <l class="gr"><span>TOTAL</span><span style="font-size:12px;">KES ${fmt(sale.total)}</span></l>
  </div>
  <div class="pay">
    <l><span>Payment</span><span style="font-weight:bold;text-transform:uppercase;">${sale.payment_method}</span></l>
    ${cashRows}
  </div>
  <div class="ftr">
    <p style="font-style:italic;">${s?.receipt_footer || 'Thank you for your business!'}</p>
    <p style="margin-top:4px;letter-spacing:1px;">* * * * * * * * * * * * * * * * *</p>
    <div class="bc"></div>
    <p style="margin-top:2px;font-size:7px;letter-spacing:2px;">${receiptId}</p>
    <p style="margin-top:4px;font-size:7px;color:#999;">Powered by Maestro POS</p>
  </div>
</body>
</html>`;
	}

	function printReceipt() {
		const win = window.open('', '_blank', 'width=420,height=700');
		if (!win) return;
		win.document.write(buildReceiptHTML());
		win.document.close();
		win.focus();
		setTimeout(() => { win.print(); }, 300);
	}

	async function downloadPDF() {
		if (!receiptRef) return;
		try {
			const canvas = await html2canvas(receiptRef, {
				scale: 3,
				useCORS: true,
				logging: false,
				backgroundColor: '#ffffff'
			});
			const imgData = canvas.toDataURL('image/png');
			const pdf = new jsPDF('p', 'mm', [80, 297]);
			const imgW = 80;
			const imgH = (canvas.height / canvas.width) * imgW;
			pdf.addImage(imgData, 'PNG', 0, 0, imgW, imgH);
			pdf.save(`receipt-${receiptId}.pdf`);
		} catch (e) {
			printReceipt();
		}
	}

	async function sendWhatsApp() {
		const num = waNumber.trim();
		if (!num) { waError = 'Enter a WhatsApp number'; return; }
		waError = '';

		if (!receiptRef) return;
		try {
			const canvas = await html2canvas(receiptRef, {
				scale: 3,
				useCORS: true,
				logging: false,
				backgroundColor: '#ffffff'
			});
			const imgData = canvas.toDataURL('image/png');
			const pdf = new jsPDF('p', 'mm', [80, 297]);
			const imgW = 80;
			const imgH = (canvas.height / canvas.width) * imgW;
			pdf.addImage(imgData, 'PNG', 0, 0, imgW, imgH);
			const blob = pdf.output('blob');

			const file = new File([blob], `receipt-${receiptId}.pdf`, { type: 'application/pdf' });
			if (navigator.canShare && navigator.canShare({ files: [file] })) {
				await navigator.share({ files: [file], title: `Receipt #${receiptId}` });
			} else {
				const url = URL.createObjectURL(blob);
				window.open(`https://wa.me/${toWa(num)}?text=${encodeURIComponent('Here is your receipt: ' + url)}`, '_blank');
				setTimeout(() => URL.revokeObjectURL(url), 30000);
			}
		} catch (e) {
			waError = 'Failed to generate PDF. Try the print button instead.';
		}
	}

	async function sendEmail() {
		const err = validateEmail(emailAddr);
		if (err) { emailError = err; return; }
		emailError   = '';
		emailSending = true;
		try {
			await emailService.sendReceipt({
				sale_id:  sale.id,
				email:    emailAddr.trim(),
				tendered: amountTendered,
				change:   change
			});
			emailSent = true;
		} catch (e) {
			emailError = e instanceof Error ? e.message : 'Failed to send email';
		} finally {
			emailSending = false;
		}
	}
</script>

<!-- Backdrop -->
<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-3 sm:p-6">
	<div class="relative w-full max-w-4xl max-h-[96vh] flex flex-col lg:flex-row bg-white shadow-2xl overflow-hidden">

		<!-- LEFT PANEL: Actions -->
		<div class="flex flex-col w-full lg:w-[44%] bg-white overflow-y-auto">
			<button onclick={onclose} class="absolute top-3 right-3 z-20 h-8 w-8 flex items-center justify-center bg-slate-100 hover:bg-slate-200 text-slate-500 transition-colors">
				<X size={16} />
			</button>

			<div class="flex flex-col items-center pt-8 pb-5 px-6 bg-gradient-to-b from-indigo-50 to-white">
				<div class="h-16 w-16 flex items-center justify-center rounded-full bg-emerald-500 text-white mb-3 success-ring shadow-lg shadow-emerald-200">
					<CheckCircle size={36} />
				</div>
				<h1 class="text-xl font-bold text-slate-900">Payment Successful!</h1>
				<p class="text-sm text-slate-500 mt-0.5">Receipt #{receiptId}</p>
			</div>

			<div class="px-5 pb-3">
				<div class="bg-slate-900 text-white px-4 py-4 flex items-center justify-between">
					<div>
						<p class="text-xs text-slate-400">Grand Total</p>
						<p class="text-2xl font-bold">KES {fmt(sale.total)}</p>
					</div>
					{#if sale.payment_method === 'cash'}
						<div class="text-right">
							<p class="text-xs text-slate-400">Tendered / Change</p>
							<p class="text-sm font-semibold">{fmt(amountTendered)} / {fmt(change)}</p>
						</div>
					{/if}
				</div>
			</div>

			<div class="px-5 pb-5 space-y-3">
				<div class="grid grid-cols-3 gap-2">
					<button onclick={printReceipt} class="flex flex-col items-center gap-1 py-2.5 border border-slate-200 text-xs font-medium text-white transition-all active:scale-[.97]" style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
						<Printer size={16} />
						Print
					</button>
					<button onclick={downloadPDF} class="flex flex-col items-center gap-1 py-2.5 border border-slate-200 text-xs font-medium text-white transition-all active:scale-[.97]" style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
						<Download size={16} />
						PDF
					</button>
					<button onclick={printReceipt} class="flex flex-col items-center gap-1 py-2.5 border border-slate-200 text-xs font-medium text-white transition-all active:scale-[.97]" style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
						<FileText size={16} />
						Reprint
					</button>
				</div>

				<!-- Loyalty phone -->
				<div class="border border-slate-200">
					<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide px-3 pt-2 mb-1 flex items-center gap-1.5">
						<Store size={11} /> Loyalty / Points
					</p>
					<div class="flex gap-0 px-3 pb-3">
						<div class="relative flex-1">
							<Phone size={12} class="absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400" />
							<input bind:value={loyaltyPhone} type="tel" placeholder="Phone for points"
								class="w-full border border-slate-200 border-r-0 pl-8 pr-2.5 py-1.5 text-xs focus:outline-none focus:border-indigo-400 min-w-0" />
						</div>
						<button onclick={() => {}}
							class="flex items-center gap-1 bg-indigo-500 hover:bg-indigo-600 text-white px-3 py-1.5 text-xs font-semibold transition-colors whitespace-nowrap">
							<User size={12} /> Apply
						</button>
					</div>
				</div>

				<!-- WhatsApp -->
				<div class="border border-slate-200">
					<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide px-3 pt-2 mb-1 flex items-center gap-1.5">
						<MessageCircle size={11} class="text-emerald-500" /> WhatsApp PDF
					</p>
					<div class="flex gap-0 px-3 pb-3">
						<div class="relative flex-1">
							<MessageCircle size={12} class="absolute left-2.5 top-1/2 -translate-y-1/2 text-emerald-400" />
							<input bind:value={waNumber} type="tel" placeholder="e.g. 0712 345 678"
								class="w-full border border-slate-200 border-r-0 pl-8 pr-2.5 py-1.5 text-xs focus:outline-none focus:border-emerald-400 min-w-0" />
						</div>
						<button onclick={sendWhatsApp}
							class="flex items-center gap-1.5 bg-emerald-500 hover:bg-emerald-600 text-white px-3 py-1.5 text-xs font-semibold transition-colors whitespace-nowrap">
							<Send size={12} />
						</button>
					</div>
					{#if waError}<p class="text-xs text-red-500 px-3 pb-2">{waError}</p>{/if}
				</div>

				<!-- Email -->
				<div class="border border-slate-200">
					<p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide px-3 pt-2 mb-1 flex items-center gap-1.5">
						<Mail size={11} class="text-blue-500" /> Email Receipt
					</p>
					<div class="flex gap-0 px-3 pb-3">
						<div class="relative flex-1">
							<Mail size={12} class="absolute left-2.5 top-1/2 -translate-y-1/2 text-blue-400" />
							<input bind:value={emailAddr} type="email" placeholder="customer@email.com"
								class="w-full border border-slate-200 border-r-0 pl-8 pr-2.5 py-1.5 text-xs focus:outline-none focus:border-blue-400 min-w-0" />
						</div>
						<button onclick={sendEmail} disabled={emailSending || emailSent}
							class="flex items-center gap-1.5 bg-blue-600 hover:bg-blue-700 text-white px-3 py-1.5 text-xs font-semibold disabled:opacity-60 transition-colors whitespace-nowrap">
							{#if emailSending}
								<Loader size={12} class="animate-spin" />
							{:else if emailSent}
								<CheckCircle size={12} />
							{:else}
								<Send size={12} />
							{/if}
						</button>
					</div>
					{#if emailError}<p class="text-xs text-red-500 px-3 pb-2">{emailError}</p>{/if}
					{#if emailSent}<p class="text-xs text-emerald-600 px-3 pb-2">Receipt sent to {emailAddr}</p>{/if}
				</div>

				<button onclick={onnewsale}
					class="w-full py-3 text-sm font-bold text-white transition-all active:scale-[.98] flex items-center justify-center gap-2"
					style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
					Start New Sale
				</button>
			</div>
		</div>

		<!-- RIGHT PANEL: Receipt preview (hidden, used as capture source) -->
		<div class="hidden lg:flex flex-col w-[56%] bg-slate-100 overflow-y-auto">
			<div class="flex items-center justify-between px-5 py-3 bg-slate-200">
				<p class="text-xs font-semibold text-slate-500 uppercase tracking-wide">Receipt Preview (80mm)</p>
				<button onclick={printReceipt} class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white transition-colors" style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
					<Printer size={12} /> Print
				</button>
			</div>
			<div class="flex-1 flex justify-center py-6 px-4">
				<div bind:this={receiptRef} class="w-[80mm] bg-white shadow-lg px-5 py-4" style="font-family:'Courier New',Courier,monospace;font-size:10px;color:#000;">

					<div style="text-align:center;border-bottom:2px solid #000;padding-bottom:6px;margin-bottom:6px;">
						<p style="font-size:13px;font-weight:bold;letter-spacing:2px;text-transform:uppercase;">{shopInfo?.shop?.name || 'MAESTRO POS'}</p>
						{#if shopInfo?.shop?.address}<p style="font-size:8px;margin-top:1px;">{shopInfo.shop.address}</p>{/if}
						{#if shopInfo?.shop?.phone || shopInfo?.shop?.email}
							<p style="font-size:8px;margin-top:1px;">{shopInfo?.shop?.phone ?? ''}{shopInfo?.shop?.phone && shopInfo?.shop?.email ? ' | ' : ''}{shopInfo?.shop?.email ?? ''}</p>
						{/if}
					</div>

					<p style="text-align:center;font-size:9px;font-weight:bold;letter-spacing:3px;margin-bottom:6px;">* * * R E C E I P T * * *</p>

					<div style="font-size:8px;margin-bottom:6px;border-bottom:1px dashed #999;padding-bottom:4px;">
						<div style="display:flex;justify-content:space-between;"><span>Receipt No.</span><span style="font-weight:bold;">#{receiptId}</span></div>
						<div style="display:flex;justify-content:space-between;"><span>Date</span><span>{fmtDate(sale.created_at)}</span></div>
						<div style="display:flex;justify-content:space-between;"><span>Time</span><span>{fmtTime(sale.created_at)}</span></div>
						<div style="display:flex;justify-content:space-between;"><span>Cashier</span><span>{sale.cashier_name || '—'}</span></div>
						{#if sale.customer_name}<div style="display:flex;justify-content:space-between;"><span>Customer</span><span>{sale.customer_name}</span></div>{/if}
						{#if loyaltyPhone}<div style="display:flex;justify-content:space-between;"><span>Phone</span><span>{loyaltyPhone}</span></div>{/if}
					</div>

					<div style="border-bottom:2px solid #000;padding-bottom:4px;margin-bottom:6px;">
						<div style="display:flex;justify-content:space-between;font-size:8px;font-weight:bold;margin-bottom:3px;padding-bottom:2px;border-bottom:1px solid #000;">
							<span style="flex:1;">ITEM</span>
							<span style="width:50px;text-align:center;">QTY</span>
							<span style="width:70px;text-align:right;">AMOUNT</span>
						</div>
						{#each lineItems as item}
							<div style="margin-bottom:3px;">
								<p style="font-size:8px;font-weight:600;margin:0;">{item.product_name || 'Item'}</p>
								<div style="display:flex;justify-content:space-between;padding-left:4px;font-size:8px;">
									<span style="flex:1;">@ {fmt(item.unit_price)}</span>
									<span style="width:50px;text-align:center;">{item.quantity}</span>
									<span style="width:70px;text-align:right;font-weight:bold;">{fmt(item.total)}</span>
								</div>
							</div>
						{/each}
					</div>

					<div style="font-size:9px;margin-bottom:6px;">
						<div style="display:flex;justify-content:space-between;"><span>Subtotal</span><span>{fmt(sale.subtotal)}</span></div>
						{#if sale.discount > 0}<div style="display:flex;justify-content:space-between;"><span>Discount</span><span style="color:#d00;">-{fmt(sale.discount)}</span></div>{/if}
						{#if sale.tax > 0}<div style="display:flex;justify-content:space-between;"><span>Tax</span><span>{fmt(sale.tax)}</span></div>{/if}
						<div style="display:flex;justify-content:space-between;font-weight:bold;border-top:2px solid #000;margin-top:3px;padding-top:3px;font-size:11px;">
							<span>TOTAL</span><span style="font-size:12px;">KES {fmt(sale.total)}</span>
						</div>
					</div>

					<div style="border-top:1px dashed #999;padding-top:4px;font-size:9px;margin-bottom:8px;">
						<div style="display:flex;justify-content:space-between;"><span>Payment</span><span style="font-weight:bold;text-transform:uppercase;">{sale.payment_method}</span></div>
						{#if sale.payment_method === 'cash'}
							<div style="display:flex;justify-content:space-between;"><span>Tendered</span><span>{fmt(amountTendered)}</span></div>
							<div style="display:flex;justify-content:space-between;"><span>Change</span><span>{fmt(change)}</span></div>
						{/if}
					</div>

					<div style="border-top:2px solid #000;padding-top:6px;text-align:center;font-size:8px;">
						<p style="font-style:italic;">{shopInfo?.settings?.receipt_footer || 'Thank you for your business!'}</p>
						<p style="margin-top:4px;letter-spacing:1px;">* * * * * * * * * * * * * * * * *</p>
						<div style="margin:6px auto 0;width:120px;height:24px;background:repeating-linear-gradient(90deg,#000 0,#000 2px,transparent 2px,transparent 4px);"></div>
						<p style="margin-top:2px;font-size:7px;letter-spacing:2px;">{receiptId}</p>
						<p style="margin-top:4px;font-size:7px;color:#999;">Powered by Maestro POS</p>
					</div>
				</div>
			</div>
		</div>

	</div>
</div>

<style>
	@keyframes ring {
		0%   { box-shadow: 0 0 0 0 rgba(16,185,129,.5); }
		70%  { box-shadow: 0 0 0 12px rgba(16,185,129,0); }
		100% { box-shadow: 0 0 0 0 rgba(16,185,129,0); }
	}
	:global(.success-ring) {
		animation: ring 1.2s ease 0.1s 2;
	}
</style>
