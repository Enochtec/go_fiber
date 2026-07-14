<script lang="ts">
	import { onMount } from 'svelte';
	import { suppliersService, type SupplierInput } from '$lib/services/suppliers';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import type { Supplier } from '$lib/types';
	import { Plus, Pencil, Trash2 } from '@lucide/svelte';

	let suppliers = $state<Supplier[]>([]);
	let loading = $state(true);
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

	onMount(fetch);
</script>

<svelte:head><title>Suppliers — POS</title></svelte:head>

<div class="p-6 space-y-5">
	<div class="flex items-center justify-between">
		<h1 class="text-xl font-semibold text-gray-900">Suppliers</h1>
		<button onclick={openCreate} class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
			<Plus size={16} />Add Supplier
		</button>
	</div>

	<div class="rounded-xl shadow-sm bg-white overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				<tr class="bg-slate-50">
					<th class="px-4 py-3 font-medium text-gray-600">Name</th>
					<th class="px-4 py-3 font-medium text-gray-600">Phone</th>
					<th class="px-4 py-3 font-medium text-gray-600">Email</th>
					<th class="px-4 py-3 font-medium text-gray-600">Address</th>
					<th class="px-4 py-3 font-medium text-gray-600 text-right">Actions</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100">
				{#if loading}
					{#each Array(4) as _}
						<tr>{#each Array(5) as _}<td class="px-4 py-3"><div class="h-4 bg-gray-100 rounded animate-pulse"></div></td>{/each}</tr>
					{/each}
				{:else if suppliers.length === 0}
					<tr><td colspan="5" class="px-4 py-12 text-center text-gray-400">No suppliers yet</td></tr>
				{:else}
					{#each suppliers as s}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 font-medium text-gray-900">{s.name}</td>
							<td class="px-4 py-3 text-gray-500">{s.phone ?? '—'}</td>
							<td class="px-4 py-3 text-gray-500">{s.email ?? '—'}</td>
							<td class="px-4 py-3 text-gray-500 truncate max-w-48">{s.address ?? '—'}</td>
							<td class="px-4 py-3 text-right">
								<div class="flex items-center justify-end gap-2">
									<button onclick={() => openEdit(s)} class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-blue-600"><Pencil size={14} /></button>
									<button onclick={() => del(s)} class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-red-600"><Trash2 size={14} /></button>
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
		<div class="space-y-3">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
				<input bind:value={form.name} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Phone</label>
				<input bind:value={form.phone} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
				<input type="email" bind:value={form.email} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
				<textarea bind:value={form.address} rows="2" class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none resize-none"></textarea>
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-lg border px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60">
			{submitting ? 'Saving…' : editing ? 'Save Changes' : 'Create'}
		</button>
	{/snippet}
</Modal>
