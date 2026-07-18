<script lang="ts">
	import { onMount } from 'svelte';
	import { productsService, type ProductInput } from '$lib/services/products';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import type { Product, Category } from '$lib/types';
	import { Plus, Search, Pencil, Trash2, AlertTriangle, Package } from '@lucide/svelte';
	import ImagePicker from '$lib/components/ImagePicker.svelte';

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

<div class="p-4 md:p-6 space-y-5 dark:bg-slate-950 min-h-full">

	<!-- Page header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
		<div>
			<h1 class="text-xl font-bold text-slate-900 dark:text-slate-100">Products</h1>
			<p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
				{total} product{total !== 1 ? 's' : ''} total
			</p>
		</div>
		<button
			onclick={openCreate}
			class="inline-flex items-center gap-2 rounded-xl px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-all active:scale-95"
			style="background-color: #00008B;"
		>
			<Plus size={16} />
			Add Product
		</button>
	</div>

	<!-- Filters -->
	<div class="flex gap-2.5 flex-wrap">
		<div class="relative flex-1 min-w-48">
			<Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
			<input
				bind:value={search}
				oninput={onSearch}
				placeholder="Search by name, barcode, SKU…"
				class="w-full rounded-xl border border-slate-200 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 dark:placeholder-slate-500 bg-white py-2.5 pl-9 pr-3 text-sm shadow-sm focus:outline-none"
			/>
		</div>
		<select
			bind:value={categoryFilter}
			onchange={() => { page = 1; fetchProducts(); }}
			class="rounded-xl border border-slate-200 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 bg-white px-3 py-2.5 text-sm shadow-sm focus:outline-none"
		>
			<option value="">All Categories</option>
			{#each categories as cat}
				<option value={cat.id}>{cat.name}</option>
			{/each}
		</select>
		<label class="inline-flex items-center gap-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-3 py-2.5 text-sm font-medium text-slate-600 dark:text-slate-300 cursor-pointer shadow-sm select-none">
			<input type="checkbox" bind:checked={lowStockFilter} onchange={() => { page = 1; fetchProducts(); }} class="rounded accent-teal-600" />
			Low Stock Only
		</label>
	</div>

	<!-- Table card -->
	<div class="rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 shadow-sm overflow-hidden">
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-slate-100 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/50">
						<th class="px-5 py-3.5 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide w-12">Image</th>
						<th class="px-5 py-3.5 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Product</th>
						<th class="px-5 py-3.5 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Category</th>
						<th class="px-5 py-3.5 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Cost</th>
						<th class="px-5 py-3.5 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Price</th>
						<th class="px-5 py-3.5 text-center text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Stock</th>
						<th class="px-5 py-3.5 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Margin</th>
						<th class="px-5 py-3.5"></th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
					{#if loading}
						{#each Array(8) as _}
							<tr>
								{#each Array(8) as _}
									<td class="px-5 py-3.5">
										<div class="h-4 bg-slate-100 dark:bg-slate-700 rounded-lg animate-pulse"></div>
									</td>
								{/each}
							</tr>
						{/each}
					{:else if products.length === 0}
						<tr>
							<td colspan="8" class="px-5 py-16 text-center">
								<div class="flex flex-col items-center gap-3 text-slate-400 dark:text-slate-500">
									<Package size={40} class="opacity-30" />
									<p class="text-sm font-medium">No products found</p>
									<p class="text-xs">Try adjusting your search or filters</p>
								</div>
							</td>
						</tr>
					{:else}
						{#each products as p}
							{@const margin = p.buying_price > 0 ? ((p.selling_price - p.buying_price) / p.buying_price * 100) : 0}
							<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-colors group">
								<td class="px-4 py-3.5">
									<div class="h-10 w-10 rounded-lg overflow-hidden bg-slate-100 dark:bg-slate-700 flex items-center justify-center">
										{#if p.image_url}
											<img src={p.image_url} alt={p.name} class="h-full w-full object-cover" loading="lazy" />
										{:else}
											<Package size={14} class="text-slate-300" />
										{/if}
									</div>
								</td>
								<td class="px-5 py-3.5">
									<p class="font-semibold text-slate-800 dark:text-slate-100">{p.name}</p>
									<div class="flex items-center gap-2 mt-0.5">
										{#if p.barcode}
											<p class="text-xs text-slate-400 font-mono">{p.barcode}</p>
										{/if}
										{#if p.sku}
											<p class="text-xs text-slate-400">SKU: {p.sku}</p>
										{/if}
									</div>
								</td>
								<td class="px-5 py-3.5">
									{#if p.category_name}
										<span class="inline-flex rounded-full bg-slate-100 dark:bg-slate-700 px-2.5 py-1 text-xs font-medium text-slate-600 dark:text-slate-300">
											{p.category_name}
										</span>
									{:else}
										<span class="text-slate-400 text-xs">—</span>
									{/if}
								</td>
								<td class="px-5 py-3.5 text-right text-slate-500 dark:text-slate-400 tabular-nums">
									{fmt(p.buying_price)}
								</td>
								<td class="px-5 py-3.5 text-right font-semibold text-slate-800 dark:text-slate-100 tabular-nums">
									{fmt(p.selling_price)}
								</td>
								<td class="px-5 py-3.5 text-center">
									{#if p.stock_qty === 0}
										<span class="inline-flex items-center gap-1 rounded-full bg-red-100 dark:bg-red-900/30 px-2.5 py-1 text-xs font-semibold text-red-700 dark:text-red-400">
											<AlertTriangle size={10} /> Out of stock
										</span>
									{:else if p.stock_qty <= p.reorder_level}
										<span class="inline-flex items-center gap-1 rounded-full bg-amber-100 dark:bg-amber-900/30 px-2.5 py-1 text-xs font-semibold text-amber-700 dark:text-amber-400">
											<AlertTriangle size={10} /> Low: {p.stock_qty}
										</span>
									{:else}
										<span class="inline-flex rounded-full bg-emerald-100 dark:bg-emerald-900/30 px-2.5 py-1 text-xs font-semibold text-emerald-700 dark:text-emerald-400">
											{p.stock_qty}
										</span>
									{/if}
								</td>
								<td class="px-5 py-3.5 text-right tabular-nums">
									<span class="text-xs font-semibold {margin >= 20 ? 'text-emerald-600 dark:text-emerald-400' : margin >= 5 ? 'text-amber-600 dark:text-amber-400' : 'text-red-600 dark:text-red-400'}">
										{margin.toFixed(1)}%
									</span>
								</td>
								<td class="px-5 py-3.5">
									<div class="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
										<button
											onclick={() => openEdit(p)}
											class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 hover:text-slate-700 dark:hover:text-slate-200 transition-colors"
											title="Edit"
										>
											<Pencil size={14} />
										</button>
										<button
											onclick={() => deleteProduct(p)}
											class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 hover:bg-red-50 dark:hover:bg-red-900/30 hover:text-red-600 transition-colors"
											title="Delete"
										>
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
	</div>

	<Pagination {page} {total} {limit} onchange={(p) => { page = p; fetchProducts(); }} />
</div>

<!-- Product Modal -->
<Modal
	open={showModal}
	title={editingProduct ? 'Edit Product' : 'New Product'}
	onclose={() => showModal = false}
	size="lg"
>
	{#snippet children()}
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
			<div class="sm:col-span-2">
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Product Name *</label>
				<input bind:value={form.name} class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" placeholder="e.g. Coca-Cola 500ml" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Barcode</label>
				<input bind:value={form.barcode} class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none font-mono" placeholder="Scan or leave blank" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">SKU</label>
				<input bind:value={form.sku} class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" placeholder="Optional" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Category</label>
				<select bind:value={form.category_id} class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none">
					<option value={null}>No Category</option>
					{#each categories as cat}
						<option value={cat.id}>{cat.name}</option>
					{/each}
				</select>
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Stock Qty</label>
				<input type="number" bind:value={form.stock_qty} min="0" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Reorder Level</label>
				<input type="number" bind:value={form.reorder_level} min="0" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Cost Price (KES)</label>
				<input type="number" bind:value={form.buying_price} min="0" step="0.01" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Selling Price (KES)</label>
				<input type="number" bind:value={form.selling_price} min="0" step="0.01" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div class="sm:col-span-2">
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Product Image</label>
				<ImagePicker value={form.image_url} onchange={(url) => form.image_url = url} />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-xl border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
			Cancel
		</button>
		<button onclick={save} disabled={submitting} class="rounded-xl px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition-all active:scale-95 disabled:opacity-60 disabled:cursor-not-allowed" style="background-color: #00008B;">
			{submitting ? 'Saving…' : editingProduct ? 'Save Changes' : 'Create Product'}
		</button>
	{/snippet}
</Modal>

