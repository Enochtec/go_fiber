import { api } from './api';

export async function uploadImage(file: File): Promise<string> {
	const fd = new FormData();
	fd.append('file', file);

	const res = await api.upload<{ url: string }>('/upload', fd);

	if (!res.success || !res.data) {
		throw new Error(res.error ?? 'Upload failed');
	}

	return res.data.url;
}
