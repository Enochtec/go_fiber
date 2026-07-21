import { writable, derived } from 'svelte/store';

export type ConnectionStatus = 'online' | 'offline' | 'syncing';

export const connectionStatus = writable<ConnectionStatus>(
	navigator.onLine ? 'online' : 'offline'
);

export const isOnline = derived(connectionStatus, ($s) => $s === 'online' || $s === 'syncing');

let healthCheckInterval: ReturnType<typeof setInterval> | null = null;
const HEALTH_CHECK_URL = import.meta.env.VITE_API_URL || '/api';
const HEALTH_CHECK_PATH = '/health';
const CHECK_INTERVAL = 15_000;

async function checkHealth(): Promise<boolean> {
	try {
		const controller = new AbortController();
		const id = setTimeout(() => controller.abort(), 5000);
		const res = await fetch(HEALTH_CHECK_URL + HEALTH_CHECK_PATH, {
			method: 'HEAD',
			signal: controller.signal,
			cache: 'no-store'
		});
		clearTimeout(id);
		return res.ok;
	} catch {
		return false;
	}
}

export function startConnectionMonitor() {
	window.addEventListener('online', () => {
		connectionStatus.set('online');
		runHealthCheck();
	});

	window.addEventListener('offline', () => {
		connectionStatus.set('offline');
	});

	if (healthCheckInterval) clearInterval(healthCheckInterval);
	healthCheckInterval = setInterval(runHealthCheck, CHECK_INTERVAL);

	runHealthCheck();
}

export function stopConnectionMonitor() {
	if (healthCheckInterval) {
		clearInterval(healthCheckInterval);
		healthCheckInterval = null;
	}
}

async function runHealthCheck() {
	const current: ConnectionStatus = 'online';
	let latest = 'online';
	connectionStatus.subscribe((v) => (latest = v))();
	if (latest === 'syncing') return;

	const ok = await checkHealth();
	if (!ok && latest !== 'offline') {
		connectionStatus.set('offline');
	} else if (ok && latest === 'offline') {
		connectionStatus.set('online');
	}
}
