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

<svelte:head><title>Settings — POS</title></svelte:head>

<div class="p-6 space-y-8 max-w-3xl">
	<h1 class="text-xl font-semibold text-gray-900">Settings</h1>

	<section class="space-y-4">
		<div class="flex items-center justify-between">
			<div>
				<h2 class="text-base font-medium text-gray-900">Product Categories</h2>
				<p class="text-sm text-gray-500 mt-0.5">Organize products into categories.</p>
			</div>
			<button onclick={openCreate} class="flex items-center gap-2 rounded-lg bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700">
				<Plus size={14} />Add Category
			</button>
		</div>

		<div class="rounded-xl shadow-sm bg-white overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="bg-slate-50">
						<th class="px-4 py-3 font-medium text-gray-600">Name</th>
						<th class="px-4 py-3 font-medium text-gray-600">Description</th>
						<th class="px-4 py-3 font-medium text-gray-600 text-right">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-slate-100">
					{#if loading}
						{#each Array(3) as _}
							<tr>{#each Array(3) as _}<td class="px-4 py-3"><div class="h-4 bg-gray-100 rounded animate-pulse"></div></td>{/each}</tr>
						{/each}
					{:else if categories.length === 0}
						<tr><td colspan="3" class="px-4 py-10 text-center text-gray-400">No categories yet</td></tr>
					{:else}
						{#each categories as c}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-3 font-medium text-gray-900">{c.name}</td>
								<td class="px-4 py-3 text-gray-500">{c.description ?? '—'}</td>
								<td class="px-4 py-3 text-right">
									<div class="flex items-center justify-end gap-2">
										<button onclick={() => openEdit(c)} class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-blue-600"><Pencil size={14} /></button>
										<button onclick={() => deleteCategory(c)} class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-red-600"><Trash2 size={14} /></button>
									</div>
								</td>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</section>
</div>

<Modal open={showModal} title={editing ? 'Edit Category' : 'Add Category'} onclose={() => showModal = false} size="sm">
	{#snippet children()}
		<div class="space-y-3">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
				<input bind:value={catName} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
				<input bind:value={catDesc} class="w-full rounded-lg border px-3 py-2 text-sm focus:border-blue-500 focus:outline-none" />
			</div>
		</div>
	{/snippet}
	{#snippet footer()}
		<button onclick={() => showModal = false} class="rounded-lg border px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">Cancel</button>
		<button onclick={saveCategory} disabled={submitting} class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-60">
			{submitting ? 'Saving…' : editing ? 'Save' : 'Create'}
		</button>
	{/snippet}
</Modal>
