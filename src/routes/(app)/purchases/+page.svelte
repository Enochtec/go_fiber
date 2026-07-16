<script lang="ts">
	import { onMount } from 'svelte';
	import { purchasesService, type CreatePurchaseInput, type PurchaseItemInput } from '$lib/services/purchases';
	import { suppliersService } from '$lib/services/suppliers';
	import { productsService } from '$lib/services/products';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import type { Purchase, Supplier, Product } from '$lib/types';
	import { Plus, Trash2, Search } from '@lucide/svelte';

	let purchases = $state<Purchase[]>([]);
	let suppliers = $state<Supplier[]>([]);
	let total = $state(0);
	let page = $state(1);
	const limit = 20;
	let loading = $state(true);

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

<svelte:head><title>Purchases — POS</title></svelte:head>

<div class="p-6 space-y-5">
	<div class="flex items-center justify-between">
		<h1 class="text-xl font-semibold text-gray-900">Purchase Orders</h1>
		<button onclick={() => { showModal = true; addItem(); }} class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
			<Plus size={16} />New Purchase
		</button>
	</div>

	<div class="rounded-xl shadow-sm bg-white overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				<tr class="bg-slate-50">
					<th class="px-4 py-3 font-medium text-gray-600">Date</th>
					<th class="px-4 py-3 font-medium text-gray-600">Supplier</th>
					<th class="px-4 py-3 font-medium text-gray-600">Created By</th>
					<th class="px-4 py-3 font-medium text-gray-600">Status</th>
					<th class="px-4 py-3 font-medium text-gray-600 text-right">Total</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100">
				{#if loading}
					{#each Array(5) as _}
						<tr>{#each Array(5) as _}<td class="px-4 py-3"><div class="h-4 bg-gray-100 rounded animate-pulse"></div></td>{/each}</tr>
					{/each}
				{:else if purchases.length === 0}
					<tr><td colspan="5" class="px-4 py-12 text-center text-gray-400">No purchases yet</td></tr>
				{:else}
					{#each purchases as p}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-gray-700">{fmtDate(p.created_at)}</td>
							<td class="px-4 py-3 font-medium text-gray-900">{p.supplier_name ?? 'Direct'}</td>
							<td class="px-4 py-3 text-gray-500">{p.user_name ?? '—'}</td>
							<td class="px-4 py-3">
								<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium
									{p.status === 'received' ? 'bg-green-50 text-green-700' : 'bg-yellow-50 text-yellow-700'}">
									{p.status}
								</span>
							</td>
							<td class="px-4 py-3 text-right font-semibold text-gray-900">KES {fmt(p.total)}</td>
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
		<div class="space-y-4">
			<div class="grid grid-cols-2 gap-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Supplier</label>
					<select bind:value={supplierId} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none">
						<option value="">Direct Purchase</option>
						{#each suppliers as s}
							<option value={s.id}>{s.name}</option>
						{/each}
					</select>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
					<select bind:value={status} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none">
						<option value="received">Received</option>
						<option value="pending">Pending</option>
					</select>
				</div>
			</div>

			<div>
				<div class="flex items-center justify-between mb-2">
					<label class="text-sm font-medium text-gray-700">Items</label>
					<button onclick={addItem} class="text-xs text-blue-600 hover:text-blue-700 font-medium">+ Add Item</button>
				</div>
				<div class="space-y-2">
					{#each items as item, i}
						<div class="rounded-lg border p-3 space-y-2">
							<div class="relative">
								<Search size={13} class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
								<input
									bind:value={item.search}
									oninput={() => onSearch(i)}
									placeholder="Search product…"
									class="w-full rounded-lg border px-3 py-2 pl-8 text-sm focus:border-blue-500 focus:outline-none"
								/>
								{#if item.search.trim() && (item.searching || item.results.length > 0)}
									<ul class="absolute z-10 mt-1 w-full rounded-lg border bg-white shadow-lg max-h-32 overflow-y-auto">
										{#if item.searching}
											<li class="px-3 py-2 text-sm text-gray-400">Searching…</li>
										{:else}
											{#each item.results as p}
												<li><button onclick={() => selectProduct(i, p)} class="w-full px-3 py-2 text-left text-sm hover:bg-blue-50">{p.name}</button></li>
											{/each}
										{/if}
									</ul>
								{/if}
							</div>
							<div class="grid grid-cols-3 gap-2">
								<div>
									<label class="block text-xs text-gray-500 mb-1">Qty</label>
									<input type="number" bind:value={item.quantity} min="1" class="w-full rounded border px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none" />
								</div>
								<div>
									<label class="block text-xs text-gray-500 mb-1">Unit Price</label>
									<input type="number" bind:value={item.unit_price} min="0" step="0.01" class="w-full rounded border px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none" />
								</div>
								<div class="flex items-end justify-between">
									<div>
										<p class="text-xs text-gray-500 mb-1">Total</p>
										<p class="text-sm font-semibold">KES {fmt(item.unit_price * item.quantity)}</p>
									</div>
									<button onclick={() => removeItem(i)} class="mb-1 text-gray-400 hover:text-red-500"><Trash2 size={14} /></button>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Note</label>
				<input bind:value={note} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>

			<div class="flex justify-end text-sm font-bold text-gray-900 border-t pt-3">
				Total: KES {fmt(purchaseTotal)}
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-lg border px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60">
			{submitting ? 'Saving…' : 'Create Purchase'}
		</button>
	{/snippet}
</Modal>
