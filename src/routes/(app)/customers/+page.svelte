<script lang="ts">
	import { onMount } from 'svelte';
	import { customersService, type CustomerInput, type CustomerStats } from '$lib/services/customers';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import type { Customer, Sale } from '$lib/types';
	import {
		Plus, Search, Pencil, Trash2, X, User, Phone, Mail,
		MapPin, ShoppingBag, TrendingUp, Clock, ChevronRight, Download
	} from '@lucide/svelte';
	import ExportModal from '$lib/components/ExportModal.svelte';
	import { shopService } from '$lib/services/shop';
	import { authStore } from '$lib/stores/auth.svelte';
	import { exportCustomers, downloadCSV, safeFilename } from '$lib/services/export';

	// ─── List state ──────────────────────────────────────────────
	let customers = $state<Customer[]>([]);
	let total = $state(0);
	let page = $state(1);
	const limit = 20;
	let search = $state('');
	let loading = $state(true);
	let showExport = $state(false);

	// ─── Form state ───────────────────────────────────────────────
	let showModal = $state(false);
	let editing = $state<Customer | null>(null);
	let submitting = $state(false);
	let form = $state<CustomerInput>({ name: '', email: null, phone: null, address: null });

	// ─── Detail panel ────────────────────────────────────────────
	let selected = $state<Customer | null>(null);
	let customerStats = $state<CustomerStats | null>(null);
	let purchaseHistory = $state<Sale[]>([]);
	let detailLoading = $state(false);

	// ─── Data fetching ───────────────────────────────────────────
	async function fetchList() {
		loading = true;
		try {
			const res = await customersService.list(search, page, limit);
			customers = res.data ?? [];
			total = res.total ?? 0;
		} finally {
			loading = false;
		}
	}

	let debounce: ReturnType<typeof setTimeout>;
	function onSearch() { page = 1; clearTimeout(debounce); debounce = setTimeout(fetchList, 300); }

	async function selectCustomer(c: Customer) {
		selected = c;
		detailLoading = true;
		customerStats = null;
		purchaseHistory = [];
		try {
			const [statsRes, histRes] = await Promise.all([
				customersService.stats(c.id),
				customersService.history(c.id, 10)
			]);
			customerStats = statsRes.data ?? null;
			purchaseHistory = histRes.data ?? [];
		} finally {
			detailLoading = false;
		}
	}

	// ─── Form actions ────────────────────────────────────────────
	function openCreate() {
		editing = null;
		form = { name: '', email: null, phone: null, address: null };
		showModal = true;
	}

	function openEdit(c: Customer) {
		editing = c;
		form = { name: c.name, email: c.email, phone: c.phone, address: c.address };
		showModal = true;
	}

	async function save() {
		if (!form.name.trim()) { notify.error('Name is required'); return; }
		submitting = true;
		try {
			if (editing) {
				await customersService.update(editing.id, form);
				notify.success('Customer updated');
				if (selected?.id === editing.id) {
					selected = { ...editing, ...form } as Customer;
				}
			} else {
				await customersService.create(form);
				notify.success('Customer added');
			}
			showModal = false;
			fetchList();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Save failed');
		} finally {
			submitting = false;
		}
	}

	async function del(c: Customer) {
		if (!confirm(`Delete "${c.name}"?`)) return;
		try {
			await customersService.delete(c.id);
			notify.success('Customer deleted');
			if (selected?.id === c.id) selected = null;
			fetchList();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Delete failed');
		}
	}

	// ─── Formatters ──────────────────────────────────────────────
	function fmt(n: number) {
		return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}
	function fmtDate(s: string) {
		return new Date(s).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
	}
	function fmtTime(s: string) {
		return new Date(s).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}
	function initials(name: string) {
		return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
	}

	async function handleExport(_fmt: 'csv', scope: 'all' | 'filtered' | 'current' | 'selected') {
		const info = await shopService.getInfo();
		const shopName = info?.shop?.name ?? 'Export';
		const userName = authStore.user?.name ?? 'System';
		let data: Customer[];
		if (scope === 'current') {
			data = customers;
		} else {
			const res = await customersService.list(scope === 'filtered' ? search : '', 1, 10000);
			data = res.data ?? [];
		}
		downloadCSV(exportCustomers(data, shopName, userName), safeFilename(shopName, 'Customers'));
	}

	onMount(fetchList);
