import { authStore } from '$lib/stores/auth.svelte';

const BASE = import.meta.env.VITE_API_URL || '/api';

interface CacheEntry {
	data: unknown;
	expiry: number;
}

const responseCache = new Map<string, CacheEntry>();
const inflight = new Map<string, Promise<unknown>>();

const CACHE_TTL = 30_000;

function cacheKey(method: string, path: string): string {
	const token = authStore.token;
	const suffix = token ? token.slice(-12) : 'anon';
	return method + ':' + path + ':' + suffix;
}

function getCached(method: string, path: string): unknown | undefined {
	if (method !== 'GET') return;
	const entry = responseCache.get(cacheKey(method, path));
	if (!entry) return;
	if (Date.now() > entry.expiry) {
		responseCache.delete(cacheKey(method, path));
		return;
	}
	return entry.data;
}

function setCache(method: string, path: string, data: unknown): void {
	if (method !== 'GET') return;
	responseCache.set(cacheKey(method, path), { data, expiry: Date.now() + CACHE_TTL });
}

function invalidateCache(prefix: string): void {
	for (const key of responseCache.keys()) {
		if (key.startsWith('GET:' + prefix)) {
			responseCache.delete(key);
		}
	}
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const method = init.method || 'GET';

	if (method === 'GET') {
		const cached = getCached(method, path);
		if (cached) return cached as T;
	}

	const key = cacheKey(method, path);
	if (method === 'GET' && inflight.has(key)) {
		return inflight.get(key) as Promise<T>;
	}

	const promise = doFetch<T>(path, init);
	if (method === 'GET') {
		inflight.set(key, promise);
		promise.finally(() => inflight.delete(key));
	}
	return promise;
}

async function doFetch<T>(path: string, init: RequestInit): Promise<T> {
	const token = authStore.token;
	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(init.headers as Record<string, string>)
	};
	if (token) headers['Authorization'] = `Bearer ${token}`;

	const res = await fetch(BASE + path, { ...init, headers });

	if (res.status === 401) {
		authStore.clear();
		window.location.href = '/login';
		throw new Error('Unauthorized');
	}

	const data = await res.json();
	if (!data.success) throw new Error(data.error || 'Request failed');

	const method = init.method || 'GET';
	setCache(method, path, data);
	return data;
}

async function doFetchForm<T>(path: string, fd: FormData): Promise<T> {
	const token = authStore.token;
	const headers: Record<string, string> = {};
	if (token) headers['Authorization'] = `Bearer ${token}`;

	const res = await fetch(BASE + path, { method: 'POST', headers, body: fd });

	if (res.status === 401) {
		authStore.clear();
		window.location.href = '/login';
		throw new Error('Unauthorized');
	}

	return res.json();
}

export const api = {
	get: <T>(path: string) => request<T>(path),
	post: <T>(path: string, body: unknown) =>
		request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
	put: <T>(path: string, body: unknown) =>
		request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
	delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
	upload: <T>(path: string, fd: FormData) =>
		doFetchForm<T>(path, fd),
	invalidate: (prefix: string) => invalidateCache(prefix)
};
