<script lang="ts">
	import { onMount } from 'svelte';
	import { cart } from '$lib/stores/cart.svelte';
	import { notify } from '$lib/stores/notification.svelte';
	import { productsService } from '$lib/services/products';
	import { customersService } from '$lib/services/customers';
	import { salesService } from '$lib/services/sales';
	import Modal from '$lib/components/Modal.svelte';
	import type { Product, Customer, Sale } from '$lib/types';
	import {
		Search, ShoppingCart, Plus, Minus, Trash2, User,
		Clock, Printer, X, Banknote, Smartphone,
		Landmark, CreditCard, Receipt, ChevronDown, PlusCircle,
		Percent, FileText
	} from '@lucide/svelte';

	let products = $state<Product[]>([]);
	let customers = $state<Customer[]>([]);
	let heldSales = $state<Sale[]>([]);

	let search = $state('');
	let selectedCategory = $state('');
	let categories = $state<{ id: string; name: string }[]>([]);
	let selectedIndex = $state(-1);

	let paymentMethod = $state<'cash' | 'mpesa' | 'bank' | 'card' | 'credit'>('cash');
	let amountTendered = $state(0);
	let checkingOut = $state(false);
	let searchLoading = $state(false);
	let cartOpen = $state(false);

	let showCustomerModal = $state(false);
	let showHeldModal = $state(false);
	let showDiscountModal = $state(false);
	let showNoteModal = $state(false);
	let showNewCustomerModal = $state(false);
	let customerSearch = $state('');
	let discountInput = $state(0);
	let taxRateInput = $state(0);
	let noteInput = $state('');

	let newCustomerName = $state('');
	let newCustomerPhone = $state('');
	let newCustomerEmail = $state('');

	let lastSale: Sale | null = $state(null);
	let lastAmountTendered = $state(0);
	let lastChange = $state(0);

	let searchInput: HTMLInputElement;
	let searchResultsEl: HTMLDivElement;
	let receiptEl: HTMLDivElement;

	const change = $derived(
		paymentMethod === 'cash' ? Math.max(0, amountTendered - cart.total) : 0
	);

	const payments = [
		{ id: 'cash' as const, label: 'Cash', icon: Banknote },
		{ id: 'mpesa' as const, label: 'M-Pesa', icon: Smartphone },
		{ id: 'bank' as const, label: 'Bank', icon: Landmark },
		{ id: 'card' as const, label: 'Card', icon: CreditCard },
		{ id: 'credit' as const, label: 'Credit', icon: Receipt },
	];

	const paymentIcons = {
		cash: Banknote, mpesa: Smartphone, bank: Landmark, card: CreditCard, credit: Receipt
	};

	function fmt(n: number) {
		return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}

	let debounceTimer: ReturnType<typeof setTimeout>;
	function onSearch() {
		selectedIndex = -1;
		clearTimeout(debounceTimer);
		if (search.trim()) {
			debounceTimer = setTimeout(fetchProducts, 200);
		} else {
			products = [];
			searchLoading = false;
		}
	}

	async function fetchProducts() {
		searchLoading = true;
		try {
			const res = await productsService.list({
				search,
				category_id: selectedCategory || undefined,
				limit: 48
			});
			products = res.data ?? [];
		} finally {
			searchLoading = false;
		}
	}

	async function searchBarcodeOrSKU(query: string) {
		try {
			const res = await productsService.getByBarcode(query);
			if (res.data) {
				addToCart(res.data);
				return true;
			}
		} catch { /* not a barcode */ }

		try {
			const res = await productsService.list({ search: query, limit: 1 });
			if (res.data && res.data.length > 0 &&
				(res.data[0].sku === query || res.data[0].barcode === query)) {
				addToCart(res.data[0]);
				return true;
			}
		} catch { /* ignore */ }

		await fetchProducts();
		return false;
	}

	function handleSearchKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			if (selectedIndex >= 0 && products[selectedIndex]) {
				addToCart(products[selectedIndex]);
				search = '';
				products = [];
				return;
			}
			if (search.trim()) {
				searchBarcodeOrSKU(search.trim());
			}
			search = '';
			products = [];
			return;
		}
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = Math.min(selectedIndex + 1, products.length - 1);
			scrollToSelected();
			return;
		}
		if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = Math.max(selectedIndex - 1, -1);
			scrollToSelected();
			return;
		}
		if (e.key === 'Escape') {
			search = '';
			products = [];
			selectedIndex = -1;
		}
	}

	function scrollToSelected() {
		if (selectedIndex >= 0 && searchResultsEl) {
			const el = searchResultsEl.children[selectedIndex] as HTMLElement;
			el?.scrollIntoView({ block: 'nearest' });
		}
	}

	function addToCart(product: Product) {
		if (product.stock_qty <= 0) {
			notify.error(`"${product.name}" is out of stock`);
			return;
		}
		cart.addProduct(product);
		notify.success(`Added ${product.name}`);
		search = '';
		products = [];
		cartOpen = true;
		searchInput?.focus();
	}

	function setQuantity(productId: string, qty: number) {
		const item = cart.items.find(i => i.product.id === productId);
		if (!item) return;
		if (qty > item.product.stock_qty) {
			notify.error(`Only ${item.product.stock_qty} in stock`);
			qty = item.product.stock_qty;
		}
		cart.updateQuantity(productId, qty);
	}

	async function fetchCustomers() {
		const res = await customersService.list(customerSearch, 1, 50);
		customers = res.data ?? [];
	}

	async function fetchHeld() {
		const res = await salesService.list({ status: 'held', page: 1, limit: 20 });
		heldSales = res.data ?? [];
	}

	function selectCustomer(c: Customer) {
		cart.setCustomer(c.id, c.name);
		showCustomerModal = false;
	}

	function removeCustomer() {
		cart.setCustomer(null, null);
	}

	async function createAndSelectCustomer() {
		if (!newCustomerName.trim()) {
			notify.error('Customer name is required');
			return;
		}
		try {
			const res = await customersService.create({
				name: newCustomerName.trim(),
				phone: newCustomerPhone.trim() || null,
				email: newCustomerEmail.trim() || null,
			});
			if (res.data) {
				selectCustomer(res.data);
				showNewCustomerModal = false;
				newCustomerName = '';
				newCustomerPhone = '';
				newCustomerEmail = '';
				notify.success('Customer created');
			}
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to create customer');
		}
	}

	async function resumeSale(sale: Sale) {
		if (!sale.items) return;
		cart.clear();
		for (const item of sale.items) {
			const res = await productsService.getById(item.product_id);
			if (res.data) {
				cart.addProduct(res.data);
				cart.updateQuantity(item.product_id, item.quantity);
				cart.updatePrice(item.product_id, item.unit_price);
			}
		}
		try { await salesService.void(sale.id); } catch { /* ignore */ }
		showHeldModal = false;
		notify.success('Sale resumed');
	}

	async function holdSale() {
		if (cart.items.length === 0) return;
		try {
			await salesService.create({
				customer_id: cart.customerId,
				items: cart.items.map((i) => ({
					product_id: i.product.id,
					quantity: i.quantity,
					unit_price: i.unit_price
				})),
				discount: cart.discount,
				tax_rate: cart.taxRate,
				payment_method: 'cash',
				status: 'held',
				note: null,
			});
			cart.clear();
			notify.success('Sale held (F5)');
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to hold sale');
		}
	}

	function openCustomerModal() {
		showCustomerModal = true;
		fetchCustomers();
	}

	function openDiscountModal() {
		discountInput = cart.discount;
		taxRateInput = cart.taxRate;
		showDiscountModal = true;
	}

	function openNoteModal() {
		noteInput = cart.note;
		showNoteModal = true;
	}

	function applyDiscount() {
		cart.setDiscount(discountInput);
		cart.setTaxRate(taxRateInput);
		showDiscountModal = false;
	}

	function applyNote() {
		cart.setNote(noteInput);
		showNoteModal = false;
	}

	async function completeSale() {
		if (cart.items.length === 0) {
			notify.error('Cart is empty');
			return;
		}
		if (paymentMethod === 'cash' && amountTendered < cart.total) {
			notify.error('Amount tendered is less than total');
			return;
		}
		checkingOut = true;
		try {
			const res = await salesService.create({
				customer_id: cart.customerId,
				items: cart.items.map((i) => ({
					product_id: i.product.id,
					quantity: i.quantity,
					unit_price: i.unit_price
				})),
				discount: cart.discount,
				tax_rate: cart.taxRate,
				payment_method: paymentMethod,
				status: 'completed',
				note: cart.note || null,
			});
			if (res.data) {
				lastSale = res.data;
				lastAmountTendered = amountTendered;
				lastChange = change;
			}
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Checkout failed');
		} finally {
			checkingOut = false;
		}
	}

	function resetAfterSale() {
		cart.clear();
		amountTendered = 0;
		paymentMethod = 'cash';
		lastSale = null;
		searchInput?.focus();
	}

	function printReceipt() {
		window.print();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (lastSale !== null || showCustomerModal || showHeldModal || showDiscountModal || showNoteModal || showNewCustomerModal) return;
		if (e.key === 'F2') { e.preventDefault(); searchInput?.focus(); }
		if (e.key === 'F4') { e.preventDefault(); openCustomerModal(); }
		if (e.key === 'F5') { e.preventDefault(); holdSale(); }
		if (e.key === 'F6') { e.preventDefault(); document.getElementById('payment-section')?.scrollIntoView({ behavior: 'smooth' }); }
		if (e.key === 'F8' && !checkingOut) { e.preventDefault(); completeSale(); }
		if (e.key === 'Delete' && cart.items.length > 0) {
			const last = cart.items[cart.items.length - 1];
			cart.removeItem(last.product.id);
		}
	}

	function categoryFilter() {
		fetchProducts();
	}

	$effect(() => {
		if (!showCustomerModal && !showHeldModal && lastSale === null && !showDiscountModal && !showNoteModal && !showNewCustomerModal) {
			searchInput?.focus();
		}
	});

	onMount(async () => {
		await fetchProducts();
		const catRes = await productsService.listCategories();
		categories = catRes.data ?? [];
		searchInput?.focus();
	});
