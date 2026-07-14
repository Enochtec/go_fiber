<script lang="ts">
	import { onMount } from 'svelte';
	import { inventoryService, type AdjustInput } from '$lib/services/inventory';
	import { productsService } from '$lib/services/products';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import type { StockAdjustment, Product } from '$lib/types';
	import { Plus, AlertTriangle, Search } from '@lucide/svelte';

	let adjustments = $state<StockAdjustment[]>([]);
	let products = $state<Product[]>([]);
	let total = $state(0);
	let page = $state(1);
	const limit = 20;
	let loading = $state(true);

	let showModal = $state(false);
	let submitting = $state(false);
	let productSearch = $state('');
	let selectedProduct = $state<Product | null>(null);

	let form = $state<AdjustInput>({ product_id: '', quantity: 0, reason: '' });

	async function fetch() {
		loading = true;
		try {
			const res = await inventoryService.listAdjustments(page, limit);
			adjustments = res.data ?? [];
			total = res.total ?? 0;
		} finally {
			loading = false;
		}
	}

	async function searchProducts() {
		if (!productSearch.trim()) { products = []; return; }
		const res = await productsService.list({ search: productSearch, limit: 10 });
		products = res.data ?? [];
	}

	function selectProduct(p: Product) {
		selectedProduct = p;
		form.product_id = p.id;
		productSearch = p.name;
		products = [];
	}

	async function save() {
		if (!form.product_id) { notify.error('Select a product'); return; }
		if (form.quantity === 0) { notify.error('Quantity cannot be zero'); return; }
		if (!form.reason.trim()) { notify.error('Reason is required'); return; }
		submitting = true;
		try {
			await inventoryService.adjust(form);
			notify.success('Stock adjusted');
			showModal = false;
			form = { product_id: '', quantity: 0, reason: '' };
			selectedProduct = null;
			productSearch = '';
			fetch();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Adjustment failed');
		} finally {
			submitting = false;
		}
	}

	function fmtDate(s: string) {
		return new Date(s).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	onMount(fetch);
</script>

<svelte:head><title>Inventory — POS</title></svelte:head>

<div class="p-6 space-y-5">
	<div class="flex items-center justify-between">
		<h1 class="text-xl font-semibold text-gray-900">Inventory</h1>
		<button onclick={() => showModal = true} class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
			<Plus size={16} />Stock Adjustment
		</button>
	</div>

	<div class="rounded-xl shadow-sm bg-white overflow-hidden">
		<div class="px-4 py-3">
			<p class="text-sm font-medium text-gray-700">Stock Adjustments History</p>
		</div>
		<table class="w-full text-sm">
			<thead>
				<tr class="bg-slate-50">
					<th class="px-4 py-3 font-medium text-gray-600">Product</th>
					<th class="px-4 py-3 font-medium text-gray-600">Quantity</th>
					<th class="px-4 py-3 font-medium text-gray-600">Reason</th>
					<th class="px-4 py-3 font-medium text-gray-600">By</th>
					<th class="px-4 py-3 font-medium text-gray-600">Date</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100">
				{#if loading}
					{#each Array(6) as _}
						<tr>{#each Array(5) as _}<td class="px-4 py-3"><div class="h-4 bg-gray-100 rounded animate-pulse"></div></td>{/each}</tr>
					{/each}
				{:else if adjustments.length === 0}
					<tr><td colspan="5" class="px-4 py-12 text-center text-gray-400">No adjustments yet</td></tr>
				{:else}
					{#each adjustments as a}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 font-medium text-gray-900">{a.product_name ?? '—'}</td>
							<td class="px-4 py-3">
								<span class="font-semibold {a.quantity > 0 ? 'text-green-600' : 'text-red-600'}">
									{a.quantity > 0 ? '+' : ''}{a.quantity}
								</span>
							</td>
							<td class="px-4 py-3 text-gray-600">{a.reason}</td>
							<td class="px-4 py-3 text-gray-500">{a.user_name ?? '—'}</td>
							<td class="px-4 py-3 text-gray-400">{fmtDate(a.created_at)}</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<Pagination {page} {total} {limit} onchange={(p) => { page = p; fetch(); }} />
</div>

<Modal open={showModal} title="Stock Adjustment" onclose={() => showModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-4">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Product *</label>
				<div class="relative">
					<Search size={13} class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
					<input
						bind:value={productSearch}
						oninput={searchProducts}
						placeholder="Search product…"
						class="w-full rounded-lg border px-3 py-2 pl-8 text-sm focus:border-blue-500 focus:outline-none"
					/>
				</div>
				{#if products.length > 0}
					<ul class="mt-1 rounded-lg border shadow-sm max-h-40 overflow-y-auto">
						{#each products as p}
							<li>
								<button onclick={() => selectProduct(p)} class="w-full px-3 py-2 text-left text-sm hover:bg-blue-50">
									<span class="font-medium">{p.name}</span>
									<span class="text-gray-400 ml-2">Stock: {p.stock_qty}</span>
								</button>
							</li>
						{/each}
					</ul>
				{/if}
				{#if selectedProduct}
					<p class="mt-1 text-xs text-blue-600">Current stock: {selectedProduct.stock_qty}</p>
				{/if}
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Quantity (negative to deduct)</label>
				<input type="number" bind:value={form.quantity} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Reason *</label>
				<input bind:value={form.reason} placeholder="e.g. Damaged goods, Stock count correction…" class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-lg border px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60">
			{submitting ? 'Saving…' : 'Apply Adjustment'}
		</button>
	{/snippet}
</Modal>
