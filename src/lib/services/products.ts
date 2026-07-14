import { api } from './api';
import type { ApiResponse, Category, PaginatedResponse, Product } from '$lib/types';

export interface ProductInput {
	name: string;
	barcode?: string | null;
	sku?: string | null;
	category_id?: string | null;
	buying_price: number;
	selling_price: number;
	stock_qty: number;
	reorder_level: number;
	image_url?: string | null;
}

export interface ProductFilter {
	search?: string;
	category_id?: string;
	low_stock?: boolean;
	page?: number;
	limit?: number;
}

function buildQuery(params: Record<string, string | number | boolean | undefined>) {
	const q = new URLSearchParams();
	for (const [k, v] of Object.entries(params)) {
		if (v !== undefined && v !== '' && v !== false) q.set(k, String(v));
	}
	const s = q.toString();
	return s ? '?' + s : '';
}

export const productsService = {
	list: (filter: ProductFilter = {}) =>
		api.get<PaginatedResponse<Product>>('/products' + buildQuery(filter as Record<string, string | number | boolean | undefined>)),

	getById: (id: string) => api.get<ApiResponse<Product>>(`/products/${id}`),

	getByBarcode: (barcode: string) => api.get<ApiResponse<Product>>(`/products/barcode/${barcode}`),

	create: (data: ProductInput) => api.post<ApiResponse<Product>>('/products', data),

	update: (id: string, data: ProductInput) => api.put<ApiResponse<Product>>(`/products/${id}`, data),

	delete: (id: string) => api.delete<ApiResponse<null>>(`/products/${id}`),

	listCategories: () => api.get<ApiResponse<Category[]>>('/categories'),

	createCategory: (name: string, description?: string) =>
		api.post<ApiResponse<Category>>('/categories', { name, description }),

	updateCategory: (id: string, name: string, description?: string) =>
		api.put<ApiResponse<null>>(`/categories/${id}`, { name, description }),

	deleteCategory: (id: string) => api.delete<ApiResponse<null>>(`/categories/${id}`)
};
