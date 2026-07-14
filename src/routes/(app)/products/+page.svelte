<script lang="ts">
	import { onMount } from 'svelte';
	import { productsService, type ProductInput } from '$lib/services/products';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import type { Product, Category } from '$lib/types';
	import { Plus, Search, Pencil, Trash2, AlertTriangle } from '@lucide/svelte';

	let products = $state<Product[]>([]);
	let categories = $state<Category[]>([]);
	let total = $state(0);
	let page = $state(1);
	const limit = 20;

	let search = $state('');
	let categoryFilter = $state('');
	let lowStockFilter = $state(false);
	let loading = $state(true);

	let showModal = $state(false);
	let editingProduct = $state<Product | null>(null);
	let submitting = $state(false);

	let form = $state<ProductInput>({
		name: '',
		barcode: null,
		sku: null,
		category_id: null,
		buying_price: 0,
		selling_price: 0,
		stock_qty: 0,
		reorder_level: 10,
		image_url: null
	});

	function fmt(n: number) {
		return new Intl.NumberFormat('en-KE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}

	async function fetchProducts() {
		loading = true;
		try {
			const res = await productsService.list({
				search,
				category_id: categoryFilter || undefined,
				low_stock: lowStockFilter || undefined,
				page,
				limit
			});
			products = res.data ?? [];
			total = res.total ?? 0;
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to load products');
		} finally {
			loading = false;
		}
	}

	let debounce: ReturnType<typeof setTimeout>;
	function onSearch() {
		page = 1;
		clearTimeout(debounce);
		debounce = setTimeout(fetchProducts, 300);
	}

	function openCreate() {
		editingProduct = null;
		form = { name: '', barcode: null, sku: null, category_id: null, buying_price: 0, selling_price: 0, stock_qty: 0, reorder_level: 10, image_url: null };
		showModal = true;
	}

	function openEdit(p: Product) {
		editingProduct = p;
		form = {
			name: p.name,
			barcode: p.barcode,
			sku: p.sku,
			category_id: p.category_id,
			buying_price: p.buying_price,
			selling_price: p.selling_price,
			stock_qty: p.stock_qty,
			reorder_level: p.reorder_level,
			image_url: p.image_url
		};
		showModal = true;
	}

	async function save() {
		if (!form.name.trim()) { notify.error('Name is required'); return; }
		submitting = true;
		try {
			if (editingProduct) {
				await productsService.update(editingProduct.id, form);
				notify.success('Product updated');
			} else {
				await productsService.create(form);
				notify.success('Product created');
			}
			showModal = false;
			fetchProducts();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Save failed');
		} finally {
			submitting = false;
		}
	}

	async function deleteProduct(p: Product) {
		if (!confirm(`Delete "${p.name}"?`)) return;
		try {
			await productsService.delete(p.id);
			notify.success('Product deleted');
			fetchProducts();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Delete failed');
		}
	}

	onMount(async () => {
		const catRes = await productsService.listCategories();
		categories = catRes.data ?? [];
		fetchProducts();
	});
</script>

<svelte:head><title>Products — POS</title></svelte:head>

<div class="p-6 space-y-5">
	<div class="flex items-center justify-between">
		<h1 class="text-xl font-semibold text-gray-900">Products</h1>
		<button onclick={openCreate} class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
			<Plus size={16} />
			Add Product
		</button>
	</div>

	<div class="flex gap-3 flex-wrap">
		<div class="relative flex-1 min-w-48">
			<Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
			<input
				bind:value={search}
				oninput={onSearch}
				placeholder="Search products…"
				class="w-full rounded-lg border border-gray-300 py-2 pl-9 pr-3 text-sm focus:border-blue-500 focus:outline-none"
			/>
		</div>
		<select
			bind:value={categoryFilter}
			onchange={() => { page = 1; fetchProducts(); }}
			class="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
		>
			<option value="">All Categories</option>
			{#each categories as cat}
				<option value={cat.id}>{cat.name}</option>
			{/each}
		</select>
		<label class="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
			<input type="checkbox" bind:checked={lowStockFilter} onchange={() => { page = 1; fetchProducts(); }} class="rounded" />
			Low Stock
		</label>
	</div>

	<div class="rounded-xl shadow-sm bg-white overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				<tr class="bg-slate-50">
					<th class="px-4 py-3 font-medium text-gray-600">Product</th>
					<th class="px-4 py-3 font-medium text-gray-600">Category</th>
					<th class="px-4 py-3 font-medium text-gray-600 text-right">Buy Price</th>
					<th class="px-4 py-3 font-medium text-gray-600 text-right">Sell Price</th>
					<th class="px-4 py-3 font-medium text-gray-600 text-right">Stock</th>
					<th class="px-4 py-3 font-medium text-gray-600 text-right">Actions</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100">
				{#if loading}
					{#each Array(8) as _}
						<tr>
							{#each Array(6) as _}
								<td class="px-4 py-3"><div class="h-4 bg-gray-100 rounded animate-pulse"></div></td>
							{/each}
						</tr>
					{/each}
				{:else if products.length === 0}
					<tr>
						<td colspan="6" class="px-4 py-12 text-center text-gray-400">No products found</td>
					</tr>
				{:else}
					{#each products as p}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3">
								<p class="font-medium text-gray-900">{p.name}</p>
								{#if p.barcode}<p class="text-xs text-gray-400">{p.barcode}</p>{/if}
							</td>
							<td class="px-4 py-3 text-gray-500">{p.category_name ?? '—'}</td>
							<td class="px-4 py-3 text-right text-gray-600">KES {fmt(p.buying_price)}</td>
							<td class="px-4 py-3 text-right font-medium">KES {fmt(p.selling_price)}</td>
							<td class="px-4 py-3 text-right">
								<span class="inline-flex items-center gap-1 {p.stock_qty <= p.reorder_level ? 'text-blue-600 font-medium' : 'text-gray-700'}">
									{#if p.stock_qty <= p.reorder_level}<AlertTriangle size={12} />{/if}
									{p.stock_qty}
								</span>
							</td>
							<td class="px-4 py-3 text-right">
								<div class="flex items-center justify-end gap-2">
									<button onclick={() => openEdit(p)} class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-blue-600">
										<Pencil size={14} />
									</button>
									<button onclick={() => deleteProduct(p)} class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-red-600">
										<Trash2 size={14} />
									</button>
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<Pagination {page} {total} {limit} onchange={(p) => { page = p; fetchProducts(); }} />
</div>

<Modal
	open={showModal}
	title={editingProduct ? 'Edit Product' : 'Add Product'}
	onclose={() => showModal = false}
	size="lg"
>
	{#snippet children()}
		<div class="grid grid-cols-2 gap-4">
			<div class="col-span-2">
				<label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
				<input bind:value={form.name} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Barcode</label>
				<input bind:value={form.barcode} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" placeholder="Leave empty if none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">SKU</label>
				<input bind:value={form.sku} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Category</label>
				<select bind:value={form.category_id} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none">
					<option value={null}>None</option>
					{#each categories as cat}
						<option value={cat.id}>{cat.name}</option>
					{/each}
				</select>
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Buying Price</label>
				<input type="number" bind:value={form.buying_price} min="0" step="0.01" class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Selling Price</label>
				<input type="number" bind:value={form.selling_price} min="0" step="0.01" class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Stock Qty</label>
				<input type="number" bind:value={form.stock_qty} min="0" class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Reorder Level</label>
				<input type="number" bind:value={form.reorder_level} min="0" class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div class="col-span-2">
				<label class="block text-sm font-medium text-gray-700 mb-1">Image URL</label>
				<input bind:value={form.image_url} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" placeholder="https://…" />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-lg border px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60">
			{submitting ? 'Saving…' : editingProduct ? 'Save Changes' : 'Create Product'}
		</button>
	{/snippet}
</Modal>

