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
		admin:   'badge-purple',
		manager: 'badge-blue',
		cashier: 'badge-slate'
	};

	onMount(fetch);
</script>

<svelte:head><title>Users — Maestro POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-5 min-h-full dark:bg-slate-950">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-lg font-bold text-slate-900 dark:text-slate-100">Users</h1>
			<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Manage system access and roles</p>
		</div>
		<button onclick={openCreate} class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-blue-700 transition-colors">
			<Plus size={13} />Add User
		</button>
	</div>

	<div class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b border-slate-100 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/40">
					<th class="px-5 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">User</th>
					<th class="px-5 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden sm:table-cell">Email</th>
					<th class="px-5 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Role</th>
					<th class="px-5 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden md:table-cell">Status</th>
					<th class="px-5 py-3 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Actions</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
				{#if loading}
					{#each Array(4) as _}
						<tr>{#each Array(5) as _}<td class="px-5 py-3"><div class="h-3.5 bg-slate-100 dark:bg-slate-700 rounded animate-pulse"></div></td>{/each}</tr>
					{/each}
				{:else if users.length === 0}
					<tr><td colspan="5" class="px-5 py-12 text-center text-sm text-slate-400 dark:text-slate-500">No users found</td></tr>
				{:else}
					{#each users as u}
						<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-colors">
							<td class="px-5 py-3">
								<div class="flex items-center gap-3">
									<div class="h-8 w-8 shrink-0 flex items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900/40 text-xs font-bold text-blue-700 dark:text-blue-300">
										{u.name.charAt(0).toUpperCase()}
									</div>
									<div class="min-w-0">
										<div class="flex items-center gap-1.5">
											<span class="font-semibold text-slate-900 dark:text-slate-100 truncate">{u.name}</span>
											{#if u.id === authStore.user?.id}
												<ShieldCheck size={12} class="text-blue-500 shrink-0" />
											{/if}
										</div>
									</div>
								</div>
							</td>
							<td class="px-5 py-3 text-slate-500 dark:text-slate-400 hidden sm:table-cell">{u.email}</td>
							<td class="px-5 py-3">
								<span class="badge {roleColors[u.role]} capitalize">{u.role}</span>
							</td>
							<td class="px-5 py-3 hidden md:table-cell">
								<span class="badge {u.is_active ? 'badge-green' : 'badge-red'}">{u.is_active ? 'Active' : 'Inactive'}</span>
							</td>
							<td class="px-5 py-3 text-right">
								<div class="flex items-center justify-end gap-1">
									<button onclick={() => openEdit(u)} class="h-7 w-7 flex items-center justify-center rounded-md text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 hover:text-blue-600 transition-colors"><Pencil size={13} /></button>
									{#if u.id !== authStore.user?.id}
										<button onclick={() => deactivate(u)} class="h-7 w-7 flex items-center justify-center rounded-md text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 hover:text-red-600 transition-colors"><Trash2 size={13} /></button>
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
		<div class="space-y-4">
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Name *</label>
				<input bind:value={form.name} class="w-full rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" placeholder="Full name" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Email *</label>
				<input type="email" bind:value={form.email} class="w-full rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" placeholder="user@example.com" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Password {editing ? '(leave blank to keep)' : '*'}</label>
				<input type="password" bind:value={form.password} class="w-full rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" placeholder="••••••••" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Role</label>
				<select bind:value={form.role} class="w-full rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors">
					<option value="cashier">Cashier</option>
					<option value="manager">Manager</option>
					<option value="admin">Admin</option>
				</select>
			</div>
			{#if editing}
				<label class="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300 cursor-pointer">
					<input type="checkbox" bind:checked={form.is_active} class="rounded accent-blue-600" />
					Account active
				</label>
			{/if}
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-lg border border-slate-200 dark:border-slate-600 px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
		<button onclick={save} disabled={submitting} class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-60 transition-colors">
			{submitting ? 'Saving…' : editing ? 'Save Changes' : 'Create User'}
		</button>
	{/snippet}
</Modal>
