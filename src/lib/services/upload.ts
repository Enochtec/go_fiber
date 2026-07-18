import { authStore } from '$lib/stores/auth.svelte';

const BASE = import.meta.env.VITE_API_URL || '/api';

export async function uploadImage(file: File): Promise<string> {
	const form = new FormData();
	form.append('file', file);

	const token = authStore.token;
	const headers: Record<string, string> = {};
	if (token) headers['Authorization'] = `Bearer ${token}`;

	const res = await fetch(BASE + '/upload', { method: 'POST', headers, body: form });

	if (res.status === 401) {
		authStore.clear();
		window.location.href = '/login';
		throw new Error('Unauthorized');
	}

	const data = await res.json();
	if (!data.success) throw new Error(data.error || 'Upload failed');
	return data.data.url;
}
