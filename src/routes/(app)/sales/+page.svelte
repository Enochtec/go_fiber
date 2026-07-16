<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { cart } from '$lib/stores/cart.svelte';
	import { shiftStore } from '$lib/stores/shift.svelte';
	import { notify } from '$lib/stores/notification.svelte';
	import { productsService } from '$lib/services/products';
	import { customersService } from '$lib/services/customers';
	import { salesService } from '$lib/services/sales';
	import Modal from '$lib/components/Modal.svelte';
	import type { Product, Customer, Sale } from '$lib/types';
	import {
		Search, ShoppingCart, Plus, Minus, Trash2, User,
		Clock, Printer, X, Banknote, Smartphone, Package,
		Landmark, CreditCard, Receipt, ChevronDown, PlusCircle,
		Percent, FileText, Tag, AlertTriangle, Check,
		MessageCircle, ArrowLeft, Hash
	} from '@lucide/svelte';

	// ─── State ───────────────────────────────────────────────────────
	let products = $state<Product[]>([]);
	let allProducts = $state<Product[]>([]);
	let searchResults = $state<Product[]>([]);
	let customers = $state<Customer[]>([]);
	let heldSales = $state<Sale[]>([]);

	let search = $state('');
	let selectedCategory = $state('');
	let categories = $state<{ id: string; name: string }[]>([]);
	let selectedIndex = $state(-1);
	let showSearchDropdown = $state(false);

	let paymentMethod = $state<'cash' | 'mpesa' | 'bank' | 'card' | 'credit'>('cash');
	let amountTendered = $state(0);
	let checkingOut = $state(false);
	let productsLoading = $state(false);
	let cartOpen = $state(false);

	let showCustomerModal = $state(false);
	let showHeldModal = $state(false);
	let showDiscountModal = $state(false);
	let showNoteModal = $state(false);
	let showNewCustomerModal = $state(false);
	let showShiftModal = $state(false);
	let showCloseShiftModal = $state(false);

	let customerSearch = $state('');
	let discountInput = $state(0);
	let taxRateInput = $state(0);
	let noteInput = $state('');
	let openingFloat = $state(0);
	let shiftNotes = $state('');
	let closingCash = $state(0);

	let newCustomerName = $state('');
	let newCustomerPhone = $state('');
	let newCustomerEmail = $state('');

	let lastSale: Sale | null = $state(null);
	let lastAmountTendered = $state(0);
	let lastChange = $state(0);
	let whatsappNumber = $state('');
	let clockStr = $state('');

	let searchInput: HTMLInputElement;
	let searchResultsEl: HTMLDivElement;

	let clockTimer: ReturnType<typeof setInterval>;

	function updateClock() {
		const d = new Date();
		const opts: Intl.DateTimeFormatOptions = {
			weekday: 'short', year: 'numeric', month: 'short', day: 'numeric',
			hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: true
		};
		clockStr = d.toLocaleDateString('en-US', opts);
	}

	onDestroy(() => clearInterval(clockTimer));

	// ─── Derived ────────────────────────────────────────────────────
	const change = $derived(
		paymentMethod === 'cash' ? Math.max(0, amountTendered - cart.total) : 0
	);

	const payments = [
		{ id: 'cash' as const, label: 'Cash', icon: Banknote },
		{ id: 'mpesa' as const, label: 'M-Pesa', icon: Smartphone },
		{ id: 'card' as const, label: 'Card', icon: CreditCard },
		{ id: 'bank' as const, label: 'Bank', icon: Landmark },
		{ id: 'credit' as const, label: 'Credit', icon: Receipt },
	];

	// ─── Formatters ─────────────────────────────────────────────────
	function fmt(n: number) {
		return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}

	// ─── Products ───────────────────────────────────────────────────
	let searchTimer: ReturnType<typeof setTimeout>;
	let searchReqId = $state(0);
	let searching = $state(false);

	function onSearch() {
		selectedIndex = -1;
		clearTimeout(searchTimer);
		if (search.trim()) {
			showSearchDropdown = true;
			searchTimer = setTimeout(doSearch, 50);
		} else {
			showSearchDropdown = false;
			searching = false;
			searchResults = [];
		}
	}

	async function doSearch() {
		const q = search.trim();
		if (!q) return;
		const id = ++searchReqId;
		searching = true;
		const res = await productsService.list({ search: q, limit: 8 });
		if (id !== searchReqId) return;
		searchResults = res.data ?? [];
		searching = false;
	}

	async function loadProducts() {
		productsLoading = true;
		try {
			const res = await productsService.list({
				category_id: selectedCategory || undefined,
				limit: 60
			});
			allProducts = res.data ?? [];
			if (!search.trim()) products = allProducts;
		} finally {
			productsLoading = false;
		}
	}

	async function searchBarcodeOrSKU(query: string) {
		try {
			const res = await productsService.getByBarcode(query);
			if (res.data) { addToCart(res.data); return; }
		} catch { /* not a barcode */ }
		const q = query.trim();
		if (!q) return;
		const id = ++searchReqId;
		const res = await productsService.list({ search: q, limit: 8 });
		if (id !== searchReqId) return;
		searchResults = res.data ?? [];
	}

	function handleSearchKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			if (selectedIndex >= 0 && searchResults[selectedIndex]) {
				addToCart(searchResults[selectedIndex]);
				return;
			}
			if (search.trim()) searchBarcodeOrSKU(search.trim());
			closeSearch();
			return;
		}
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = Math.min(selectedIndex + 1, searchResults.length - 1);
			scrollToSelected();
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = Math.max(selectedIndex - 1, -1);
			scrollToSelected();
		} else if (e.key === 'Escape') {
			closeSearch();
		}
	}

	function scrollToSelected() {
		if (selectedIndex >= 0 && searchResultsEl) {
			const el = searchResultsEl.children[selectedIndex] as HTMLElement;
			el?.scrollIntoView({ block: 'nearest' });
		}
	}

	function closeSearch() {
		search = '';
		showSearchDropdown = false;
		searchResults = [];
		selectedIndex = -1;
	}

	function addToCart(product: Product) {
		if (product.stock_qty <= 0) {
			notify.error(`"${product.name}" is out of stock`);
			return;
		}
		cart.addProduct(product);
		closeSearch();
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

	// ─── Customers ──────────────────────────────────────────────────
	let customerSearchTimer: ReturnType<typeof setTimeout>;
	let customerSearchId = $state(0);

	async function fetchCustomers() {
		const res = await customersService.list(customerSearch, 1, 50);
		customers = res.data ?? [];
	}

	function onCustomerSearch() {
		clearTimeout(customerSearchTimer);
		customerSearchTimer = setTimeout(doCustomerSearch, 80);
	}

	async function doCustomerSearch() {
		const q = customerSearch.trim();
		const id = ++customerSearchId;
		const res = await customersService.list(q, 1, 50);
		if (id !== customerSearchId) return;
		customers = res.data ?? [];
	}

	function selectCustomer(c: Customer) {
		cart.setCustomer(c.id, c.name);
		if (c.phone && !whatsappNumber) whatsappNumber = c.phone;
		showCustomerModal = false;
	}

	function removeCustomer() { cart.setCustomer(null, null); }

	async function createNewCustomer() {
		if (!newCustomerName.trim()) return;
		try {
			const res = await customersService.create({
				name: newCustomerName, phone: newCustomerPhone, email: newCustomerEmail, address: ''
			});
			if (res.data) {
				if (newCustomerPhone) whatsappNumber = newCustomerPhone;
				selectCustomer(res.data);
				showNewCustomerModal = false;
				newCustomerName = ''; newCustomerPhone = ''; newCustomerEmail = '';
			}
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to create customer');
		}
	}

	// ─── Held sales ─────────────────────────────────────────────────
	async function fetchHeld() {
		const res = await salesService.list({ status: 'held', page: 1, limit: 20 });
		heldSales = res.data ?? [];
	}

	async function holdSale() {
		if (cart.items.length === 0) return;
		try {
			const res = await salesService.create({
				customer_id: cart.customerId,
				items: cart.items.map(i => ({ product_id: i.product.id, quantity: i.quantity, unit_price: i.unit_price })),
				discount: cart.discount, tax_rate: cart.taxRate,
				payment_method: paymentMethod, status: 'held', note: cart.note || null,
			});
			if (res.data) { cart.clear(); notify.success('Sale held'); }
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to hold sale');
		}
	}

	async function resumeSale(sale: Sale) {
		cart.clear();
		for (const item of sale.items ?? []) {
			try {
				const p = await productsService.getById(item.product_id);
				if (p.data) { cart.addProduct(p.data); cart.updateQuantity(p.data.id, item.quantity); }
			} catch { /* skip */ }
		}
		if (sale.customer_id && sale.customer_name)
			cart.setCustomer(sale.customer_id, sale.customer_name);
		cart.setDiscount(sale.discount ?? 0);
		showHeldModal = false;
		await salesService.void(sale.id);
	}

	// ─── Checkout ───────────────────────────────────────────────────
	function openCustomerModal() {
		customerSearch = '';
		fetchCustomers();
		showCustomerModal = true;
	}

	function openDiscountModal() {
		discountInput = cart.discount;
		taxRateInput = cart.taxRate;
		showDiscountModal = true;
	}

	function applyDiscount() {
		cart.setDiscount(discountInput);
		cart.setTaxRate(taxRateInput);
		showDiscountModal = false;
	}

	function applyNote() { cart.setNote(noteInput); showNoteModal = false; }

	async function completeSale() {
		if (cart.items.length === 0) { notify.error('Cart is empty'); return; }
		if (paymentMethod === 'cash' && amountTendered < cart.total) {
			notify.error('Amount tendered is less than total');
			return;
		}
		checkingOut = true;
		try {
			const res = await salesService.create({
				customer_id: cart.customerId,
				items: cart.items.map(i => ({ product_id: i.product.id, quantity: i.quantity, unit_price: i.unit_price })),
				discount: cart.discount, tax_rate: cart.taxRate,
				payment_method: paymentMethod, status: 'completed', note: cart.note || null,
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
		whatsappNumber = '';
		lastSale = null;
		loadProducts();
		searchInput?.focus();
	}

	function printReceipt() { window.print(); }

	function formatPhone(num: string) {
		const d = num.replace(/\D/g, '');
		if (d.startsWith('0')) return '254' + d.slice(1);
		if (d.startsWith('254')) return d;
		if (d.startsWith('+')) return d.slice(1);
		return d;
	}

	function sendWhatsApp() {
		if (!lastSale) return;
		const num = whatsappNumber.trim();
		if (!num) { notify.error('Enter a phone number for WhatsApp'); return; }
		const items = (lastSale.items ?? []).map(i =>
			`• ${i.product_name ?? 'Item'} x${i.quantity} = KES ${fmt(i.total)}`
		).join('\n');
		const msg = `*POS Receipt*\nReceipt: #${lastSale.id.slice(0, 8).toUpperCase()}\nDate: ${new Date(lastSale.created_at).toLocaleString()}\n\n${items}\n\nSubtotal: KES ${fmt(lastSale.subtotal)}\nTotal: *KES ${fmt(lastSale.total)}*\nPayment: ${lastSale.payment_method}\n\nThank you for your purchase! 🙏`;
		window.open(`https://wa.me/${formatPhone(num)}?text=${encodeURIComponent(msg)}`, '_blank');
	}

	// ─── Shift ──────────────────────────────────────────────────────
	async function openShift() {
		try {
			await shiftStore.open(openingFloat, shiftNotes);
			showShiftModal = false;
			openingFloat = 0;
			shiftNotes = '';
			notify.success('Shift opened');
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to open shift');
		}
	}

	async function closeShift() {
		try {
			const shift = await shiftStore.close(closingCash, shiftNotes);
			showCloseShiftModal = false;
			notify.success(`Shift closed. Variance: KES ${fmt(shift.variance ?? 0)}`);
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to close shift');
		}
	}

	// ─── Keyboard ───────────────────────────────────────────────────
	function handleKeydown(e: KeyboardEvent) {
		if (lastSale !== null || showCustomerModal || showHeldModal || showDiscountModal || showNoteModal || showNewCustomerModal) return;
		if (e.key === 'F2') { e.preventDefault(); searchInput?.focus(); }
		if (e.key === 'F4') { e.preventDefault(); openCustomerModal(); }
		if (e.key === 'F5') { e.preventDefault(); holdSale(); }
		if (e.key === 'F8' && !checkingOut) { e.preventDefault(); completeSale(); }
	}

	$effect(() => {
		if (!showCustomerModal && !showHeldModal && lastSale === null && !showDiscountModal && !showNoteModal && !showNewCustomerModal) {
			searchInput?.focus();
		}
	});

	onMount(async () => {
		updateClock();
		clockTimer = setInterval(updateClock, 1000);
		await shiftStore.fetch();
		const catRes = await productsService.listCategories();
		categories = catRes.data ?? [];
		await loadProducts();
		searchInput?.focus();
	});
</script>

<svelte:head><title>Checkout — POS</title></svelte:head>
<svelte:window onkeydown={handleKeydown} />

<!-- ─── Main layout ───────────────────────────────────────────────── -->
<div class="flex h-full overflow-hidden bg-slate-100">

	<!-- Mobile cart overlay -->
	{#if cartOpen}
		<div class="fixed inset-0 z-20 bg-black/60 lg:hidden" role="presentation" onclick={() => cartOpen = false}></div>
	{/if}

	<!-- ── LEFT: Product Browser ─────────────────────────────────── -->
	<div class="flex flex-1 flex-col overflow-hidden min-w-0">

		<!-- Unified header -->
		<div class="sticky top-0 z-30 shrink-0 bg-white border-b border-slate-200 shadow-sm">
			<!-- Top row: shift status + clock + search + actions -->
			<div class="flex items-center gap-2 px-3 py-2">
				<!-- Shift status -->
				{#if shiftStore.checked}
					{#if shiftStore.isOpen}
						<div class="flex items-center gap-2 shrink-0">
							<span class="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></span>
							<span class="text-xs font-semibold text-emerald-700 hidden sm:inline">Shift Open</span>
							<button
								onclick={() => showCloseShiftModal = true}
								class="rounded-lg px-2.5 py-1.5 text-xs font-semibold text-white transition-all active:scale-95 bg-red-500 hover:bg-red-600 shrink-0"
							>
								Close
							</button>
						</div>
					{:else}
						<div class="flex items-center gap-2 shrink-0">
							<div class="h-2 w-2 rounded-full bg-amber-400"></div>
							<span class="text-xs font-semibold text-amber-700 hidden sm:inline">No Shift</span>
							<button
								onclick={() => showShiftModal = true}
								class="rounded-lg px-2.5 py-1.5 text-xs font-semibold text-white transition-all active:scale-95 shrink-0"
								style="background-color:#008B8B;"
							>
								Open
							</button>
						</div>
					{/if}
				{/if}

				<!-- Digital clock -->
				<div class="hidden md:flex items-center gap-1.5 text-xs text-slate-500 font-mono ml-1 shrink-0" aria-live="polite">
					<span id="clock-display">{clockStr}</span>
				</div>

				<!-- Search -->
				<div class="relative flex-1 min-w-0">
					<Search size={14} class="absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
					<input
						bind:this={searchInput}
						bind:value={search}
						oninput={onSearch}
						onkeydown={handleSearchKeydown}
						onfocus={() => { if (search) showSearchDropdown = true; }}
						placeholder="Search product… (F2)"
						class="w-full rounded-lg border border-slate-200 bg-slate-50 py-2 pl-8 pr-3 text-sm focus:outline-none focus:border-slate-300 focus:bg-white transition-colors"
					/>
					{#if search}
						<button onclick={closeSearch} class="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600">
							<X size={13} />
						</button>
					{/if}
				</div>

				<!-- Held sales -->
				<button
					onclick={async () => { await fetchHeld(); showHeldModal = true; }}
					class="relative flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-500 hover:bg-slate-50 shrink-0"
					title="Held Sales"
				>
					<Clock size={15} />
					{#if heldSales.length > 0}
						<span class="absolute -top-1 -right-1 h-4 w-4 rounded-full text-[9px] font-bold text-white flex items-center justify-center" style="background-color:#008B8B;">{heldSales.length}</span>
					{/if}
				</button>
			</div>
			<!-- Mobile clock row -->
			<div class="md:hidden flex items-center justify-center px-3 pb-1.5">
				<span class="text-[10px] text-slate-400 font-mono">{clockStr}</span>
			</div>
		</div>

		<!-- Search dropdown -->
		{#if showSearchDropdown}
			<div
				bind:this={searchResultsEl}
				class="absolute top-[52px] left-0 right-0 lg:right-[400px] z-20 mx-3 max-h-72 overflow-y-auto rounded-xl border border-slate-200 bg-white shadow-xl"
			>
				{#if searching}
					<div class="flex items-center justify-center gap-2 px-4 py-6 text-sm text-slate-400">
						<span class="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-slate-500"></span>
						Searching…
					</div>
				{:else if searchResults.length === 0}
					<div class="px-4 py-6 text-center text-sm text-slate-400">No products found</div>
				{:else}
					{#each searchResults as product, idx}
						<button
							onclick={() => addToCart(product)}
							onmouseenter={() => selectedIndex = idx}
							class="w-full flex items-center gap-3 px-4 py-3 text-left transition-colors border-b border-slate-100 last:border-0"
							class:bg-blue-50={selectedIndex === idx}
						>
							<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-slate-100">
								{#if product.image_url}
									<img src={product.image_url} alt={product.name} class="h-full w-full object-cover rounded-lg" />
								{:else}
									<Package size={14} class="text-slate-400" />
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-sm font-semibold text-slate-800 truncate">{product.name}</p>
								<p class="text-xs text-slate-400">
									{product.category_name ? `${product.category_name} · ` : ''}Stock: {product.stock_qty}
								</p>
							</div>
							<div class="text-right shrink-0">
								<p class="text-sm font-bold" style="color:#008B8B;">KES {fmt(product.selling_price)}</p>
								{#if product.barcode}<p class="text-xs text-slate-400 font-mono">{product.barcode}</p>{/if}
							</div>
						</button>
					{/each}
				{/if}
			</div>
		{/if}

		<!-- Category pills -->
		<div class="flex gap-2 px-3 py-2 overflow-x-auto shrink-0 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 scrollbar-none">
			<button
				onclick={() => { selectedCategory = ''; loadProducts(); }}
				class="inline-flex shrink-0 items-center rounded-full px-3.5 py-1.5 text-xs font-semibold transition-all"
				style={selectedCategory === '' ? 'background-color:#008B8B; color:white;' : 'background-color:#f1f5f9; color:#64748b;'}
			>
				All
			</button>
			{#each categories as cat}
				<button
					onclick={() => { selectedCategory = cat.id; loadProducts(); }}
					class="inline-flex shrink-0 items-center rounded-full px-3.5 py-1.5 text-xs font-semibold transition-all"
					style={selectedCategory === cat.id ? 'background-color:#008B8B; color:white;' : 'background-color:#f1f5f9; color:#64748b;'}
				>
					{cat.name}
				</button>
			{/each}
		</div>

		<!-- Product grid -->
		<div class="flex-1 overflow-y-auto p-3 pb-24 lg:pb-3">
			{#if productsLoading}
				<div class="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 gap-2.5">
					{#each Array(12) as _}
						<div class="rounded-2xl bg-white dark:bg-slate-800 h-36 animate-pulse border border-slate-100 dark:border-slate-700"></div>
					{/each}
				</div>
			{:else if products.length === 0}
				<div class="flex flex-col items-center justify-center h-full gap-3 text-slate-400">
					<Package size={48} class="opacity-20" />
					<p class="text-sm font-medium">No products found</p>
					<div class="flex gap-3 text-xs text-slate-300 mt-2">
						<span><kbd class="rounded border border-slate-200 px-1.5 py-0.5 font-mono text-xs">F2</kbd> Search</span>
						<span><kbd class="rounded border border-slate-200 px-1.5 py-0.5 font-mono text-xs">F8</kbd> Pay</span>
					</div>
				</div>
			{:else}
				<div class="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 gap-2.5">
					{#each products as product}
						<button
							onclick={() => addToCart(product)}
							disabled={product.stock_qty === 0}
							class="group relative rounded-2xl border bg-white dark:bg-slate-800 p-3 text-left transition-all active:scale-[0.97] disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md hover:-translate-y-0.5 flex flex-col"
							style={product.stock_qty === 0 ? '' : 'border-color:#e2e8f0;'}
						>
							<!-- Product image -->
							<div class="relative w-full aspect-square rounded-xl overflow-hidden bg-slate-100 dark:bg-slate-700 mb-2.5 flex items-center justify-center">
								{#if product.image_url}
									<img src={product.image_url} alt={product.name} class="w-full h-full object-cover" />
								{:else}
									<Package size={28} class="text-slate-300 dark:text-slate-500" />
								{/if}
								<!-- Stock badge -->
								{#if product.stock_qty === 0}
									<div class="absolute inset-0 flex items-center justify-center bg-white/80 dark:bg-slate-800/80">
										<span class="rounded-full bg-red-100 px-2 py-0.5 text-[10px] font-bold text-red-700">Out of Stock</span>
									</div>
								{:else if product.stock_qty <= product.reorder_level}
									<span class="absolute top-1.5 right-1.5 rounded-full bg-amber-100 px-1.5 py-0.5 text-[9px] font-bold text-amber-700">Low</span>
								{/if}
							</div>

							<p class="text-xs font-semibold text-slate-800 dark:text-slate-100 line-clamp-2 leading-tight flex-1">{product.name}</p>
							{#if product.category_name}
								<p class="text-[10px] text-slate-400 mt-0.5 truncate">{product.category_name}</p>
							{/if}
							<div class="flex items-center justify-between mt-2">
								<p class="text-sm font-bold" style="color:#008B8B;">KES {fmt(product.selling_price)}</p>
								<span class="text-[10px] text-slate-400">{product.stock_qty} left</span>
							</div>

							<!-- Add overlay on hover -->
							<div class="absolute inset-0 rounded-2xl flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" style="background-color:rgba(0,139,139,0.08);">
								<div class="rounded-full text-white p-1.5" style="background-color:#008B8B;">
									<Plus size={14} />
								</div>
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>

	<!-- Mobile FAB -->
	<button
		onclick={() => cartOpen = true}
		class="fixed bottom-24 left-1/2 -translate-x-1/2 z-10 flex items-center gap-2.5 rounded-full px-5 py-3 text-white shadow-xl lg:hidden active:scale-95 transition-all"
		style="background-color:#008B8B; box-shadow:0 10px 25px rgba(0,139,139,0.4);"
		class:opacity-0={cart.items.length === 0}
		class:pointer-events-none={cart.items.length === 0}
	>
		<ShoppingCart size={18} />
		<span class="text-sm font-semibold">{cart.count} item{cart.count !== 1 ? 's' : ''} · KES {fmt(cart.total)}</span>
	</button>

	<!-- ── RIGHT: Cart + Payment Panel ───────────────────────────── -->
	<div
		class="flex flex-col bg-white dark:bg-slate-900 border-l border-slate-200 dark:border-slate-700 transition-transform duration-300
			fixed inset-x-0 bottom-0 z-30 max-h-[92vh] rounded-t-3xl shadow-2xl
			lg:relative lg:inset-auto lg:z-auto lg:max-h-full lg:w-[400px] lg:shrink-0 lg:rounded-none lg:shadow-none lg:translate-y-0
			{cartOpen ? 'translate-y-0' : 'translate-y-full'}"
	>
		<!-- Mobile drag handle -->
		<div class="flex justify-center pt-3 pb-1 lg:hidden shrink-0">
			<div class="h-1 w-10 rounded-full bg-slate-200 dark:bg-slate-600"></div>
		</div>

		<!-- Cart header -->
		<div class="flex items-center justify-between px-4 py-2.5 border-b border-slate-100 dark:border-slate-700 shrink-0">
			<div class="flex items-center gap-2">
				<ShoppingCart size={16} class="text-slate-500" />
				<span class="font-semibold text-slate-800 dark:text-slate-100 text-sm">Order</span>
				{#if cart.count > 0}
					<span class="inline-flex h-5 min-w-5 items-center justify-center rounded-full text-[10px] font-bold text-white px-1" style="background-color:#008B8B;">{cart.count}</span>
				{/if}
			</div>
			<div class="flex items-center gap-1">
				<button onclick={openCustomerModal} class="flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs font-medium transition-colors border border-slate-200 dark:border-slate-600 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">
					<User size={12} />
					{cart.customerName ?? 'Walk-in'}
					<ChevronDown size={10} />
				</button>
				{#if cart.customerName}
					<button onclick={removeCustomer} class="h-7 w-7 flex items-center justify-center rounded-lg text-slate-400 hover:text-red-500 hover:bg-red-50 transition-colors">
						<X size={13} />
					</button>
				{/if}
				<button onclick={() => cartOpen = false} class="lg:hidden h-7 w-7 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100">
					<X size={15} />
				</button>
			</div>
		</div>

		<!-- Cart items (scrollable) -->
		<div class="flex-1 overflow-y-auto min-h-0">
			{#if cart.items.length === 0}
				<div class="flex flex-col items-center justify-center h-full gap-3 text-slate-300 dark:text-slate-600 p-8">
					<ShoppingCart size={40} class="opacity-50" />
					<p class="text-sm text-center font-medium">Cart is empty</p>
					<p class="text-xs text-center text-slate-400">Click a product or scan a barcode to add items</p>
				</div>
			{:else}
				<ul class="divide-y divide-slate-50 dark:divide-slate-800">
					{#each cart.items as item (item.product.id)}
						<li class="px-4 py-3 hover:bg-slate-50/50 dark:hover:bg-slate-800/50 transition-colors">
							<div class="flex items-start gap-2.5">
								<!-- Thumbnail -->
								<div class="h-10 w-10 shrink-0 rounded-lg overflow-hidden bg-slate-100 dark:bg-slate-700 flex items-center justify-center">
									{#if item.product.image_url}
										<img src={item.product.image_url} alt={item.product.name} class="h-full w-full object-cover" />
									{:else}
										<Package size={14} class="text-slate-300" />
									{/if}
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-slate-800 dark:text-slate-100 leading-tight truncate">{item.product.name}</p>
									<p class="text-xs text-slate-400 mt-0.5">KES {fmt(item.unit_price)} each</p>
								</div>
								<button
									onclick={() => cart.removeItem(item.product.id)}
									class="shrink-0 h-6 w-6 flex items-center justify-center rounded-md text-slate-300 hover:text-red-500 hover:bg-red-50 transition-colors mt-0.5"
								>
									<X size={13} />
								</button>
							</div>
							<div class="flex items-center justify-between mt-2.5 pl-[52px]">
								<!-- Qty stepper -->
								<div class="flex items-center gap-1 rounded-xl border border-slate-200 dark:border-slate-600 overflow-hidden">
									<button
										onclick={() => setQuantity(item.product.id, item.quantity - 1)}
										class="h-8 w-8 flex items-center justify-center text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
									>
										<Minus size={12} />
									</button>
									<input
										type="number"
										value={item.quantity}
										onchange={(e) => setQuantity(item.product.id, Math.max(1, parseInt((e.target as HTMLInputElement).value) || 1))}
										class="w-10 text-center text-sm font-bold text-slate-800 dark:text-slate-100 border-x border-slate-200 dark:border-slate-600 bg-transparent py-0 focus:outline-none [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none"
									/>
									<button
										onclick={() => setQuantity(item.product.id, item.quantity + 1)}
										class="h-8 w-8 flex items-center justify-center text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
									>
										<Plus size={12} />
									</button>
								</div>
								<p class="text-sm font-bold text-slate-800 dark:text-slate-100 tabular-nums">KES {fmt(item.unit_price * item.quantity)}</p>
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		</div>

		<!-- Totals -->
		<div class="border-t border-slate-100 dark:border-slate-700 px-4 py-3 space-y-1.5 text-sm bg-slate-50 dark:bg-slate-900/60 shrink-0">
			<div class="flex justify-between text-slate-500 dark:text-slate-400">
				<span>Subtotal</span>
				<span class="tabular-nums">KES {fmt(cart.subtotal)}</span>
			</div>
			{#if cart.discount > 0}
				<div class="flex justify-between text-emerald-600">
					<span>Discount</span>
					<span class="tabular-nums">-KES {fmt(cart.discount)}</span>
				</div>
			{/if}
			{#if cart.taxRate > 0}
				<div class="flex justify-between text-slate-500 dark:text-slate-400">
					<span>Tax ({cart.taxRate}%)</span>
					<span class="tabular-nums">KES {fmt(cart.taxAmount)}</span>
				</div>
			{/if}
			{#if cart.note}
				<div class="flex items-center gap-1 text-xs text-slate-400">
					<FileText size={11} /><span class="truncate italic">{cart.note}</span>
				</div>
			{/if}
			<div class="flex justify-between font-bold text-base text-slate-900 dark:text-slate-100 pt-1.5 border-t border-slate-200 dark:border-slate-600">
				<span>Total</span>
				<span class="tabular-nums">KES {fmt(cart.total)}</span>
			</div>
		</div>

		<!-- Payment section -->
		<div id="payment-section" class="border-t border-slate-100 dark:border-slate-700 px-4 py-3 space-y-3 shrink-0 bg-white dark:bg-slate-900">

			<!-- Payment method tabs -->
			<div class="grid grid-cols-5 gap-1.5">
				{#each payments as p}
					<button
						onclick={() => paymentMethod = p.id}
						class="flex flex-col items-center gap-1 rounded-xl border py-2.5 text-[10px] font-semibold transition-all active:scale-95"
						style={paymentMethod === p.id ? 'background-color:#008B8B;border-color:#008B8B;color:white;' : 'border-color:#e2e8f0;color:#64748b;'}
					>
						<svelte:component this={p.icon} size={15} />
						{p.label}
					</button>
				{/each}
			</div>

			<!-- Cash amount input -->
			{#if paymentMethod === 'cash'}
				<div class="space-y-2">
					<div class="relative">
						<span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-sm font-semibold text-slate-400">KES</span>
						<input
							type="number"
							bind:value={amountTendered}
							min={0}
							step="50"
							placeholder="0.00"
							class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100 py-3 pl-14 pr-4 text-xl font-bold text-right tabular-nums focus:outline-none"
						/>
					</div>
					<!-- Quick amounts -->
					<div class="grid grid-cols-4 gap-1.5">
						{#each [500, 1000, 2000, 5000] as amt}
							<button
								onclick={() => amountTendered = amt}
								class="rounded-lg border border-slate-200 dark:border-slate-600 py-1.5 text-xs font-semibold text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
							>
								{amt.toLocaleString()}
							</button>
						{/each}
					</div>
					{#if amountTendered > 0}
						<div class="flex items-center justify-between rounded-xl px-4 py-2.5 {change > 0 ? 'bg-emerald-50 dark:bg-emerald-900/20' : 'bg-red-50 dark:bg-red-900/20'}">
							<span class="text-sm font-semibold {change > 0 ? 'text-emerald-700 dark:text-emerald-400' : 'text-red-600'}">
								{change > 0 ? 'Change' : 'Short'}
							</span>
							<span class="text-lg font-bold tabular-nums {change > 0 ? 'text-emerald-700 dark:text-emerald-400' : 'text-red-600'}">
								KES {fmt(Math.abs(change > 0 ? change : amountTendered - cart.total))}
							</span>
						</div>
					{/if}
				</div>
			{:else if paymentMethod === 'mpesa'}
				<div class="rounded-xl bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 p-3 text-center">
					<Smartphone size={20} class="mx-auto mb-1 text-emerald-600" />
					<p class="text-sm font-semibold text-emerald-700 dark:text-emerald-400">M-Pesa</p>
					<p class="text-xs text-emerald-600 dark:text-emerald-500 mt-0.5">Amount: KES {fmt(cart.total)}</p>
				</div>
			{/if}

			<!-- WhatsApp number -->
			<div class="relative">
				<MessageCircle size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
				<input
					type="tel"
					bind:value={whatsappNumber}
					placeholder="WhatsApp: 0792 397 476"
					class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100 dark:placeholder-slate-500 py-2.5 pl-9 pr-4 text-sm focus:outline-none"
				/>
			</div>

			<!-- Utility row -->
			<div class="grid grid-cols-3 gap-1.5">
				<button onclick={openDiscountModal} class="flex items-center justify-center gap-1 rounded-xl border border-slate-200 dark:border-slate-600 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
					<Tag size={12} /> Discount
				</button>
				<button onclick={() => { noteInput = cart.note; showNoteModal = true; }} class="flex items-center justify-center gap-1 rounded-xl border border-slate-200 dark:border-slate-600 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
					<FileText size={12} /> Notes
				</button>
				<button onclick={holdSale} disabled={cart.items.length === 0} class="flex items-center justify-center gap-1 rounded-xl border border-slate-200 dark:border-slate-600 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-40 transition-colors">
					<Clock size={12} /> Hold
				</button>
			</div>

			<!-- Checkout buttons -->
			<div class="grid grid-cols-5 gap-2">
				<button
					onclick={() => cart.clear()}
					disabled={cart.items.length === 0}
					class="col-span-1 rounded-xl border border-red-200 dark:border-red-800 py-3.5 text-xs font-semibold text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-40 transition-colors"
				>
					Clear
				</button>
				<button
					onclick={completeSale}
					disabled={cart.items.length === 0 || checkingOut}
					class="col-span-4 rounded-xl py-3.5 text-sm font-bold text-white disabled:opacity-50 transition-all active:scale-[0.98]"
					style="background-color:#008B8B; box-shadow:0 4px 14px rgba(0,139,139,0.3);"
				>
					{checkingOut ? 'Processing…' : `Charge KES ${fmt(cart.total)}`}
				</button>
			</div>

			<!-- Keyboard hints -->
			<div class="flex justify-center gap-4 text-[10px] text-slate-300 dark:text-slate-600">
				<span><kbd class="rounded border border-slate-200 dark:border-slate-700 px-1 font-mono">F4</kbd> Customer</span>
				<span><kbd class="rounded border border-slate-200 dark:border-slate-700 px-1 font-mono">F5</kbd> Hold</span>
				<span><kbd class="rounded border border-slate-200 dark:border-slate-700 px-1 font-mono">F8</kbd> Pay</span>
			</div>
		</div>
	</div>
</div>

<!-- ─── Customer modal ───────────────────────────────────────── -->
<Modal open={showCustomerModal} title="Select Customer" onclose={() => showCustomerModal = false}>
	{#snippet children()}
		<div class="space-y-3">
			<div class="relative">
				<Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
				<input
					bind:value={customerSearch}
					oninput={onCustomerSearch}
					placeholder="Search customers…"
					autofocus
					class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3 py-2.5 pl-9 text-sm focus:outline-none"
				/>
			</div>
			<ul class="max-h-60 overflow-y-auto divide-y divide-slate-50 dark:divide-slate-700 rounded-xl border border-slate-200 dark:border-slate-600">
				{#each customers as c}
					<li>
						<button
							onclick={() => selectCustomer(c)}
							class="w-full px-4 py-3 text-left hover:bg-slate-50 dark:hover:bg-slate-700 text-sm flex items-center gap-3 transition-colors"
						>
							<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full font-bold text-xs text-white" style="background-color:#008B8B;">
								{c.name.charAt(0).toUpperCase()}
							</div>
							<div class="min-w-0 flex-1">
								<p class="font-semibold text-slate-900 dark:text-slate-100 truncate">{c.name}</p>
								{#if c.phone}<p class="text-slate-400 text-xs">{c.phone}</p>{/if}
							</div>
							<Check size={14} class="{cart.customerId === c.id ? 'text-teal-600' : 'opacity-0'}" />
						</button>
					</li>
				{:else}
					<li class="px-4 py-8 text-center text-sm text-slate-400">No customers found</li>
				{/each}
			</ul>
			<button
				onclick={() => { showCustomerModal = false; showNewCustomerModal = true; }}
				class="w-full flex items-center justify-center gap-2 rounded-xl border border-dashed border-slate-300 dark:border-slate-600 py-2.5 text-sm font-medium text-slate-500 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
			>
				<PlusCircle size={15} /> Add New Customer
			</button>
		</div>
	{/snippet}
</Modal>

<!-- ─── New customer modal ───────────────────────────────────── -->
<Modal open={showNewCustomerModal} title="New Customer" onclose={() => showNewCustomerModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-3.5">
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Name *</label>
				<input type="text" bind:value={newCustomerName} placeholder="Full name" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Phone</label>
				<input type="tel" bind:value={newCustomerPhone} placeholder="+254…" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Email</label>
				<input type="email" bind:value={newCustomerEmail} placeholder="customer@email.com" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showNewCustomerModal = false} class="rounded-xl border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
		<button onclick={createNewCustomer} class="rounded-xl px-5 py-2.5 text-sm font-semibold text-white transition-all active:scale-95" style="background-color:#008B8B;">Save</button>
	{/snippet}
</Modal>

<!-- ─── Held sales modal ─────────────────────────────────────── -->
<Modal open={showHeldModal} title="Held Sales" onclose={() => showHeldModal = false}>
	{#snippet children()}
		<ul class="divide-y divide-slate-100 dark:divide-slate-700 max-h-80 overflow-y-auto">
			{#each heldSales as sale}
				<li class="flex items-center gap-3 py-3">
					<div class="flex-1 min-w-0">
						<p class="text-sm font-semibold text-slate-800 dark:text-slate-100">{new Date(sale.created_at).toLocaleString()}</p>
						<p class="text-xs text-slate-400">{sale.items?.length ?? 0} items · KES {fmt(sale.total)}</p>
					</div>
					<button onclick={() => resumeSale(sale)} class="rounded-xl px-3 py-1.5 text-xs font-semibold text-white" style="background-color:#008B8B;">Resume</button>
				</li>
			{:else}
				<li class="py-12 text-center text-sm text-slate-400">No held sales</li>
			{/each}
		</ul>
	{/snippet}
</Modal>

<!-- ─── Discount & Tax modal ─────────────────────────────────── -->
<Modal open={showDiscountModal} title="Discount & Tax" onclose={() => showDiscountModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-4">
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Discount Amount (KES)</label>
				<input type="number" bind:value={discountInput} min="0" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Tax Rate (%)</label>
				<input type="number" bind:value={taxRateInput} min="0" max="100" step="0.5" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showDiscountModal = false} class="rounded-xl border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">Cancel</button>
		<button onclick={applyDiscount} class="rounded-xl px-5 py-2.5 text-sm font-semibold text-white" style="background-color:#008B8B;">Apply</button>
	{/snippet}
</Modal>

<!-- ─── Note modal ───────────────────────────────────────────── -->
<Modal open={showNoteModal} title="Sale Notes" onclose={() => showNoteModal = false} size="sm">
	{#snippet children()}
		<textarea bind:value={noteInput} rows={3} placeholder="Add a note…" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none resize-none"></textarea>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showNoteModal = false} class="rounded-xl border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">Cancel</button>
		<button onclick={applyNote} class="rounded-xl px-5 py-2.5 text-sm font-semibold text-white" style="background-color:#008B8B;">Save</button>
	{/snippet}
</Modal>

<!-- ─── Open Shift modal ─────────────────────────────────────── -->
<Modal open={showShiftModal} title="Open Shift" onclose={() => showShiftModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-4">
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Opening Float (KES)</label>
				<input type="number" bind:value={openingFloat} min="0" step="50" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" placeholder="e.g. 5000" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Notes (optional)</label>
				<textarea bind:value={shiftNotes} rows={2} class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none resize-none" placeholder="Opening notes…"></textarea>
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showShiftModal = false} class="rounded-xl border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">Cancel</button>
		<button onclick={openShift} class="rounded-xl px-5 py-2.5 text-sm font-semibold text-white" style="background-color:#008B8B;">Open Shift</button>
	{/snippet}
</Modal>

<!-- ─── Close Shift modal ────────────────────────────────────── -->
<Modal open={showCloseShiftModal} title="Close Shift" onclose={() => showCloseShiftModal = false} size="sm">
	{#snippet children()}
		{#if shiftStore.current}
			<div class="space-y-4">
				<div class="rounded-xl bg-slate-50 dark:bg-slate-800 p-4 space-y-2 text-sm">
					<div class="flex justify-between"><span class="text-slate-500">Opening Float</span><span class="font-semibold">KES {fmt(shiftStore.current.opening_float)}</span></div>
					<div class="flex justify-between"><span class="text-slate-500">Cash Sales</span><span class="font-semibold text-emerald-600">+ KES {fmt(shiftStore.current.cash_sales)}</span></div>
					<div class="flex justify-between border-t border-slate-200 dark:border-slate-600 pt-2"><span class="text-slate-700 dark:text-slate-300 font-semibold">Expected Cash</span><span class="font-bold">KES {fmt(shiftStore.current.opening_float + shiftStore.current.cash_sales)}</span></div>
				</div>
				<div>
					<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Actual Cash Counted (KES)</label>
					<input type="number" bind:value={closingCash} min="0" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" placeholder="Count the cash drawer" />
				</div>
				{#if closingCash > 0}
					{@const variance = closingCash - (shiftStore.current.opening_float + shiftStore.current.cash_sales)}
					<div class="rounded-xl p-3 text-center {variance >= 0 ? 'bg-emerald-50 dark:bg-emerald-900/20' : 'bg-red-50 dark:bg-red-900/20'}">
						<p class="text-xs font-semibold {variance >= 0 ? 'text-emerald-700 dark:text-emerald-400' : 'text-red-600'}">{variance >= 0 ? 'Surplus' : 'Shortage'}</p>
						<p class="text-xl font-bold tabular-nums {variance >= 0 ? 'text-emerald-700 dark:text-emerald-400' : 'text-red-600'}">KES {fmt(Math.abs(variance))}</p>
					</div>
				{/if}
			</div>
		{/if}
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showCloseShiftModal = false} class="rounded-xl border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">Cancel</button>
		<button onclick={closeShift} class="rounded-xl px-5 py-2.5 text-sm font-semibold text-white bg-red-600 hover:bg-red-700 transition-colors">Close Shift</button>
	{/snippet}
</Modal>

<!-- ─── Receipt overlay ──────────────────────────────────────── -->
{#if lastSale}
	<div class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 backdrop-blur-sm p-0 sm:p-4">
		<div class="relative w-full max-w-sm max-h-[92vh] flex flex-col rounded-t-3xl sm:rounded-2xl bg-white dark:bg-slate-800 shadow-2xl overflow-hidden">
			<!-- Header -->
			<div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-700 px-5 py-4 shrink-0">
				<div class="flex items-center gap-2">
					<div class="h-8 w-8 rounded-full flex items-center justify-center" style="background-color:#008B8B20;">
						<Check size={16} style="color:#008B8B;" />
					</div>
					<h2 class="text-base font-bold text-slate-900 dark:text-slate-100">Sale Complete!</h2>
				</div>
				<p class="text-sm font-bold tabular-nums" style="color:#008B8B;">KES {fmt(lastSale.total)}</p>
			</div>

			<!-- Receipt body -->
			<div class="px-5 py-4 overflow-y-auto flex-1 receipt-print">
				<div class="text-center mb-4">
					<h3 class="text-lg font-bold text-slate-900 dark:text-slate-100">POS System</h3>
					<p class="text-xs text-slate-400">#{lastSale.id.slice(0, 8).toUpperCase()}</p>
				</div>
				<div class="text-xs text-slate-500 space-y-1 mb-4">
					<div class="flex justify-between"><span>Date</span><span>{new Date(lastSale.created_at).toLocaleString()}</span></div>
					<div class="flex justify-between"><span>Cashier</span><span>{lastSale.cashier_name || '—'}</span></div>
					{#if lastSale.customer_name}<div class="flex justify-between"><span>Customer</span><span>{lastSale.customer_name}</span></div>{/if}
				</div>

				<table class="w-full text-xs mb-4 border-t border-b border-dashed border-slate-200 dark:border-slate-600 py-2">
					<thead><tr class="text-slate-400"><th class="py-1.5 text-left">Item</th><th class="py-1.5 text-center">Qty</th><th class="py-1.5 text-right">Total</th></tr></thead>
					<tbody>
						{#each lastSale.items || [] as item}
							<tr class="border-t border-dashed border-slate-100 dark:border-slate-700">
								<td class="py-1.5 pr-2 text-slate-700 dark:text-slate-300">{item.product_name ?? '—'}</td>
								<td class="py-1.5 text-center text-slate-500">{item.quantity}</td>
								<td class="py-1.5 text-right font-semibold text-slate-800 dark:text-slate-100 tabular-nums">KES {fmt(item.total)}</td>
							</tr>
						{/each}
					</tbody>
				</table>

				<div class="text-sm space-y-1">
					{#if lastSale.discount > 0}
						<div class="flex justify-between text-emerald-600"><span>Discount</span><span>-KES {fmt(lastSale.discount)}</span></div>
					{/if}
					<div class="flex justify-between font-bold text-slate-900 dark:text-slate-100 border-t border-slate-200 dark:border-slate-600 pt-1">
						<span>Total</span><span class="tabular-nums">KES {fmt(lastSale.total)}</span>
					</div>
					<div class="flex justify-between text-slate-400 text-xs pt-1">
						<span>Paid via</span><span class="capitalize">{lastSale.payment_method}</span>
					</div>
					{#if lastSale.payment_method === 'cash'}
						<div class="flex justify-between text-slate-400 text-xs">
							<span>Amount Tendered</span><span class="tabular-nums">KES {fmt(lastAmountTendered)}</span>
						</div>
						<div class="flex justify-between font-bold text-emerald-600 text-xs">
							<span>Change</span><span class="tabular-nums">KES {fmt(lastChange)}</span>
						</div>
					{/if}
				</div>
				<p class="text-center text-xs text-slate-400 mt-4 pt-4 border-t border-slate-100 dark:border-slate-700">Thank you for your purchase! 🙏</p>
			</div>

			<!-- Actions -->
			<div class="flex items-center gap-2 border-t border-slate-100 dark:border-slate-700 px-5 py-4 shrink-0">
				<button onclick={printReceipt} class="flex items-center gap-1.5 rounded-xl border border-slate-200 dark:border-slate-600 px-3 py-2.5 text-xs font-semibold text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
					<Printer size={14} /> Print
				</button>
				<button onclick={sendWhatsApp} class="flex items-center gap-1.5 rounded-xl border border-emerald-200 dark:border-emerald-800 bg-emerald-50 dark:bg-emerald-900/20 px-3 py-2.5 text-xs font-semibold text-emerald-700 dark:text-emerald-400 hover:bg-emerald-100 transition-colors">
					<MessageCircle size={14} /> WhatsApp
				</button>
				<button onclick={resetAfterSale} class="flex-1 rounded-xl py-2.5 text-sm font-bold text-white transition-all active:scale-95" style="background-color:#008B8B;">
					New Sale
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.scrollbar-none { scrollbar-width: none; }
	.scrollbar-none::-webkit-scrollbar { display: none; }

	@media print {
		:global(body *) { visibility: hidden; }
		:global(.receipt-print), :global(.receipt-print *) { visibility: visible; }
		:global(.receipt-print) { position: absolute; left: 0; top: 0; width: 80mm; }
	}
</style>
