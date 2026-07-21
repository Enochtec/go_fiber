<script lang="ts">
	import { onMount } from 'svelte';
	import { purchasesService, type CreatePurchaseInput, type PurchaseItemInput } from '$lib/services/purchases';
	import { suppliersService } from '$lib/services/suppliers';
	import { offlineProducts as productsService } from '$lib/services/offline';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import type { Purchase, Supplier, Product } from '$lib/types';
	import { Plus, Trash2, Search, Download } from '@lucide/svelte';
	import ExportModal from '$lib/components/ExportModal.svelte';
	import { shopService } from '$lib/services/shop';
	import { authStore } from '$lib/stores/auth.svelte';
	import { exportPurchases } from '$lib/services/export';

	let purchases = $state<Purchase[]>([]);
	let suppliers = $state<Supplier[]>([]);
	let total = $state(0);
	let page = $state(1);
	const limit = 20;
	let loading = $state(true);
	let showExport = $state(false);

	async function handleExport(_fmt: 'csv', _scope: 'all' | 'filtered' | 'current' | 'selected') {
		const info = await shopService.getInfo();
		const shopName = info?.shop?.name ?? 'Export';
		const userName = authStore.user?.name ?? 'System';
		await exportPurchases(purchases, shopName, userName);
	}

	let showModal = $state(false);
	let submitting = $state(false);

	let supplierId = $state<string>('');
	let status = $state<'received' | 'pending'>('received');
	let note = $state('');
	let items = $state<Array<{ product: Product | null; quantity: number; unit_price: number; search: string; results: Product[]; searchId: number; searching: boolean }>>([]);

	function addItem() {
		items.push({ product: null, quantity: 1, unit_price: 0, search: '', results: [], searchId: 0, searching: false });
	}

	function removeItem(i: number) {
		items.splice(i, 1);
	}

	let searchTimers: ReturnType<typeof setTimeout>[] = [];

	function onSearch(i: number) {
		clearTimeout(searchTimers[i]);
		const q = items[i].search;
		if (!q.trim()) {
			items[i].results = [];
			items[i].searching = false;
			return;
		}
		searchTimers[i] = setTimeout(() => doSearch(i), 80);
	}

	async function doSearch(i: number) {
		const q = items[i].search.trim();
		if (!q) return;
		const id = ++items[i].searchId;
		items[i].searching = true;
		const res = await productsService.list({ search: q, limit: 8 });
		if (id !== items[i].searchId) return;
		items[i].results = res.data ?? [];
		items[i].searching = false;
	}

	function selectProduct(i: number, p: Product) {
		items[i].product = p;
		items[i].unit_price = p.buying_price;
		items[i].search = p.name;
		items[i].results = [];
	}

	function fmt(n: number) {
		return new Intl.NumberFormat('en-KE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}

	const purchaseTotal = $derived(
		items.reduce((sum, i) => sum + (i.unit_price * i.quantity), 0)
	);

	async function fetch() {
		loading = true;
		try {
			const res = await purchasesService.list(page, limit);
			purchases = res.data ?? [];
			total = res.total ?? 0;
		} finally {
			loading = false;
		}
	}

	async function save() {
		const validItems = items.filter((i) => i.product !== null && i.quantity > 0);
		if (validItems.length === 0) { notify.error('Add at least one item'); return; }
		submitting = true;
		try {
			const payload: CreatePurchaseInput = {
				supplier_id: supplierId || null,
				items: validItems.map((i) => ({
					product_id: i.product!.id,
					quantity: i.quantity,
					unit_price: i.unit_price
				})),
				status,
				note: note || null
			};
			await purchasesService.create(payload);
			notify.success('Purchase order created');
			showModal = false;
			supplierId = '';
			note = '';
			items = [];
			fetch();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to create purchase');
		} finally {
			submitting = false;
		}
	}

	function fmtDate(s: string) {
		return new Date(s).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
	}

	onMount(async () => {
		const sRes = await suppliersService.list();
		suppliers = sRes.data ?? [];
		fetch();
	});
</script>

<svelte:head><title>Purchases — Maestro POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-5 min-h-full dark:bg-slate-950">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-lg font-bold text-slate-900 dark:text-slate-100">Purchase Orders</h1>
			<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Record stock purchases from suppliers</p>
		</div>
		<div class="flex gap-2">
			<button onclick={() => showExport = true} class="flex items-center gap-1.5 rounded-[1px] border border-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-600 hover:bg-slate-50 transition-colors">
				<Download size={13} /> Export
			</button>
			<button onclick={() => { showModal = true; addItem(); }} class="flex items-center gap-1.5 rounded-[1px] px-3 py-1.5 text-xs font-semibold text-white transition-all active:scale-95" style="background:linear-gradient(135deg,#ef4444,#dc2626);">
				<Plus size={13} />New Purchase
			</button>
		</div>
	</div>

	<div class="rounded-[1px] bg-white dark:bg-slate-800 overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				<tr style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
					<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide">Date</th>
					<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide">Supplier</th>
					<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide hidden md:table-cell">Created By</th>
					<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide">Status</th>
					<th class="px-4 py-2.5 text-right text-xs font-semibold text-white uppercase tracking-wide">Total</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
				{#if loading}
					{#each Array(5) as _}
						<tr>{#each Array(5) as _}<td class="px-4 py-2.5"><div class="h-3.5 bg-slate-100 dark:bg-slate-700 rounded-[1px] animate-pulse"></div></td>{/each}</tr>
					{/each}
				{:else if purchases.length === 0}
					<tr><td colspan="5" class="px-5 py-12 text-center text-sm text-slate-400 dark:text-slate-500">No purchase orders yet.</td></tr>
				{:else}
					{#each purchases as p}
						<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-colors">
							<td class="px-4 py-2.5 text-slate-600 dark:text-slate-300 whitespace-nowrap">{fmtDate(p.created_at)}</td>
							<td class="px-4 py-2.5 font-semibold text-slate-900 dark:text-slate-100">{p.supplier_name ?? 'Direct'}</td>
							<td class="px-4 py-2.5 text-slate-500 dark:text-slate-400 hidden md:table-cell">{p.user_name ?? '—'}</td>
							<td class="px-4 py-2.5">
								<span class="badge {p.status === 'received' ? 'badge-green' : 'badge-amber'} capitalize">{p.status}</span>
							</td>
							<td class="px-4 py-2.5 text-right font-semibold text-slate-900 dark:text-slate-100 tabular-nums">KES {fmt(p.total)}</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<Pagination {page} {total} {limit} onchange={(p) => { page = p; fetch(); }} />
</div>

<Modal open={showModal} title="New Purchase Order" onclose={() => showModal = false} size="lg">
	{#snippet children()}
		<div class="space-y-5">
			<div class="grid grid-cols-2 gap-4">
				<div>
					<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Supplier</label>
					<select bind:value={supplierId} class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors">
						<option value="">Direct Purchase</option>
						{#each suppliers as s}
							<option value={s.id}>{s.name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Status</label>
					<select bind:value={status} class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors">
						<option value="received">Received</option>
						<option value="pending">Pending</option>
					</select>
				</div>
			</div>

			<div>
				<div class="flex items-center justify-between mb-2">
					<label class="text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Items</label>
					<button onclick={addItem} class="text-xs font-semibold text-blue-600 hover:text-blue-700 transition-colors">+ Add Item</button>
				</div>
				<div class="space-y-2">
					{#each items as item, i}
						<div class="border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-900/30 p-3 space-y-2">
							<div class="relative">
								<Search size={13} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
								<input
									bind:value={item.search}
									oninput={() => onSearch(i)}
									placeholder="Search product…"
									class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 pl-8 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors"
								/>
								{#if item.search.trim() && (item.searching || item.results.length > 0)}
									<ul class="absolute z-20 mt-1 w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 shadow-lg max-h-32 overflow-y-auto">
										{#if item.searching}
											<li class="px-3 py-2 text-sm text-slate-400">Searching…</li>
										{:else}
											{#each item.results as p}
												<li><button onclick={() => selectProduct(i, p)} class="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-blue-50 dark:hover:bg-slate-700 transition-colors">{p.name}</button></li>
											{/each}
										{/if}
									</ul>
								{/if}
							</div>
							<div class="grid grid-cols-3 gap-2">
								<div>
									<label class="block text-xs text-slate-500 dark:text-slate-400 mb-1">Qty</label>
									<input type="number" bind:value={item.quantity} min="1" class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-2 py-1.5 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" />
								</div>
								<div>
									<label class="block text-xs text-slate-500 dark:text-slate-400 mb-1">Unit Price</label>
									<input type="number" bind:value={item.unit_price} min="0" step="0.01" class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-2 py-1.5 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" />
								</div>
								<div class="flex items-end justify-between">
									<div>
										<p class="text-xs text-slate-500 dark:text-slate-400 mb-1">Total</p>
										<p class="text-sm font-semibold text-slate-900 dark:text-slate-100 tabular-nums">KES {fmt(item.unit_price * item.quantity)}</p>
									</div>
									<button onclick={() => removeItem(i)} class="mb-1 text-slate-400 hover:text-red-500 transition-colors"><Trash2 size={14} /></button>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Note</label>
				<input bind:value={note} class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" placeholder="Optional note" />
			</div>

			<div class="flex justify-between items-center border-t border-slate-100 dark:border-slate-700 pt-3">
				<span class="text-xs text-slate-500 dark:text-slate-400">{items.length} item(s)</span>
				<span class="text-sm font-bold text-slate-900 dark:text-slate-100 tabular-nums">Total: KES {fmt(purchaseTotal)}</span>
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-[1px] border border-slate-200 dark:border-slate-600 px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-[1px] bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-60 transition-colors">
			{submitting ? 'Saving…' : 'Create Purchase'}
		</button>
	{/snippet}
</Modal>

<ExportModal
	open={showExport}
	title="Export Purchases"
	onclose={() => showExport = false}
	onexport={handleExport}
/>
