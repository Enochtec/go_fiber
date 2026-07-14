import { authStore } from '$lib/stores/auth.svelte';

const BASE = '/api';

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
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
	return data;
}

export const api = {
	get: <T>(path: string) => request<T>(path),
	post: <T>(path: string, body: unknown) =>
		request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
	put: <T>(path: string, body: unknown) =>
		request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
	delete: <T>(path: string) => request<T>(path, { method: 'DELETE' })
};