</script>

<svelte:head><title>Customers — POS</title></svelte:head>

<div class="flex h-full overflow-hidden bg-slate-50 dark:bg-slate-950">

	<!-- ── Customer list ──────────────────────────────────────── -->
	<div class="flex flex-col flex-1 overflow-hidden min-w-0 {selected ? 'hidden lg:flex' : 'flex'}">

		<!-- Header -->
		<div class="flex items-center justify-between px-5 py-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 shrink-0 pl-3 border-l-4 border-violet-500">
			<div>
				<h1 class="text-lg font-bold text-slate-900 dark:text-slate-100">Customers</h1>
				<p class="text-xs text-slate-400 mt-0.5">{total} total</p>
			</div>
			<div class="flex gap-2">
				<button onclick={() => showExport = true} class="flex items-center gap-2 px-3 py-2 text-sm font-semibold text-slate-600 border border-slate-200 hover:bg-slate-50 transition-all">
					<Download size={14} /> Export
				</button>
				<button onclick={openCreate} class="flex items-center gap-2 px-4 py-2 text-sm font-semibold text-white transition-all active:scale-95" style="background:linear-gradient(135deg,#9333ea,#7c3aed);">
					<Plus size={15} /> Add Customer
				</button>
			</div>
		</div>

		<!-- Search -->
		<div class="px-4 py-3 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 shrink-0">
			<div class="relative">
				<Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
				<input
					bind:value={search}
					oninput={onSearch}
					placeholder="Search by name, phone, email…"
					class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100 dark:placeholder-slate-500 bg-slate-50 py-2.5 pl-9 pr-4 text-sm focus:outline-none focus:bg-white dark:focus:bg-slate-700 transition-colors"
				/>
			</div>
		</div>

		<!-- List -->
		<div class="flex-1 overflow-y-auto">
			{#if loading}
				<ul class="divide-y divide-slate-100 dark:divide-slate-800">
					{#each Array(8) as _}
						<li class="flex items-center gap-3 px-5 py-4">
							<div class="h-10 w-10 rounded-full bg-slate-100 dark:bg-slate-700 animate-pulse shrink-0"></div>
							<div class="flex-1 space-y-2">
								<div class="h-3.5 rounded bg-slate-100 dark:bg-slate-700 w-40 animate-pulse"></div>
								<div class="h-3 rounded bg-slate-100 dark:bg-slate-700 w-28 animate-pulse"></div>
							</div>
						</li>
					{/each}
				</ul>
			{:else if customers.length === 0}
				<div class="flex flex-col items-center justify-center h-full gap-3 text-slate-300 dark:text-slate-600">
					<User size={40} class="opacity-50" />
					<p class="text-sm font-medium text-slate-400">No customers found</p>
				</div>
			{:else}
				<ul class="divide-y divide-slate-100 dark:divide-slate-800">
					{#each customers as c}
						<li>
							<button
								onclick={() => selectCustomer(c)}
								class="w-full flex items-center gap-3 px-5 py-3.5 text-left transition-colors hover:bg-slate-50 dark:hover:bg-slate-800/50 {selected?.id === c.id ? 'bg-slate-50 dark:bg-slate-800/50' : ''}"
							>
								<!-- Avatar -->
								<div class="h-10 w-10 shrink-0 rounded-full flex items-center justify-center text-sm font-bold text-white" style="background-color:#3F00FF;">
									{initials(c.name)}
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">{c.name}</p>
									<p class="text-xs text-slate-400 truncate">{c.phone ?? c.email ?? 'No contact info'}</p>
								</div>
								<ChevronRight size={14} class="text-slate-300 shrink-0" />
							</button>
						</li>
					{/each}
				</ul>
				<div class="px-4 py-3">
					<Pagination {page} {total} {limit} onchange={(p) => { page = p; fetchList(); }} />
				</div>
			{/if}
		</div>
	</div>

	<!-- ── Customer detail panel ──────────────────────────────── -->
	<div class="flex flex-col bg-white dark:bg-slate-900 border-l border-slate-200 dark:border-slate-700 overflow-hidden transition-all
		{selected ? 'flex w-full lg:w-[420px] lg:shrink-0' : 'hidden lg:hidden'}">

		{#if !selected}
			<div class="flex flex-col items-center justify-center h-full gap-3 text-slate-200 dark:text-slate-700">
				<User size={48} class="opacity-30" />
				<p class="text-sm text-slate-400">Select a customer to view details</p>
			</div>
		{:else}
			<!-- Panel header -->
			<div class="flex items-center justify-between px-5 py-4 border-b border-slate-100 dark:border-slate-700 shrink-0">
				<div class="flex items-center gap-3">
					<button onclick={() => selected = null} class="lg:hidden h-8 w-8 flex items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800">
						<X size={16} />
					</button>
					<h2 class="text-base font-bold text-slate-900 dark:text-slate-100">Customer Profile</h2>
				</div>
				<div class="flex items-center gap-1">
					<button onclick={() => openEdit(selected!)} class="h-8 w-8 flex items-center justify-center rounded-lg text-slate-400 hover:text-teal-600 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
						<Pencil size={14} />
					</button>
					<button onclick={() => del(selected!)} class="h-8 w-8 flex items-center justify-center rounded-lg text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors">
						<Trash2 size={14} />
					</button>
				</div>
			</div>

			<div class="flex-1 overflow-y-auto">
				<!-- Profile section -->
				<div class="px-5 py-5 border-b border-slate-100 dark:border-slate-700">
					<div class="flex items-center gap-4 mb-4">
						<div class="h-16 w-16 shrink-0 rounded-2xl flex items-center justify-center text-xl font-bold text-white" style="background-color:#3F00FF;">
							{initials(selected.name)}
						</div>
						<div>
							<h3 class="text-lg font-bold text-slate-900 dark:text-slate-100">{selected.name}</h3>
							<p class="text-xs text-slate-400 mt-0.5">Since {fmtDate(selected.created_at)}</p>
						</div>
					</div>
					<div class="space-y-2">
						{#if selected.phone}
							<div class="flex items-center gap-2.5 text-sm text-slate-600 dark:text-slate-300">
								<Phone size={13} class="text-slate-400 shrink-0" />
								<span>{selected.phone}</span>
							</div>
						{/if}
						{#if selected.email}
							<div class="flex items-center gap-2.5 text-sm text-slate-600 dark:text-slate-300">
								<Mail size={13} class="text-slate-400 shrink-0" />
								<span>{selected.email}</span>
							</div>
						{/if}
						{#if selected.address}
							<div class="flex items-center gap-2.5 text-sm text-slate-600 dark:text-slate-300">
								<MapPin size={13} class="text-slate-400 shrink-0" />
								<span>{selected.address}</span>
							</div>
						{/if}
					</div>
				</div>

				<!-- Stats section -->
				<div class="px-5 py-4 border-b border-slate-100 dark:border-slate-700">
					<h4 class="text-xs font-bold text-slate-400 uppercase tracking-wide mb-3">Activity</h4>
					{#if detailLoading}
						<div class="grid grid-cols-2 gap-3">
							{#each Array(4) as _}
								<div class="rounded-xl bg-slate-100 dark:bg-slate-800 h-16 animate-pulse"></div>
							{/each}
						</div>
					{:else if customerStats}
						<div class="grid grid-cols-2 gap-3">
							<div class="bg-slate-50 dark:bg-slate-800 p-3">
								<p class="text-xs text-slate-400 mb-1">Total Orders</p>
								<p class="text-xl font-bold text-slate-900 dark:text-slate-100">{customerStats.total_orders}</p>
							</div>
							<div class="bg-slate-50 dark:bg-slate-800 p-3">
								<p class="text-xs text-slate-400 mb-1">Lifetime Spend</p>
								<p class="text-lg font-bold" style="color:#3F00FF;">KES {fmt(customerStats.lifetime_spend)}</p>
							</div>
							<div class="bg-slate-50 dark:bg-slate-800 p-3">
								<p class="text-xs text-slate-400 mb-1">Avg Order</p>
								<p class="text-lg font-bold text-slate-900 dark:text-slate-100">KES {fmt(customerStats.avg_order)}</p>
							</div>
							<div class="bg-slate-50 dark:bg-slate-800 p-3">
								<p class="text-xs text-slate-400 mb-1">Last Visit</p>
								<p class="text-sm font-semibold text-slate-700 dark:text-slate-200">
									{customerStats.last_visit ? fmtDate(customerStats.last_visit) : 'Never'}
								</p>
							</div>
						</div>
					{/if}
				</div>

				<!-- Purchase history -->
				<div class="px-5 py-4">
					<h4 class="text-xs font-bold text-slate-400 uppercase tracking-wide mb-3">Purchase History</h4>
					{#if detailLoading}
						<div class="space-y-2">
							{#each Array(4) as _}
								<div class="h-12 rounded-xl bg-slate-100 dark:bg-slate-800 animate-pulse"></div>
							{/each}
						</div>
					{:else if purchaseHistory.length === 0}
						<div class="flex flex-col items-center gap-2 py-8 text-slate-300 dark:text-slate-600">
							<ShoppingBag size={28} class="opacity-50" />
							<p class="text-xs text-slate-400">No purchases yet</p>
						</div>
					{:else}
						<ul class="space-y-2">
							{#each purchaseHistory as sale}
								<li class="flex items-center justify-between gap-3 bg-slate-50 dark:bg-slate-800 px-3.5 py-3">
									<div class="min-w-0">
										<p class="text-xs font-bold text-slate-700 dark:text-slate-200 font-mono">#{sale.id.slice(0, 8).toUpperCase()}</p>
										<p class="text-[10px] text-slate-400 mt-0.5">{fmtTime(sale.created_at)}</p>
									</div>
									<div class="text-right shrink-0">
										<p class="text-sm font-bold text-slate-800 dark:text-slate-100 tabular-nums">KES {fmt(sale.total)}</p>
										<p class="text-[10px] text-slate-400 capitalize">{sale.payment_method}</p>
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- ─── Add / Edit Modal ─────────────────────────────────────── -->
<Modal open={showModal} title={editing ? 'Edit Customer' : 'Add Customer'} onclose={() => showModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-3.5">
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Name *</label>
				<input bind:value={form.name} placeholder="Full name" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Phone</label>
				<input type="tel" bind:value={form.phone} placeholder="+254…" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Email</label>
				<input type="email" bind:value={form.email} placeholder="email@example.com" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">Address</label>
				<textarea bind:value={form.address} rows={2} placeholder="Physical address" class="w-full rounded-xl border border-slate-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 px-3.5 py-2.5 text-sm focus:outline-none resize-none"></textarea>
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-xl border border-slate-200 dark:border-slate-600 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-xl px-5 py-2.5 text-sm font-semibold text-white disabled:opacity-50 transition-all active:scale-95" style="background-color:#00008B;">
			{submitting ? 'Saving…' : editing ? 'Save Changes' : 'Add Customer'}
		</button>
	{/snippet}
</Modal>

<ExportModal
	open={showExport}
	title="Export Customers"
	hasFiltered={!!search}
	onclose={() => showExport = false}
	onexport={handleExport}
/>
