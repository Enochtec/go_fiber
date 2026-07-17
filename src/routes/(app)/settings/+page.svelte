<script lang="ts">
	import { onMount } from 'svelte';
	import { productsService } from '$lib/services/products';
	import { notify } from '$lib/stores/notification.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import type { Category } from '$lib/types';
	import { Plus, Pencil, Trash2 } from '@lucide/svelte';

	let categories = $state<Category[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editing = $state<Category | null>(null);
	let submitting = $state(false);
	let catName = $state('');
	let catDesc = $state('');

	async function fetchCats() {
		loading = true;
		try {
			const res = await productsService.listCategories();
			categories = res.data ?? [];
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		editing = null;
		catName = '';
		catDesc = '';
		showModal = true;
	}

	function openEdit(c: Category) {
		editing = c;
		catName = c.name;
		catDesc = c.description ?? '';
		showModal = true;
	}

	async function saveCategory() {
		if (!catName.trim()) { notify.error('Name is required'); return; }
		submitting = true;
		try {
			if (editing) {
				await productsService.updateCategory(editing.id, catName, catDesc || undefined);
				notify.success('Category updated');
			} else {
				await productsService.createCategory(catName, catDesc || undefined);
				notify.success('Category created');
			}
			showModal = false;
			fetchCats();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Save failed');
		} finally {
			submitting = false;
		}
	}

	async function deleteCategory(c: Category) {
		if (!confirm(`Delete category "${c.name}"? Products will be uncategorized.`)) return;
		try {
			await productsService.deleteCategory(c.id);
			notify.success('Category deleted');
			fetchCats();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Delete failed');
		}
	}

	onMount(fetchCats);
</script>

<svelte:head><title>Settings — Maestro POS</title></svelte:head>

<div class="p-4 md:p-6 space-y-6 max-w-3xl dark:bg-slate-950 min-h-full">
	<div>
		<h1 class="text-lg font-bold text-slate-900 dark:text-slate-100">Settings</h1>
		<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Manage system configuration</p>
	</div>

	<section class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden">
		<div class="flex items-center justify-between px-5 py-4 border-b border-slate-100 dark:border-slate-700">
			<div>
				<h2 class="text-sm font-semibold text-slate-800 dark:text-slate-100">Product Categories</h2>
				<p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Organize products into categories</p>
			</div>
			<button onclick={openCreate} class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-blue-700 transition-colors">
				<Plus size={13} />Add Category
			</button>
		</div>

		<table class="w-full text-sm">
			<thead>
				<tr class="border-b border-slate-100 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/40">
					<th class="px-5 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Name</th>
					<th class="px-5 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Description</th>
					<th class="px-5 py-3 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Actions</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100 dark:divide-slate-700">
				{#if loading}
					{#each Array(3) as _}
						<tr>{#each Array(3) as _}<td class="px-5 py-3"><div class="h-3.5 bg-slate-100 dark:bg-slate-700 rounded animate-pulse"></div></td>{/each}</tr>
					{/each}
				{:else if categories.length === 0}
					<tr><td colspan="3" class="px-5 py-12 text-center text-sm text-slate-400 dark:text-slate-500">No categories yet. Add one to get started.</td></tr>
				{:else}
					{#each categories as c}
						<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/40 transition-colors">
							<td class="px-5 py-3 font-medium text-slate-900 dark:text-slate-100">{c.name}</td>
							<td class="px-5 py-3 text-slate-500 dark:text-slate-400">{c.description ?? '—'}</td>
							<td class="px-5 py-3 text-right">
								<div class="flex items-center justify-end gap-1">
									<button onclick={() => openEdit(c)} class="h-7 w-7 flex items-center justify-center rounded-md text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 hover:text-blue-600 transition-colors"><Pencil size={13} /></button>
									<button onclick={() => deleteCategory(c)} class="h-7 w-7 flex items-center justify-center rounded-md text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 hover:text-red-600 transition-colors"><Trash2 size={13} /></button>
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</section>
</div>

<Modal open={showModal} title={editing ? 'Edit Category' : 'Add Category'} onclose={() => showModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-4">
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Name *</label>
				<input bind:value={catName} class="w-full rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" placeholder="e.g. Beverages" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Description</label>
				<input bind:value={catDesc} class="w-full rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:border-blue-500 focus:outline-none transition-colors" placeholder="Optional description" />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-lg border border-slate-200 dark:border-slate-600 px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
		<button onclick={saveCategory} disabled={submitting} class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-60 transition-colors">
			{submitting ? 'Saving…' : editing ? 'Save Changes' : 'Create'}
		</button>
	{/snippet}
</Modal>
