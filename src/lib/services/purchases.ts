import { api } from './api';
import type { ApiResponse, PaginatedResponse, Purchase, PurchaseStatus } from '$lib/types';

export interface PurchaseItemInput {
	product_id: string;
	quantity: number;
	unit_price: number;
}

export interface CreatePurchaseInput {
	supplier_id?: string | null;
	items: PurchaseItemInput[];
	status?: PurchaseStatus;
	note?: string | null;
}

export const purchasesService = {
	list: (page = 1, limit = 20) =>
		api.get<PaginatedResponse<Purchase>>(`/purchases?page=${page}&limit=${limit}`),

	getById: (id: string) => api.get<ApiResponse<Purchase>>(`/purchases/${id}`),

	create: (data: CreatePurchaseInput) => api.post<ApiResponse<Purchase>>('/purchases', data)
};