</script>

<svelte:head><title>Checkout — POS</title></svelte:head>
<svelte:window onkeydown={handleKeydown} />

<div class="flex h-full bg-slate-50">
	<!-- Mobile cart overlay -->
	{#if cartOpen}
		<div class="fixed inset-0 z-20 bg-black/50 lg:hidden" role="presentation" onclick={() => cartOpen = false}></div>
	{/if}

	<!-- Product panel (left) -->
	<div class="flex flex-1 flex-col overflow-hidden min-w-0">
		<!-- Search bar -->
		<div class="bg-white border-b px-3 py-2.5 flex items-center gap-2 shadow-sm">
			<div class="relative flex-1">
				<Search size={15} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
				<input
					bind:this={searchInput}
					bind:value={search}
					oninput={onSearch}
					onkeydown={handleSearchKeydown}
					placeholder="F2: Search by name, barcode or SKU…"
					class="w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-10 text-sm focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
				/>

			</div>
			<select
				bind:value={selectedCategory}
				onchange={categoryFilter}
				class="rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-blue-500 focus:outline-none bg-white"
			>
				<option value="">All Categories</option>
				{#each categories as cat}
					<option value={cat.id}>{cat.name}</option>
				{/each}
			</select>
		</div>

		<!-- Search results overlay -->
		{#if search && products.length > 0}
			<div class="absolute top-[57px] left-0 right-0 z-10 mx-3 max-h-72 overflow-y-auto rounded-xl border bg-white shadow-xl" bind:this={searchResultsEl}>
				{#each products as product, idx}
					<button
						onclick={() => addToCart(product)}
						onmouseenter={() => selectedIndex = idx}
						class="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-blue-50 transition-colors"
						class:bg-blue-50={selectedIndex === idx}
						class:border-b={idx < products.length - 1}
					>
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-slate-800 truncate">{product.name}</p>
							<p class="text-xs text-slate-400">
								{product.category_name ? `${product.category_name} · ` : ''}
								Stock: {product.stock_qty}
							</p>
						</div>
						<div class="text-right shrink-0">
							<p class="text-sm font-bold text-blue-600">KES {fmt(product.selling_price)}</p>
							{#if product.barcode}
								<p class="text-xs text-slate-400">{product.barcode}</p>
							{/if}
						</div>
					</button>
				{/each}
			</div>
		{/if}

		<!-- Product grid -->
		<div class="flex-1 overflow-y-auto p-3 pb-24 lg:pb-3">
			{#if searchLoading}
				<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 gap-2.5">
					{#each Array(12) as _}
						<div class="rounded-2xl bg-slate-200 animate-pulse h-32"></div>
					{/each}
				</div>
			{:else if products.length === 0 && !search}
				<div class="flex flex-col items-center justify-center h-full text-slate-400 gap-3">
					<ShoppingCart size={48} class="opacity-20" />
					<p class="text-sm">Search products above or scan a barcode</p>
					<div class="flex gap-4 mt-4 text-xs text-slate-300">
						<span class="flex items-center gap-1"><kbd class="rounded border px-1.5 py-0.5 font-mono text-xs">F2</kbd> Search</span>
						<span class="flex items-center gap-1"><kbd class="rounded border px-1.5 py-0.5 font-mono text-xs">F4</kbd> Customer</span>
						<span class="flex items-center gap-1"><kbd class="rounded border px-1.5 py-0.5 font-mono text-xs">F6</kbd> Payment</span>
						<span class="flex items-center gap-1"><kbd class="rounded border px-1.5 py-0.5 font-mono text-xs">F8</kbd> Pay</span>
					</div>
				</div>
			{:else if products.length === 0 && search}
				<div class="flex flex-col items-center justify-center h-48 text-slate-400">
					<Search size={32} class="mb-2 opacity-40" />
					<p class="text-sm">No products matching "{search}"</p>
				</div>
			{:else}
				<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 gap-2.5">
					{#each products as product}
						<button
							onclick={() => addToCart(product)}
							disabled={product.stock_qty === 0}
							class="rounded-2xl border border-slate-200 bg-white p-4 text-left hover:border-blue-400 hover:shadow-md transition-all active:scale-[0.97] disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
						>
							<div class="flex items-start justify-between">
								<p class="text-sm font-semibold text-slate-800 line-clamp-2 leading-tight flex-1">{product.name}</p>
								{#if product.stock_qty <= product.reorder_level && product.stock_qty > 0}
									<span class="ml-1 shrink-0 rounded-full bg-blue-100 px-1.5 py-0.5 text-[10px] font-medium text-blue-600">{product.stock_qty}</span>
								{/if}
							</div>
							{#if product.category_name}
								<p class="text-xs text-slate-400 mt-1">{product.category_name}</p>
							{/if}
							<p class="mt-3 text-lg font-bold text-blue-600">KES {fmt(product.selling_price)}</p>
							<p class="text-xs mt-0.5 {product.stock_qty === 0 ? 'text-red-500 font-medium' : product.stock_qty <= product.reorder_level ? 'text-blue-500 font-medium' : 'text-slate-400'}">
								{product.stock_qty === 0 ? 'Out of stock' : `Stock: ${product.stock_qty}`}
							</p>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>

	<!-- Mobile FAB: View Cart -->
	<button
		onclick={() => cartOpen = true}
		class="fixed bottom-24 left-1/2 -translate-x-1/2 z-10 flex items-center gap-2.5 rounded-full bg-blue-600 px-5 py-3 text-white shadow-xl shadow-blue-600/40 lg:hidden active:scale-95 transition-all"
		class:opacity-0={cart.items.length === 0}
		class:pointer-events-none={cart.items.length === 0}
	>
		<ShoppingCart size={18} />
		<span class="text-sm font-semibold">{cart.count} item{cart.count !== 1 ? 's' : ''} · KES {fmt(cart.total)}</span>
	</button>

	<!-- Cart panel (right) -->
	<div
		class="flex flex-col bg-white border-l border-slate-200 transition-transform duration-300
			fixed inset-x-0 bottom-0 z-30 max-h-[90vh] rounded-t-3xl shadow-2xl
			lg:relative lg:inset-auto lg:z-auto lg:max-h-full lg:w-96 lg:shrink-0 lg:rounded-none lg:shadow-none lg:translate-y-0
			{cartOpen ? 'translate-y-0' : 'translate-y-full'}"
	>
		<!-- Mobile header -->
		<div class="flex items-center justify-between px-4 pt-3 pb-2 lg:hidden border-b">
			<h3 class="font-semibold text-slate-800">Cart ({cart.count})</h3>
			<button onclick={() => cartOpen = false} class="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100">
				<X size={18} />
			</button>
		</div>

		<!-- Customer bar -->
		<div class="border-b px-4 py-2.5 flex items-center justify-between bg-slate-50/50">
			{#if cart.customerName}
				<div class="flex items-center gap-2 text-sm min-w-0">
					<User size={14} class="text-blue-600 shrink-0" />
					<span class="font-medium text-slate-800 truncate">{cart.customerName}</span>
				</div>
				<div class="flex gap-1">
					<button onclick={openCustomerModal} class="text-xs text-blue-600 hover:text-blue-800">Change</button>
					<button onclick={removeCustomer} class="text-slate-400 hover:text-red-500"><X size={13} /></button>
				</div>
			{:else}
				<button onclick={openCustomerModal} class="flex items-center gap-2 text-sm text-slate-500 hover:text-blue-600">
					<User size={14} />
					Walk-in Customer
					<ChevronDown size={12} />
				</button>
			{/if}
		</div>

		<!-- Cart items -->
		<div class="flex-1 overflow-y-auto min-h-0">
			{#if cart.items.length === 0}
				<div class="flex flex-col items-center justify-center h-full text-slate-400 p-8">
					<ShoppingCart size={36} class="mb-2 opacity-30" />
					<p class="text-sm text-center">Cart is empty</p>
					<p class="text-xs text-center mt-1">Click a product or scan a barcode</p>
				</div>
			{:else}
				<ul class="divide-y divide-slate-100">
					{#each cart.items as item (item.product.id)}
						<li class="px-4 py-3 hover:bg-slate-50/50 transition-colors">
							<div class="flex items-start justify-between gap-2">
								<p class="text-sm font-medium text-slate-800 flex-1 leading-tight">{item.product.name}</p>
								<button
									onclick={() => cart.removeItem(item.product.id)}
									class="text-slate-300 hover:text-red-500 shrink-0 mt-0.5"
									title="Remove item"
								>
									<Trash2 size={13} />
								</button>
							</div>
							<div class="flex items-center justify-between mt-2">
								<div class="flex items-center gap-1">
									<button
										onclick={() => setQuantity(item.product.id, item.quantity - 1)}
										class="h-7 w-7 rounded-lg border flex items-center justify-center hover:bg-slate-100 text-slate-600 active:bg-slate-200 transition-colors"
									>
										<Minus size={12} />
									</button>
									<input
										type="number"
										value={item.quantity}
										onchange={(e) => setQuantity(item.product.id, Math.max(1, parseInt((e.target as HTMLInputElement).value) || 1))}
										class="w-10 text-center text-sm font-medium border-0 bg-transparent [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
									/>
									<button
										onclick={() => setQuantity(item.product.id, item.quantity + 1)}
										class="h-7 w-7 rounded-lg border flex items-center justify-center hover:bg-slate-100 text-slate-600 active:bg-slate-200 transition-colors"
									>
										<Plus size={12} />
									</button>
								</div>
								<div class="text-right">
									<p class="text-sm font-bold text-slate-800">KES {fmt(item.unit_price * item.quantity)}</p>
									<p class="text-xs text-slate-400">KES {fmt(item.unit_price)} each</p>
								</div>
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		</div>

		<!-- Totals -->
		<div class="border-t px-4 py-3 space-y-1.5 text-sm bg-slate-50/80">
			<div class="flex justify-between text-slate-600">
				<span>Subtotal</span>
				<span>KES {fmt(cart.subtotal)}</span>
			</div>
			{#if cart.discount > 0}
				<div class="flex justify-between text-green-600">
					<span>Discount</span>
					<span>-KES {fmt(cart.discount)}</span>
				</div>
			{/if}
			{#if cart.taxRate > 0}
				<div class="flex justify-between text-slate-600">
					<span>Tax ({cart.taxRate}%)</span>
					<span>KES {fmt(cart.taxAmount)}</span>
				</div>
			{/if}
			{#if cart.note}
				<div class="flex items-center gap-1 text-xs text-slate-400">
					<FileText size={11} />
					<span class="truncate">{cart.note}</span>
				</div>
			{/if}
			<div class="flex justify-between font-bold text-lg text-slate-900 pt-2 border-t border-slate-200">
				<span>Total</span>
				<span>KES {fmt(cart.total)}</span>
			</div>
		</div>

		<!-- Payment section -->
		<div id="payment-section" class="border-t px-4 py-3 space-y-3">
			<p class="text-xs font-semibold text-slate-500 uppercase tracking-wider">Payment</p>
			<div class="grid grid-cols-5 gap-1.5">
				{#each payments as p}
					<button
						onclick={() => paymentMethod = p.id}
						class="flex flex-col items-center gap-1 rounded-xl border py-2 text-[10px] font-medium transition-all"
						class:bg-blue-600={paymentMethod === p.id}
						class:text-white={paymentMethod === p.id}
						class:border-blue-600={paymentMethod === p.id}
						class:text-slate-600={paymentMethod !== p.id}
						class:border-slate-200={paymentMethod !== p.id}
						class:hover:bg-slate-50={paymentMethod !== p.id}
					>
						{#if p.id === 'cash'}
							<Banknote size={16} />
						{:else if p.id === 'mpesa'}
							<Smartphone size={16} />
						{:else if p.id === 'bank'}
							<Landmark size={16} />
						{:else if p.id === 'card'}
							<CreditCard size={16} />
						{:else if p.id === 'credit'}
							<Receipt size={16} />
						{/if}
						{p.label}
					</button>
				{/each}
			</div>

			{#if paymentMethod === 'cash'}
				<div>
					<label class="block text-xs font-medium text-slate-600 mb-1">Amount Received</label>
					<div class="relative">
						<span class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-sm font-medium">KES</span>
						<input
							type="number"
							bind:value={amountTendered}
							min={0}
							step="0.01"
							placeholder="0.00"
							class="w-full rounded-xl border border-slate-300 py-2.5 pl-11 pr-3 text-lg font-bold text-right focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20"
						/>
					</div>
					{#if amountTendered > 0}
						<div class="mt-2 flex justify-between items-center px-1">
							<span class="text-sm text-slate-500">Change</span>
							<span class="text-lg font-bold {change > 0 ? 'text-green-600' : 'text-red-500'}">
								KES {fmt(change)}
							</span>
						</div>
					{/if}
				</div>
			{:else if paymentMethod === 'mpesa'}
				<div class="rounded-xl bg-green-50 border border-green-200 p-3 text-center text-sm text-green-700">
					<Smartphone size={20} class="mx-auto mb-1" />
					<p class="font-medium">M-Pesa Payment</p>
					<p class="text-xs mt-1">Enter M-Pesa transaction code after payment</p>
				</div>
			{/if}

			<!-- Action buttons -->
			<div class="grid grid-cols-3 gap-2 pt-1">
				<button onclick={openDiscountModal} class="flex items-center justify-center gap-1.5 rounded-xl border border-slate-200 py-2.5 text-xs font-medium text-slate-600 hover:bg-slate-50">
					<Percent size={13} /> Discount
				</button>
				<button onclick={openNoteModal} class="flex items-center justify-center gap-1.5 rounded-xl border border-slate-200 py-2.5 text-xs font-medium text-slate-600 hover:bg-slate-50">
					<FileText size={13} /> Notes
				</button>
				<button onclick={holdSale} disabled={cart.items.length === 0} class="flex items-center justify-center gap-1.5 rounded-xl border border-slate-200 py-2.5 text-xs font-medium text-slate-600 hover:bg-slate-50 disabled:opacity-40">
					<Clock size={13} /> Hold
				</button>
			</div>

			<div class="grid grid-cols-2 gap-2">
				<button
					onclick={() => cart.clear()}
					disabled={cart.items.length === 0}
					class="rounded-xl border border-red-200 py-3 text-sm font-medium text-red-600 hover:bg-red-50 disabled:opacity-40 transition-colors"
				>
					Clear
				</button>
				<button
					onclick={completeSale}
					disabled={cart.items.length === 0 || checkingOut}
					class="rounded-xl bg-blue-600 py-3 text-sm font-bold text-white hover:bg-blue-700 disabled:opacity-50 transition-all active:scale-[0.98] shadow-lg shadow-blue-600/20"
				>
					{checkingOut ? 'Processing…' : `Pay KES ${fmt(cart.total)}`}
				</button>
			</div>
		</div>
	</div>
</div>

<!-- Customer modal -->
<Modal open={showCustomerModal} title="Select Customer" onclose={() => showCustomerModal = false}>
	{#snippet children()}
		<div class="space-y-3">
			<div class="relative">
				<Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
				<input
					bind:value={customerSearch}
					oninput={fetchCustomers}
					placeholder="F4: Search customers…"
					class="w-full rounded-lg border px-3 py-2 pl-9 text-sm focus:border-blue-500 focus:outline-none"
				/>
			</div>
			<ul class="max-h-60 overflow-y-auto divide-y rounded-lg border">
				{#each customers as c}
					<li>
						<button
							onclick={() => selectCustomer(c)}
							class="w-full px-4 py-3 text-left hover:bg-blue-50 text-sm flex items-center gap-3"
						>
							<div class="flex h-8 w-8 items-center justify-center rounded-full bg-blue-100 text-blue-600 font-semibold text-xs shrink-0">
								{c.name.charAt(0).toUpperCase()}
							</div>
							<div class="min-w-0">
								<p class="font-medium text-slate-900 truncate">{c.name}</p>
								{#if c.phone}<p class="text-slate-400 text-xs">{c.phone}</p>{/if}
							</div>
						</button>
					</li>
				{:else}
					<li class="px-4 py-8 text-center text-sm text-slate-400">No customers found</li>
				{/each}
			</ul>
			<button
				onclick={() => { showCustomerModal = false; showNewCustomerModal = true; }}
				class="w-full flex items-center justify-center gap-2 rounded-lg border border-dashed border-slate-300 py-2.5 text-sm text-slate-600 hover:bg-slate-50"
			>
				<PlusCircle size={15} />
				New Customer
			</button>
		</div>
	{/snippet}
</Modal>

<!-- New Customer modal -->
<Modal open={showNewCustomerModal} title="New Customer" onclose={() => showNewCustomerModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-4">
			<div>
				<label class="block text-sm font-medium text-slate-700 mb-1">Name *</label>
				<input
					bind:value={newCustomerName}
					placeholder="Customer name"
					class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium text-slate-700 mb-1">Phone</label>
				<input
					bind:value={newCustomerPhone}
					placeholder="Phone number"
					class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium text-slate-700 mb-1">Email</label>
				<input
					bind:value={newCustomerEmail}
					type="email"
					placeholder="Email address"
					class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
				/>
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showNewCustomerModal = false} class="rounded-lg border px-4 py-2 text-sm text-slate-700 hover:bg-slate-50">Cancel</button>
		<button onclick={createAndSelectCustomer} class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700">Create & Select</button>
	{/snippet}
</Modal>

<!-- Held sales modal -->
<Modal open={showHeldModal} title="Held Sales" onclose={() => showHeldModal = false}>
	{#snippet children()}
		<ul class="space-y-2 max-h-72 overflow-y-auto">
			{#each heldSales as sale}
				<li class="rounded-lg border p-4 flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-slate-900">{new Date(sale.created_at).toLocaleString()}</p>
						<p class="text-sm text-slate-500">{sale.items?.length ?? 0} items · KES {fmt(sale.total)}</p>
					</div>
					<button
						onclick={() => resumeSale(sale)}
						class="rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
					>
						Resume
					</button>
				</li>
			{:else}
				<li class="text-center text-sm text-slate-400 py-8">No held sales</li>
			{/each}
		</ul>
	{/snippet}
	{#snippet footer()}
		<button onclick={fetchHeld} class="rounded-lg border px-4 py-2 text-sm text-slate-700 hover:bg-slate-50">Refresh</button>
	{/snippet}
</Modal>

<!-- Discount modal -->
<Modal open={showDiscountModal} title="Discount & Tax" onclose={() => showDiscountModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-4">
			<div>
				<label class="block text-sm font-medium text-slate-700 mb-1">Discount (KES)</label>
				<input
					type="number"
					bind:value={discountInput}
					min="0"
					class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium text-slate-700 mb-1">Tax Rate (%)</label>
				<input
					type="number"
					bind:value={taxRateInput}
					min="0"
					max="100"
					step="0.5"
					class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
				/>
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showDiscountModal = false} class="rounded-lg border px-4 py-2 text-sm text-slate-700 hover:bg-slate-50">Cancel</button>
		<button onclick={applyDiscount} class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700">Apply</button>
	{/snippet}
</Modal>

<!-- Note modal -->
<Modal open={showNoteModal} title="Sale Notes" onclose={() => showNoteModal = false} size="sm">
	{#snippet children()}
		<div>
			<textarea
				bind:value={noteInput}
				rows={3}
				placeholder="Add a note to this sale…"
				class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none resize-none"
			></textarea>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showNoteModal = false} class="rounded-lg border px-4 py-2 text-sm text-slate-700 hover:bg-slate-50">Cancel</button>
		<button onclick={applyNote} class="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700">Save Note</button>
	{/snippet}
</Modal>

<!-- Receipt overlay -->
{#if lastSale}
	<div class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/50 backdrop-blur-sm p-0 sm:p-4" role="dialog" aria-modal="true">
		<div class="relative w-full max-w-sm max-h-[92vh] flex flex-col rounded-t-3xl sm:rounded-2xl bg-white shadow-2xl overflow-hidden" role="presentation">
			<div class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
				<h2 class="text-base font-semibold text-slate-900">Sale Complete</h2>
			</div>
			<div class="px-5 py-4 overflow-y-auto flex-1">
				<div bind:this={receiptEl} class="text-center space-y-4 receipt-print">
					<div class="border-b pb-3">
						<h3 class="text-lg font-bold text-slate-900">POS System</h3>
						<p class="text-xs text-slate-500">Receipt</p>
					</div>

					<div class="text-xs text-slate-500 text-left space-y-1">
						<div class="flex justify-between">
							<span>Date:</span>
							<span>{new Date(lastSale.created_at).toLocaleString()}</span>
						</div>
						<div class="flex justify-between">
							<span>Receipt:</span>
							<span class="font-mono">{lastSale.id.slice(0, 8).toUpperCase()}</span>
						</div>
						<div class="flex justify-between">
							<span>Cashier:</span>
							<span>{lastSale.cashier_name || '-'}</span>
						</div>
						{#if lastSale.customer_name}
							<div class="flex justify-between">
								<span>Customer:</span>
								<span>{lastSale.customer_name}</span>
							</div>
						{/if}
					</div>

					<table class="w-full text-xs text-left border-t border-b">
						<thead>
							<tr class="text-slate-500">
								<th class="py-1.5">Item</th>
								<th class="py-1.5 text-center">Qty</th>
								<th class="py-1.5 text-right">Price</th>
								<th class="py-1.5 text-right">Total</th>
							</tr>
						</thead>
						<tbody>
							{#each lastSale.items || [] as item}
								<tr class="border-t border-dashed">
									<td class="py-1.5 pr-2">{item.product_name || item.product_id.slice(0, 8)}</td>
									<td class="py-1.5 text-center">{item.quantity}</td>
									<td class="py-1.5 text-right">KES {fmt(item.unit_price)}</td>
									<td class="py-1.5 text-right font-medium">KES {fmt(item.total)}</td>
								</tr>
							{/each}
						</tbody>
					</table>

					<div class="text-sm space-y-1">
						<div class="flex justify-between text-slate-600">
							<span>Subtotal</span>
							<span>KES {fmt(lastSale.subtotal)}</span>
						</div>
						{#if lastSale.discount > 0}
							<div class="flex justify-between text-green-600">
								<span>Discount</span>
								<span>-KES {fmt(lastSale.discount)}</span>
							</div>
						{/if}
						{#if lastSale.tax > 0}
							<div class="flex justify-between text-slate-600">
								<span>Tax</span>
								<span>KES {fmt(lastSale.tax)}</span>
							</div>
						{/if}
						<div class="flex justify-between font-bold text-slate-900 border-t pt-1">
							<span>Total</span>
							<span>KES {fmt(lastSale.total)}</span>
						</div>
						<div class="flex justify-between text-slate-500 text-xs border-t pt-1">
							<span>Payment</span>
							<span class="capitalize">{lastSale.payment_method}</span>
						</div>
						{#if lastSale.payment_method === 'cash'}
							<div class="flex justify-between text-slate-500 text-xs">
								<span>Amount Tendered</span>
								<span>KES {fmt(lastAmountTendered)}</span>
							</div>
							<div class="flex justify-between font-bold text-green-600 text-xs">
								<span>Change</span>
								<span>KES {fmt(lastChange)}</span>
							</div>
						{/if}
					</div>

					<div class="border-t pt-3 text-xs text-slate-400">
						<p>Thank you for your purchase!</p>
					</div>
				</div>
			</div>
			<div class="flex items-center justify-between gap-2 border-t border-slate-100 px-5 py-4">
				<button onclick={printReceipt} class="flex items-center gap-2 rounded-lg border px-4 py-2 text-sm text-slate-700 hover:bg-slate-50">
					<Printer size={15} /> Print
				</button>
				<button onclick={resetAfterSale} class="rounded-lg bg-blue-600 px-6 py-2 text-sm font-semibold text-white hover:bg-blue-700">
					New Sale
				</button>
			</div>
		</div>
	</div>
{/if}
<style>
	@media print {
		:global(body *) { visibility: hidden; }
		:global(.receipt-print), :global(.receipt-print *) { visibility: visible; }
		:global(.receipt-print) { position: absolute; left: 0; top: 0; width: 80mm; }
	}
</style>
