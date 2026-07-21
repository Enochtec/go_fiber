<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { cart } from '$lib/stores/cart.svelte';
	import { notify } from '$lib/stores/notification.svelte';
	import { invalidation } from '$lib/stores/invalidation.svelte';
	import { offlineProducts as productsService, offlineCustomers as customersService, offlineSales as salesService } from '$lib/services/offline';
	import { mpesaService, pollMpesaStatus, validatePhone, normalizePhone } from '$lib/services/mpesa';
	import Modal from '$lib/components/Modal.svelte';
	import BarcodeScanner from '$lib/components/BarcodeScanner.svelte';
	import Calculator from '$lib/components/Calculator.svelte';
	import Receipt from '$lib/components/Receipt.svelte';
	import type { Product, Customer, Sale } from '$lib/types';
	import {
		Search, ShoppingCart, Plus, Minus, Trash2, User,
		Clock, Printer, X, Banknote, Smartphone, Package,
		ReceiptIcon, ChevronDown, PlusCircle,
		Percent, FileText, Tag, AlertTriangle, Check,
		ArrowLeft, Hash, Scan, Phone, Loader, CheckCircle, XCircle, RefreshCw,
		Calculator as CalculatorIcon
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

	let paymentMethod = $state<'cash' | 'mpesa'>('cash');
	let amountTendered = $state(0);
	let checkingOut = $state(false);
	let productsLoading = $state(false);
	let cartOpen = $state(false);

	let showCustomerModal = $state(false);
	let showHeldModal = $state(false);
	let showDiscountModal = $state(false);
	let showNoteModal = $state(false);
	let showNewCustomerModal = $state(false);
	let showScanner = $state(false);
	let showCalculator = $state(false);

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
	let clockStr = $state('');

	// M-Pesa state
	let mpesaPhone = $state('');
	let mpesaPhoneError = $state('');
	let mpesaStatus = $state<'idle' | 'initiating' | 'waiting' | 'completed' | 'failed' | 'cancelled'>('idle');
	let mpesaMessage = $state('');
	let mpesaCheckoutID = $state('');

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

	async function createSaleRecord(method: 'cash' | 'mpesa') {
		const res = await salesService.create({
			customer_id: cart.customerId,
			items: cart.items.map(i => ({ product_id: i.product.id, quantity: i.quantity, unit_price: i.unit_price })),
			discount: cart.discount, tax_rate: cart.taxRate,
			payment_method: method, status: 'completed', note: cart.note || null,
		});
		if (res.data) {
			lastSale = res.data;
			lastAmountTendered = method === 'cash' ? amountTendered : cart.total;
			lastChange = method === 'cash' ? change : 0;
		}
	}

	async function completeSale() {
		if (cart.items.length === 0) { notify.error('Cart is empty'); return; }

		// M-Pesa flow
		if (paymentMethod === 'mpesa') {
			const phoneErr = validatePhone(mpesaPhone);
			if (phoneErr) { mpesaPhoneError = phoneErr; return; }
			mpesaPhoneError = '';
			mpesaStatus = 'initiating';
			mpesaMessage = 'Sending payment request…';
			checkingOut = true;
			try {
				const stkRes = await mpesaService.initiateSTKPush({
					phone:     mpesaPhone,
					amount:    cart.total,
					reference: cart.customerId ? 'POS-CUST' : 'POS-SALE'
				});
				mpesaCheckoutID = stkRes.data!.checkout_request_id;
				mpesaStatus    = 'waiting';
				mpesaMessage   = stkRes.data!.customer_message || 'Check your phone and enter M-Pesa PIN…';

				const result = await pollMpesaStatus(
					mpesaCheckoutID,
					(s) => {
						if (s.status === 'pending') mpesaMessage = 'Waiting for M-Pesa confirmation…';
					}
				);

				if (result.status === 'completed') {
					mpesaStatus = 'completed';
					await createSaleRecord('mpesa');
				} else if (result.status === 'cancelled') {
					mpesaStatus  = 'cancelled';
					mpesaMessage = 'Payment was cancelled. Please try again.';
				} else {
					mpesaStatus  = 'failed';
					mpesaMessage = result.result_desc || 'Payment failed. Please try again.';
				}
			} catch (err) {
				mpesaStatus  = 'failed';
				mpesaMessage = err instanceof Error ? err.message : 'M-Pesa request failed';
			} finally {
				checkingOut = false;
			}
			return;
		}

		// Cash / other payment
		if (paymentMethod === 'cash' && amountTendered < cart.total) {
			notify.error('Amount tendered is less than total');
			return;
		}
		checkingOut = true;
		try {
			await createSaleRecord('cash');
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
		mpesaPhone = '';
		mpesaPhoneError = '';
		mpesaStatus = 'idle';
		mpesaMessage = '';
		mpesaCheckoutID = '';
		lastSale = null;
		loadProducts();
		searchInput?.focus();
	}

	// ─── Keyboard ───────────────────────────────────────────────────
	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') { showCalculator = false; }
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

	$effect(() => {
		invalidation.productVersion;
		if (document.visibilityState === 'visible') loadProducts();
	});

	onMount(async () => {
		updateClock();
		clockTimer = setInterval(updateClock, 1000);
		const catRes = await productsService.listCategories();
		categories = catRes.data ?? [];
		await loadProducts();
		searchInput?.focus();
	});
</script>

<svelte:head><title>Checkout — POS</title></svelte:head>
<svelte:window onkeydown={handleKeydown} />

	<div class="flex h-full overflow-hidden bg-slate-100">


	<!-- Mobile cart overlay -->
	{#if cartOpen}
		<div class="fixed inset-0 z-20 bg-black/60 lg:hidden" role="presentation" onclick={() => cartOpen = false}></div>
	{/if}

	<!-- ── LEFT: Product Browser ─────────────────────────────────── -->
	<div class="flex flex-1 flex-col overflow-hidden min-w-0">

		<!-- Unified header -->
		<div class="sticky top-0 z-30 shrink-0 bg-white shadow-sm">
			<div class="flex items-center gap-2 px-3 py-2">

				<!-- Digital clock -->
				<div class="hidden md:flex items-center gap-1.5 text-sm text-slate-500 font-mono shrink-0" aria-live="polite">
					<span id="clock-display">{clockStr}</span>
				</div>

				<!-- Search -->
				<div class="relative flex-1 min-w-0">
					<Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
					<input
						bind:this={searchInput}
						bind:value={search}
						oninput={onSearch}
						onkeydown={handleSearchKeydown}
						onfocus={() => { if (search) showSearchDropdown = true; }}
						placeholder="Search product… (F2)"
						class="w-full rounded border border-slate-200 bg-slate-50 py-2.5 pl-9 pr-3.5 text-base sm:text-sm focus:outline-none focus:border-slate-300 focus:bg-white transition-colors"
					/>
					{#if search}
						<button onclick={closeSearch} class="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600">
							<X size={13} />
						</button>
					{/if}
				</div>

				<!-- Toolbar group: Calculator | Scan | Held Sales -->
				<div class="flex items-center bg-slate-100 p-0.5 gap-px shrink-0">
					<button
						onclick={() => showCalculator = true}
						class="flex h-8 w-8 items-center justify-center bg-purple-600 text-white hover:bg-purple-500 transition-colors"
						title="Calculator"
					>
						<CalculatorIcon size={15} />
					</button>
					<button
						onclick={() => showScanner = true}
						class="flex h-8 w-8 items-center justify-center bg-blue-600 text-white hover:bg-blue-500 transition-colors"
						title="Scan Barcode"
					>
						<Scan size={15} />
					</button>
					<button
						onclick={async () => { await fetchHeld(); showHeldModal = true; }}
						class="relative flex h-8 w-8 items-center justify-center bg-amber-600 text-white hover:bg-amber-500 transition-colors"
						title="Held Sales"
					>
						<Clock size={14} />
						{#if heldSales.length > 0}
							<span class="absolute -top-1 -right-1 h-4 w-4 flex items-center justify-center bg-red-500 text-[9px] font-bold text-white">{heldSales.length}</span>
						{/if}
					</button>
				</div>
			</div>
			<!-- Mobile clock row -->
			<div class="md:hidden flex items-center justify-center px-3 pb-1.5">
				<span class="text-xs sm:text-sm text-slate-400 font-mono">{clockStr}</span>
			</div>
		</div>

		<!-- Search dropdown -->
		{#if showSearchDropdown}
			<div
				bind:this={searchResultsEl}
				class="absolute top-[52px] left-0 right-0 lg:right-[400px] z-20 mx-3 max-h-72 overflow-y-auto rounded border border-slate-200 bg-white shadow-xl"
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
							<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded bg-slate-100">
								{#if product.image_url}
									<img src={product.image_url} alt={product.name} class="h-full w-full object-cover rounded" />
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
								<p class="text-sm font-bold text-blue-600">KES {fmt(product.selling_price)}</p>
								{#if product.barcode}<p class="text-xs text-slate-400 font-mono">{product.barcode}</p>{/if}
							</div>
						</button>
					{/each}
				{/if}
			</div>
		{/if}

		<!-- Category pills -->
		<div class="flex gap-2 px-3 py-2 overflow-x-auto shrink-0 bg-white dark:bg-slate-900 scrollbar-none">
			<button
				onclick={() => { selectedCategory = ''; loadProducts(); }}
				class="inline-flex shrink-0 items-center rounded-full px-3.5 py-1.5 text-xs font-semibold transition-all
					{selectedCategory === '' ? 'bg-blue-600 text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600'}"
			>
				All
			</button>
			{#each categories as cat}
				<button
					onclick={() => { selectedCategory = cat.id; loadProducts(); }}
					class="inline-flex shrink-0 items-center rounded-full px-3.5 py-1.5 text-xs font-semibold transition-all
						{selectedCategory === cat.id ? 'bg-blue-600 text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600'}"
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
						<div class="rounded-[1px] bg-white dark:bg-slate-800 h-36 animate-pulse border border-slate-100 dark:border-slate-700"></div>
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
							class="group relative rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-3 text-left transition-all active:scale-[0.97] disabled:opacity-50 disabled:cursor-not-allowed shadow-sm hover:shadow-md hover:border-blue-200 dark:hover:border-blue-800 hover:-translate-y-0.5 flex flex-col"
						>
							<!-- Product image -->
							<div class="relative w-full aspect-square rounded overflow-hidden bg-slate-100 dark:bg-slate-700 mb-2.5 flex items-center justify-center">
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
								<p class="text-sm font-bold text-blue-600">KES {fmt(product.selling_price)}</p>
								<span class="text-[10px] text-slate-400">{product.stock_qty} left</span>
							</div>

							<!-- Add overlay on hover -->
							<div class="absolute inset-0 rounded-[1px] flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none bg-blue-600/5">
								<div class="rounded-full bg-blue-600 text-white p-1.5 shadow-lg">
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
		class="fixed bottom-24 left-1/2 -translate-x-1/2 z-10 flex items-center gap-2.5 rounded-full px-5 py-3 text-white bg-blue-600 shadow-xl shadow-blue-500/30 lg:hidden active:scale-95 transition-all"
		class:opacity-0={cart.items.length === 0}
		class:pointer-events-none={cart.items.length === 0}
	>
		<ShoppingCart size={18} />
		<span class="text-sm font-semibold">{cart.count} item{cart.count !== 1 ? 's' : ''} · KES {fmt(cart.total)}</span>
	</button>

	<!-- ── RIGHT: Cart + Payment Panel ───────────────────────────── -->
	<div
		class="flex flex-col bg-white dark:bg-slate-900 transition-transform duration-300
			fixed inset-x-0 bottom-0 z-30 max-h-[92vh] rounded-t-3xl shadow-2xl
			lg:relative lg:inset-auto lg:z-auto lg:max-h-full lg:w-[400px] lg:shrink-0 lg:rounded-none lg:shadow-none lg:translate-y-0
			{cartOpen ? 'translate-y-0' : 'translate-y-full'}"
	>
		<!-- Mobile drag handle -->
		<div class="flex justify-center pt-3 pb-1 lg:hidden shrink-0">
			<div class="h-1 w-10 rounded-full bg-slate-200 dark:bg-slate-600"></div>
		</div>

		<!-- Cart header -->
		<div class="flex items-center justify-between px-4 py-2.5 shrink-0">
			<div class="flex items-center gap-2">
				<ShoppingCart size={16} class="text-slate-500" />
				<span class="font-semibold text-slate-800 dark:text-slate-100 text-sm">Order</span>
				{#if cart.count > 0}
					<span class="inline-flex h-5 min-w-5 items-center justify-center rounded-full text-[10px] font-bold text-white px-1 bg-blue-600">{cart.count}</span>
				{/if}
			</div>
			<div class="flex items-center gap-1">
				<button onclick={openCustomerModal} class="flex items-center gap-1.5 rounded px-2.5 py-1.5 text-xs font-medium transition-colors border border-slate-200 dark:border-slate-600 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">
					<User size={12} />
					{cart.customerName ?? 'Walk-in'}
					<ChevronDown size={10} />
				</button>
				{#if cart.customerName}
					<button onclick={removeCustomer} class="h-7 w-7 flex items-center justify-center rounded text-slate-400 hover:text-red-500 hover:bg-red-50 transition-colors">
						<X size={13} />
					</button>
				{/if}
				<button onclick={() => cartOpen = false} class="lg:hidden h-7 w-7 flex items-center justify-center rounded text-slate-400 hover:bg-slate-100">
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
								<div class="h-10 w-10 shrink-0 rounded overflow-hidden bg-slate-100 dark:bg-slate-700 flex items-center justify-center">
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
								<div class="flex items-center gap-1 rounded border border-slate-200 dark:border-slate-600 overflow-hidden">
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
		<div class="px-4 py-3 space-y-1.5 text-sm bg-slate-50 dark:bg-slate-900/60 shrink-0">
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
			<div class="flex justify-between font-bold text-base text-slate-900 dark:text-slate-100 pt-1.5">
				<span>Total</span>
				<span class="tabular-nums">KES {fmt(cart.total)}</span>
			</div>
		</div>

		<!-- Payment section -->
		<div class="px-4 py-3 space-y-3 shrink-0 bg-white dark:bg-slate-900">

			<!-- Payment method tabs -->
			<div class="grid grid-cols-2 gap-1.5">
				{#each payments as p}
					<button
						onclick={() => paymentMethod = p.id}
						class="flex flex-col items-center gap-1 rounded border py-2.5 text-[10px] font-semibold transition-all active:scale-95
							{paymentMethod === p.id
								? 'bg-blue-600 border-blue-600 text-white shadow-sm'
								: 'border-slate-200 dark:border-slate-600 text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700 hover:border-slate-300'}"
					>
						<p.icon size={15} />
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
							class="w-full rounded border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100 py-3 pl-14 pr-4 text-xl font-bold text-right tabular-nums focus:outline-none"
						/>
					</div>
					<!-- Quick amounts -->
					<div class="grid grid-cols-4 gap-1.5">
						{#each [500, 1000, 2000, 5000] as amt}
							<button
								onclick={() => amountTendered = amt}
								class="rounded border border-slate-200 dark:border-slate-600 py-1.5 text-xs font-semibold text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
							>
								{amt.toLocaleString()}
							</button>
						{/each}
					</div>
					{#if amountTendered > 0}
						<div class="flex items-center justify-between rounded px-4 py-2.5 {change > 0 ? 'bg-emerald-50 dark:bg-emerald-900/20' : 'bg-red-50 dark:bg-red-900/20'}">
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
				<div class="space-y-2">
					<!-- Phone input (only when idle/failed/cancelled) -->
					{#if mpesaStatus === 'idle' || mpesaStatus === 'failed' || mpesaStatus === 'cancelled'}
						<div>
							<label class="text-[10px] font-semibold text-slate-500 uppercase tracking-wide block mb-1">Customer Phone (M-Pesa)</label>
							<div class="relative">
								<Smartphone size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-emerald-500" />
								<input
									type="tel"
									bind:value={mpesaPhone}
									placeholder="e.g. 0712 345 678"
									class="w-full border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100 py-2.5 pl-9 pr-4 text-sm focus:outline-none focus:border-emerald-400"
								/>
							</div>
							{#if mpesaPhoneError}<p class="text-xs text-red-500 mt-1">{mpesaPhoneError}</p>{/if}
						</div>
						{#if mpesaStatus === 'failed' || mpesaStatus === 'cancelled'}
							<div class="flex items-start gap-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-3 py-2">
								<XCircle size={14} class="text-red-500 mt-0.5 shrink-0" />
								<p class="text-xs text-red-700 dark:text-red-400">{mpesaMessage}</p>
							</div>
						{/if}
						<div class="bg-emerald-50 dark:bg-emerald-900/20 px-3 py-2 text-xs text-emerald-700 dark:text-emerald-400 text-center">
							<span class="font-semibold">Amount: KES {fmt(cart.total)}</span> — A push notification will be sent to the customer's phone.
						</div>
					{:else if mpesaStatus === 'initiating' || mpesaStatus === 'waiting'}
						<div class="bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 px-4 py-5 text-center space-y-2">
							<Loader size={24} class="mx-auto animate-spin text-emerald-600" />
							<p class="text-sm font-semibold text-emerald-700 dark:text-emerald-400">
								{mpesaStatus === 'initiating' ? 'Initiating STK Push…' : 'Waiting for Payment…'}
							</p>
							<p class="text-xs text-emerald-600 dark:text-emerald-500">{mpesaMessage}</p>
							<p class="text-xs text-slate-500">Amount: KES {fmt(cart.total)} → {normalizePhone(mpesaPhone)}</p>
						</div>
					{:else if mpesaStatus === 'completed'}
						<div class="bg-emerald-50 border border-emerald-200 px-4 py-3 flex items-center gap-3">
							<CheckCircle size={20} class="text-emerald-600 shrink-0" />
							<div>
								<p class="text-sm font-semibold text-emerald-800">Payment Confirmed!</p>
								<p class="text-xs text-emerald-600">KES {fmt(cart.total)} received via M-Pesa</p>
							</div>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Utility row -->
			<div class="grid grid-cols-3 gap-1.5">
				<button onclick={openDiscountModal} class="flex items-center justify-center gap-1 rounded border border-slate-200 dark:border-slate-600 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
					<Tag size={12} /> Discount
				</button>
				<button onclick={() => { noteInput = cart.note; showNoteModal = true; }} class="flex items-center justify-center gap-1 rounded border border-slate-200 dark:border-slate-600 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
					<FileText size={12} /> Notes
				</button>
				<button onclick={holdSale} disabled={cart.items.length === 0} class="flex items-center justify-center gap-1 rounded border border-slate-200 dark:border-slate-600 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-40 transition-colors">
					<Clock size={12} /> Hold
				</button>
			</div>

			<!-- Checkout buttons -->
			<div class="space-y-2">
				<button
					onclick={completeSale}
					disabled={cart.items.length === 0 || checkingOut}
					class="w-full rounded py-4 text-sm font-bold text-white bg-blue-600 hover:bg-blue-700 shadow-lg shadow-blue-200 dark:shadow-blue-900/40 disabled:opacity-50 transition-all active:scale-[0.98]"
				>
					{#if checkingOut}
						<span class="flex items-center justify-center gap-2">
							<span class="h-4 w-4 rounded-full border-2 border-white/40 border-t-white animate-spin"></span>
							{paymentMethod === 'mpesa' ? 'Waiting for M-Pesa…' : 'Processing…'}
						</span>
					{:else if paymentMethod === 'mpesa'}
						Send M-Pesa Request — KES {fmt(cart.total)}
					{:else}
						Charge KES {fmt(cart.total)}
					{/if}
				</button>
				<button
					onclick={() => cart.clear()}
					disabled={cart.items.length === 0}
					class="w-full rounded border border-red-200 dark:border-red-800 py-2.5 text-xs font-semibold text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-40 transition-colors"
				>
					Clear Cart
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
					class="w-full rounded border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3 py-2.5 pl-9 text-sm focus:outline-none"
				/>
			</div>
			<ul class="max-h-60 overflow-y-auto divide-y divide-slate-50 dark:divide-slate-700 rounded border border-slate-200 dark:border-slate-600">
				{#each customers as c}
					<li>
						<button
							onclick={() => selectCustomer(c)}
							class="w-full px-4 py-3 text-left hover:bg-slate-50 dark:hover:bg-slate-700 text-sm flex items-center gap-3 transition-colors"
						>
							<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full font-bold text-xs text-white bg-blue-600">
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
				class="w-full flex items-center justify-center gap-2 rounded border border-dashed border-slate-300 dark:border-slate-600 py-2.5 text-sm font-medium text-slate-500 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
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
				<input type="text" bind:value={newCustomerName} placeholder="Full name" class="w-full rounded border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Phone</label>
				<input type="tel" bind:value={newCustomerPhone} placeholder="+254…" class="w-full rounded border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Email</label>
				<input type="email" bind:value={newCustomerEmail} placeholder="customer@email.com" class="w-full rounded border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showNewCustomerModal = false} class="rounded border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
		<button onclick={createNewCustomer} class="rounded px-5 py-2.5 text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 transition-all active:scale-95">Save</button>
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
					<button onclick={() => resumeSale(sale)} class="rounded px-3 py-1.5 text-xs font-semibold text-white bg-blue-600 hover:bg-blue-700 transition-colors">Resume</button>
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
				<input type="number" bind:value={discountInput} min="0" class="w-full rounded border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Tax Rate (%)</label>
				<input type="number" bind:value={taxRateInput} min="0" max="100" step="0.5" class="w-full rounded border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showDiscountModal = false} class="rounded border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">Cancel</button>
		<button onclick={applyDiscount} class="rounded px-5 py-2.5 text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 transition-colors">Apply</button>
	{/snippet}
</Modal>

<!-- ─── Note modal ───────────────────────────────────────────── -->
<Modal open={showNoteModal} title="Sale Notes" onclose={() => showNoteModal = false} size="sm">
	{#snippet children()}
		<textarea bind:value={noteInput} rows={3} placeholder="Add a note…" class="w-full rounded border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none resize-none"></textarea>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showNoteModal = false} class="rounded border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700">Cancel</button>
		<button onclick={applyNote} class="rounded px-5 py-2.5 text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 transition-colors">Save</button>
	{/snippet}
</Modal>

<!-- ─── Receipt overlay ──────────────────────────────────────── -->
{#if lastSale}
	<Receipt
		sale={lastSale}
		amountTendered={lastAmountTendered}
		change={lastChange}
		onclose={() => lastSale = null}
		onnewsale={resetAfterSale}
	/>
{/if}

<!-- Barcode Scanner -->
{#if showScanner}
	<BarcodeScanner
		onscan={(code) => {
			showScanner = false;
			search = code;
			searchBarcodeOrSKU(code);
		}}
		onclose={() => showScanner = false}
	/>
{/if}

<!-- Calculator -->
{#if showCalculator}
	<Calculator onclose={() => showCalculator = false} />
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
