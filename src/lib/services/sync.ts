import { db, type SyncQueueItem } from './db';
import { connectionStatus, type ConnectionStatus } from './connection';
import { api } from './api';
import { authStore } from '$lib/stores/auth.svelte';

export const syncState = $state<{
	pending: number;
	syncing: boolean;
	lastSync: string | null;
	lastError: string | null;
	progress: { current: number; total: number };
}>({
	pending: 0,
	syncing: false,
	lastSync: null,
	lastError: null,
	progress: { current: 0, total: 0 }
});

const RETRY_DELAYS = [1000, 3000, 10_000, 30_000];
const MAX_RETRIES = RETRY_DELAYS.length;
const BATCH_SIZE = 10;
const SYNC_INTERVAL = 5000;

let syncTimer: ReturnType<typeof setInterval> | null = null;
let isSyncing = false;

function endpointFor(entity: string, action: string, entityId: string): string {
	switch (entity) {
		case 'sale':
			return action === 'create' ? '/sales' : `/sales/${entityId}`;
		case 'customer':
			return action === 'create' ? '/customers' : `/customers/${entityId}`;
		case 'product':
			return action === 'create' ? '/products' : `/products/${entityId}`;
		case 'supplier':
			return action === 'create' ? '/suppliers' : `/suppliers/${entityId}`;
		case 'category':
			return action === 'create' ? '/categories' : `/categories/${entityId}`;
		default:
			return '';
	}
}

function apiMethod(action: string) {
	if (action === 'create') return 'post';
	if (action === 'update') return 'put';
	return 'delete';
}

async function processItem(item: SyncQueueItem): Promise<boolean> {
	const endpoint = endpointFor(item.entity, item.action, item.entityId);
	if (!endpoint) return true;

	try {
		await db.syncQueue.update(item.id!, { status: 'syncing' });

		const method = apiMethod(item.action);
		if (method === 'delete') {
			await api.delete(endpoint);
		} else {
			await (api as any)[method](endpoint, item.payload);
		}

		await db.syncQueue.delete(item.id!);

		if (item.entity === 'sale' && item.payload && typeof item.payload === 'object' && 'local_id' in (item.payload as any)) {
			const localId = (item.payload as any).local_id;
			const offlineSale = await db.sales.get(localId);
			if (offlineSale) {
				await db.sales.update(localId, { synced: true, synced_at: new Date().toISOString() });
			}
		}

		return true;
	} catch (err: any) {
		const newRetryCount = (item.retry_count || 0) + 1;
		const status = newRetryCount >= MAX_RETRIES ? 'failed' : 'pending';
		await db.syncQueue.update(item.id!, {
			status,
			retry_count: newRetryCount,
			last_error: err?.message || String(err)
		});
		syncState.lastError = err?.message || String(err);
		return false;
	}
}

export async function processSyncQueue() {
	if (isSyncing) return;
	isSyncing = true;
	syncState.syncing = true;

	try {
		const pending = await db.syncQueue
			.where('status')
			.anyOf('pending', 'failed')
			.toArray();

		if (pending.length === 0) {
			syncState.syncing = false;
			syncState.pending = 0;
			return;
		}

		syncState.progress = { current: 0, total: pending.length };
		connectionStatus.set('syncing' as any);

		const sorted = pending.sort(
			(a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
		);

		for (let i = 0; i < sorted.length; i += BATCH_SIZE) {
			const batch = sorted.slice(i, i + BATCH_SIZE);
			const results = await Promise.allSettled(batch.map(processItem));

			const succeeded = results.filter((r) => r.status === 'fulfilled' && r.value).length;
			syncState.progress.current = Math.min(i + batch.length, sorted.length);
			syncState.pending = sorted.length - (i + succeeded);
		}

		const remaining = await db.syncQueue
			.where('status')
			.anyOf('pending', 'failed')
			.count();
		syncState.pending = remaining;

		if (remaining === 0) {
			syncState.lastSync = new Date().toISOString();
			syncState.lastError = null;
		}

		const currentStatus: ConnectionStatus = navigator.onLine ? 'online' : 'offline';
		connectionStatus.set(currentStatus);
	} finally {
		isSyncing = false;
		syncState.syncing = false;
	}
}

export function startSyncEngine() {
	if (!syncTimer) {
		syncTimer = setInterval(processSyncQueue, SYNC_INTERVAL);
		processSyncQueue();
	}
}

export function stopSyncEngine() {
	if (syncTimer) {
		clearInterval(syncTimer);
		syncTimer = null;
	}
}

export async function enqueueSync(
	entity: SyncQueueItem['entity'],
	action: SyncQueueItem['action'],
	entityId: string,
	payload: unknown
) {
	await db.syncQueue.add({
		entity,
		action,
		entityId,
		payload,
		created_at: new Date().toISOString(),
		status: 'pending',
		retry_count: 0,
		last_error: null
	});
	syncState.pending++;

	if (navigator.onLine) {
		processSyncQueue();
	}
}

export async function getPendingCount(): Promise<number> {
	return db.syncQueue.where('status').anyOf('pending', 'failed').count();
}
