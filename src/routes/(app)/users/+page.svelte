<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/services/api';
	import { notify } from '$lib/stores/notification.svelte';
	import { authStore } from '$lib/stores/auth.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import type { User, Role } from '$lib/types';
	import { Plus, Pencil, Trash2, ShieldCheck } from '@lucide/svelte';

	let users = $state<User[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editing = $state<User | null>(null);
	let submitting = $state(false);

	let form = $state({ name: '', email: '', password: '', role: 'cashier' as Role, is_active: true });

	async function fetch() {
		loading = true;
		try {
			const res = await api.get<{ success: boolean; data: User[] }>('/users');
			users = res.data ?? [];
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		editing = null;
		form = { name: '', email: '', password: '', role: 'cashier', is_active: true };
		showModal = true;
	}

	function openEdit(u: User) {
		editing = u;
		form = { name: u.name, email: u.email, password: '', role: u.role, is_active: u.is_active };
		showModal = true;
	}

	async function save() {
		if (!form.name.trim() || !form.email.trim()) { notify.error('Name and email are required'); return; }
		if (!editing && !form.password) { notify.error('Password is required for new users'); return; }
		submitting = true;
		try {
			if (editing) {
				await api.put(`/users/${editing.id}`, form);
				notify.success('User updated');
			} else {
				await api.post('/users', form);
				notify.success('User created');
			}
			showModal = false;
			fetch();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Save failed');
		} finally {
			submitting = false;
		}
	}

	async function deactivate(u: User) {
		if (u.id === authStore.user?.id) { notify.error("Can't deactivate your own account"); return; }
		if (!confirm(`Deactivate "${u.name}"?`)) return;
		try {
			await api.delete(`/users/${u.id}`);
			notify.success('User deactivated');
			fetch();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed');
		}
	}

	const roleColors: Record<Role, string> = {
		admin: 'bg-purple-50 text-purple-700',
		manager: 'bg-blue-50 text-blue-700',
		cashier: 'bg-gray-50 text-gray-700'
	};

	onMount(fetch);
</script>

<svelte:head><title>Users — POS</title></svelte:head>

<div class="p-6 space-y-5">
	<div class="flex items-center justify-between">
		<h1 class="text-xl font-semibold text-gray-900">Users</h1>
		<button onclick={openCreate} class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
			<Plus size={16} />Add User
		</button>
	</div>

	<div class="rounded-xl shadow-sm bg-white overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				<tr class="bg-slate-50">
					<th class="px-4 py-3 font-medium text-gray-600">Name</th>
					<th class="px-4 py-3 font-medium text-gray-600">Email</th>
					<th class="px-4 py-3 font-medium text-gray-600">Role</th>
					<th class="px-4 py-3 font-medium text-gray-600">Status</th>
					<th class="px-4 py-3 font-medium text-gray-600 text-right">Actions</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100">
				{#if loading}
					{#each Array(4) as _}
						<tr>{#each Array(5) as _}<td class="px-4 py-3"><div class="h-4 bg-gray-100 rounded animate-pulse"></div></td>{/each}</tr>
					{/each}
				{:else if users.length === 0}
					<tr><td colspan="5" class="px-4 py-12 text-center text-gray-400">No users found</td></tr>
				{:else}
					{#each users as u}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3">
								<div class="flex items-center gap-2">
									{#if u.id === authStore.user?.id}
										<ShieldCheck size={14} class="text-blue-500" />
									{/if}
									<span class="font-medium text-gray-900">{u.name}</span>
								</div>
							</td>
							<td class="px-4 py-3 text-gray-500">{u.email}</td>
							<td class="px-4 py-3">
								<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {roleColors[u.role]} capitalize">
									{u.role}
								</span>
							</td>
							<td class="px-4 py-3">
								<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {u.is_active ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}">
									{u.is_active ? 'Active' : 'Inactive'}
								</span>
							</td>
							<td class="px-4 py-3 text-right">
								<div class="flex items-center justify-end gap-2">
									<button onclick={() => openEdit(u)} class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-blue-600"><Pencil size={14} /></button>
									{#if u.id !== authStore.user?.id}
										<button onclick={() => deactivate(u)} class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-red-600"><Trash2 size={14} /></button>
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>

<Modal open={showModal} title={editing ? 'Edit User' : 'Add User'} onclose={() => showModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-3">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
				<input bind:value={form.name} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Email *</label>
				<input type="email" bind:value={form.email} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Password {editing ? '(leave blank to keep)' : '*'}</label>
				<input type="password" bind:value={form.password} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Role</label>
				<select bind:value={form.role} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none">
					<option value="cashier">Cashier</option>
					<option value="manager">Manager</option>
					<option value="admin">Admin</option>
				</select>
			</div>
			{#if editing}
				<label class="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
					<input type="checkbox" bind:checked={form.is_active} class="rounded" />
					Active
				</label>
			{/if}
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-lg border px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60">
			{submitting ? 'Saving…' : editing ? 'Save Changes' : 'Create User'}
		</button>
	{/snippet}
</Modal>
