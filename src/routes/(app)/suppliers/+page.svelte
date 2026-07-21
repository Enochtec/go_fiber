<script lang="ts">
	import { onMount } from 'svelte';
	import { offlineSuppliers as suppliersService } from '$lib/services/offline';
	import type { SupplierInput } from '$lib/services/suppliers';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import type { Supplier } from '$lib/types';
	import { Plus, Pencil, Trash2, Download } from '@lucide/svelte';
	import ExportModal from '$lib/components/ExportModal.svelte';
	import { shopService } from '$lib/services/shop';
	import { authStore } from '$lib/stores/auth.svelte';
	import { exportSuppliers } from '$lib/services/export';

	let suppliers = $state<Supplier[]>([]);
	let loading = $state(true);
	let showExport = $state(false);
	let showModal = $state(false);
	let editing = $state<Supplier | null>(null);
	let submitting = $state(false);

	let form = $state<SupplierInput>({ name: '', email: null, phone: null, address: null });

	async function fetch() {
		loading = true;
		try {
			const res = await suppliersService.list();
			suppliers = res.data ?? [];
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		editing = null;
		form = { name: '', email: null, phone: null, address: null };
		showModal = true;
	}

	function openEdit(s: Supplier) {
		editing = s;
		form = { name: s.name, email: s.email, phone: s.phone, address: s.address };
		showModal = true;
	}

	async function save() {
		if (!form.name.trim()) { notify.error('Name is required'); return; }
		submitting = true;
		try {
			if (editing) {
				await suppliersService.update(editing.id, form);
				notify.success('Supplier updated');
			} else {
				await suppliersService.create(form);
				notify.success('Supplier created');
			}
			showModal = false;
			fetch();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Save failed');
		} finally {
			submitting = false;
		}
	}

	async function del(s: Supplier) {
		if (!confirm(`Delete "${s.name}"?`)) return;
		try {
			await suppliersService.delete(s.id);
			notify.success('Supplier deleted');
			fetch();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Delete failed');
		}
	}

	async function handleExport(_fmt: 'csv', _scope: 'all' | 'filtered' | 'current' | 'selected') {
		const info = await shopService.getInfo();
		const shopName = info?.shop?.name ?? 'Export';
		const userName = authStore.user?.name ?? 'System';
		await exportSuppliers(suppliers, shopName, userName);
	}

	onMount(fetch);
</script>

<svelte:head><title>Suppliers — Maestro POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-5 min-h-full dark:bg-slate-950">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-lg font-bold text-slate-900 dark:text-slate-100">Suppliers</h1>
			<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Manage your product suppliers</p>
		</div>
		<div class="flex gap-2">
			<button onclick={() => showExport = true} class="flex items-center gap-1.5 rounded-[1px] border border-slate-200 px-3 py-1.5 text-xs font-semibold text-slate-600 hover:bg-slate-50 transition-colors">
				<Download size={13} /> Export
			</button>
			<button onclick={openCreate} class="flex items-center gap-1.5 rounded-[1px] bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-blue-700 transition-colors">
				<Plus size={13} />Add Supplier
			</button>
		</div>
	</div>

	<div class="rounded-[1px] bg-white dark:bg-slate-800 overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				<tr style="background:linear-gradient(135deg,#2563eb,#1d4ed8);">
					<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide">Name</th>
					<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide hidden sm:table-cell">Phone</th>
					<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide hidden md:table-cell">Email</th>
					<th class="px-4 py-2.5 text-left text-xs font-semibold text-white uppercase tracking-wide hidden lg:table-cell">Address</th>
					<th class="px-4 py-2.5 text-right text-xs font-semibold text-white uppercase tracking-wide">Actions</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
				{#if loading}
					{#each Array(4) as _}
						<tr>{#each Array(5) as _}<td class="px-4 py-2.5"><div class="h-3.5 bg-slate-100 dark:bg-slate-700 rounded-[1px] animate-pulse"></div></td>{/each}</tr>
					{/each}
				{:else if suppliers.length === 0}
					<tr><td colspan="5" class="px-5 py-12 text-center text-sm text-slate-400 dark:text-slate-500">No suppliers yet. Add one to get started.</td></tr>
				{:else}
					{#each suppliers as s}
						<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-colors">
							<td class="px-4 py-2.5 font-semibold text-slate-900 dark:text-slate-100">{s.name}</td>
							<td class="px-4 py-2.5 text-slate-500 dark:text-slate-400 hidden sm:table-cell">{s.phone ?? '—'}</td>
							<td class="px-4 py-2.5 text-slate-500 dark:text-slate-400 hidden md:table-cell">{s.email ?? '—'}</td>
							<td class="px-4 py-2.5 text-slate-500 dark:text-slate-400 truncate max-w-48 hidden lg:table-cell">{s.address ?? '—'}</td>
							<td class="px-4 py-2.5 text-right">
								<div class="flex items-center justify-end gap-1">
									<button onclick={() => openEdit(s)} class="h-7 w-7 flex items-center justify-center rounded-[1px] text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 hover:text-blue-600 transition-colors"><Pencil size={13} /></button>
									<button onclick={() => del(s)} class="h-7 w-7 flex items-center justify-center rounded-[1px] text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 hover:text-red-600 transition-colors"><Trash2 size={13} /></button>
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>

<Modal open={showModal} title={editing ? 'Edit Supplier' : 'Add Supplier'} onclose={() => showModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-4">
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Name *</label>
				<input bind:value={form.name} class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" placeholder="Supplier name" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Phone</label>
				<input bind:value={form.phone} class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" placeholder="+254 700 000 000" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Email</label>
				<input type="email" bind:value={form.email} class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" placeholder="supplier@example.com" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Address</label>
				<textarea bind:value={form.address} rows="2" class="w-full rounded-[1px] border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors resize-none" placeholder="Physical address"></textarea>
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-[1px] border border-slate-200 dark:border-slate-600 px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-[1px] bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-60 transition-colors">
			{submitting ? 'Saving…' : editing ? 'Save Changes' : 'Create'}
		</button>
	{/snippet}
</Modal>

<ExportModal
	open={showExport}
	title="Export Suppliers"
	onclose={() => showExport = false}
	onexport={handleExport}
/>
