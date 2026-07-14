<script lang="ts">
	import { onMount } from 'svelte';
	import { salesService } from '$lib/services/sales';
	import { notify } from '$lib/stores/notification.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import type { Sale } from '$lib/types';
	import { Eye } from '@lucide/svelte';

	let sales = $state<Sale[]>([]);
	let total = $state(0);
	let page = $state(1);
	const limit = 20;
	let loading = $state(true);

	let statusFilter = $state('');
	let dateFrom = $state('');
	let dateTo = $state('');

	let selectedSale = $state<Sale | null>(null);
	let showDetail = $state(false);

	function fmt(n: number) {
		return new Intl.NumberFormat('en-KE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n);
	}

	function fmtDate(s: string) {
		return new Date(s).toLocaleString('en-US', {
			year: 'numeric', month: 'short', day: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}

	async function fetch() {
		loading = true;
		try {
			const res = await salesService.list({
				status: statusFilter as 'completed' | 'held' | 'voided' | undefined || undefined,
				date_from: dateFrom || undefined,
				date_to: dateTo || undefined,
				page,
				limit
			});
			sales = res.data ?? [];
			total = res.total ?? 0;
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to load sales');
		} finally {
			loading = false;
		}
	}

	async function viewDetail(sale: Sale) {
		try {
			const res = await salesService.getById(sale.id);
			selectedSale = res.data ?? null;
			showDetail = true;
		} catch {
			notify.error('Failed to load sale details');
		}
	}

	async function voidSale(id: string) {
		if (!confirm('Void this sale? Stock will be returned.')) return;
		try {
			await salesService.void(id);
			notify.success('Sale voided');
			fetch();
		} catch (err) {
			notify.error(err instanceof Error ? err.message : 'Failed to void sale');
		}
	}

	const statusColors: Record<string, string> = {
		completed: 'bg-green-50 text-green-700',
		held: 'bg-yellow-50 text-yellow-700',
		voided: 'bg-red-50 text-red-700'
	};

	onMount(fetch);
</script>

<svelte:head><title>Sales History — POS</title></svelte:head>

<div class="p-6 space-y-5">
	<h1 class="text-xl font-semibold text-gray-900">Sales History</h1>

	<div class="flex gap-3 flex-wrap">
		<select
			bind:value={statusFilter}
			onchange={() => { page = 1; fetch(); }}
			class="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
		>
			<option value="">All Status</option>
			<option value="completed">Completed</option>
			<option value="held">Held</option>
			<option value="voided">Voided</option>
		</select>
		<input
			type="date"
			bind:value={dateFrom}
			onchange={() => { page = 1; fetch(); }}
			class="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
		/>
		<input
			type="date"
			bind:value={dateTo}
			onchange={() => { page = 1; fetch(); }}
			class="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
		/>
	</div>

		<div class="rounded-xl shadow-sm bg-white overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="bg-slate-50">
					<th class="px-4 py-3 font-medium text-gray-600">Date</th>
					<th class="px-4 py-3 font-medium text-gray-600">Cashier</th>
					<th class="px-4 py-3 font-medium text-gray-600">Customer</th>
					<th class="px-4 py-3 font-medium text-gray-600">Payment</th>
					<th class="px-4 py-3 font-medium text-gray-600">Status</th>
					<th class="px-4 py-3 font-medium text-gray-600 text-right">Total</th>
					<th class="px-4 py-3 font-medium text-gray-600 text-right">Actions</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-100">
				{#if loading}
					{#each Array(8) as _}
						<tr>{#each Array(7) as _}<td class="px-4 py-3"><div class="h-4 bg-gray-100 rounded animate-pulse"></div></td>{/each}</tr>
					{/each}
				{:else if sales.length === 0}
					<tr><td colspan="7" class="px-4 py-12 text-center text-gray-400">No sales found</td></tr>
				{:else}
					{#each sales as s}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-gray-600">{fmtDate(s.created_at)}</td>
							<td class="px-4 py-3 text-gray-700">{s.cashier_name ?? '—'}</td>
							<td class="px-4 py-3 text-gray-500">{s.customer_name ?? 'Walk-in'}</td>
							<td class="px-4 py-3 text-gray-500 capitalize">{s.payment_method}</td>
							<td class="px-4 py-3">
								<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {statusColors[s.status] ?? ''}">
									{s.status}
								</span>
							</td>
							<td class="px-4 py-3 text-right font-semibold text-gray-900">KES {fmt(s.total)}</td>
							<td class="px-4 py-3 text-right">
								<div class="flex items-center justify-end gap-1">
									<button onclick={() => viewDetail(s)} class="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-blue-600">
										<Eye size={14} />
									</button>
									{#if s.status !== 'voided'}
										<button onclick={() => voidSale(s.id)} class="rounded px-2 py-1 text-xs text-red-600 hover:bg-red-50">
											Void
										</button>
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<Pagination {page} {total} {limit} onchange={(p) => { page = p; fetch(); }} />
</div>

<Modal open={showDetail} title="Sale Details" onclose={() => showDetail = false} size="md">
	{#snippet children()}
		{#if selectedSale}
			<div class="space-y-4">
				<div class="grid grid-cols-2 gap-3 text-sm">
					<div><span class="text-gray-500">Date:</span> <span class="font-medium">{fmtDate(selectedSale.created_at)}</span></div>
					<div><span class="text-gray-500">Status:</span> <span class="font-medium capitalize">{selectedSale.status}</span></div>
					<div><span class="text-gray-500">Cashier:</span> <span class="font-medium">{selectedSale.cashier_name ?? '—'}</span></div>
					<div><span class="text-gray-500">Customer:</span> <span class="font-medium">{selectedSale.customer_name ?? 'Walk-in'}</span></div>
					<div><span class="text-gray-500">Payment:</span> <span class="font-medium capitalize">{selectedSale.payment_method}</span></div>
				</div>

				{#if selectedSale.items && selectedSale.items.length > 0}
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b text-left text-gray-500">
								<th class="pb-2 font-medium">Item</th>
								<th class="pb-2 font-medium text-right">Qty</th>
								<th class="pb-2 font-medium text-right">Price</th>
								<th class="pb-2 font-medium text-right">Total</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100">
							{#each selectedSale.items as item}
								<tr>
									<td class="py-2 text-gray-800">{item.product_name ?? '—'}</td>
									<td class="py-2 text-right text-gray-600">{item.quantity}</td>
									<td class="py-2 text-right text-gray-600">KES {fmt(item.unit_price)}</td>
									<td class="py-2 text-right font-medium">KES {fmt(item.total)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				{/if}

				<div class="border-t pt-3 space-y-1.5 text-sm">
					<div class="flex justify-between text-gray-600"><span>Subtotal</span><span>KES {fmt(selectedSale.subtotal)}</span></div>
					{#if selectedSale.discount > 0}
						<div class="flex justify-between text-green-600"><span>Discount</span><span>-KES {fmt(selectedSale.discount)}</span></div>
					{/if}
					{#if selectedSale.tax > 0}
						<div class="flex justify-between text-gray-600"><span>Tax</span><span>KES {fmt(selectedSale.tax)}</span></div>
					{/if}
					<div class="flex justify-between font-bold text-gray-900 text-base border-t pt-2">
						<span>Total</span><span>KES {fmt(selectedSale.total)}</span>
					</div>
				</div>
			</div>
		{/if}
	{/snippet}
</Modal>
