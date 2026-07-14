import { api } from './api';
import type { ApiResponse, PaginatedResponse, PaymentMethod, Sale, SaleStatus } from '$lib/types';

export interface SaleItemInput {
	product_id: string;
	quantity: number;
	unit_price: number;
}

export interface CreateSaleInput {
	customer_id?: string | null;
	items: SaleItemInput[];
	discount: number;
	tax_rate: number;
	payment_method: PaymentMethod;
	status?: SaleStatus;
	note?: string | null;
}

export interface SaleFilter {
	status?: SaleStatus;
	cashier_id?: string;
	date_from?: string;
	date_to?: string;
	page?: number;
	limit?: number;
}

function buildQuery(params: Record<string, string | number | boolean | undefined>) {
	const q = new URLSearchParams();
	for (const [k, v] of Object.entries(params)) {
		if (v !== undefined && v !== '') q.set(k, String(v));
	}
	const s = q.toString();
	return s ? '?' + s : '';
}

export const salesService = {
	list: (filter: SaleFilter = {}) =>
		api.get<PaginatedResponse<Sale>>('/sales' + buildQuery(filter as Record<string, string | number | boolean | undefined>)),

	getById: (id: string) => api.get<ApiResponse<Sale>>(`/sales/${id}`),

	create: (data: CreateSaleInput) => api.post<ApiResponse<Sale>>('/sales', data),

	void: (id: string) => api.put<ApiResponse<null>>(`/sales/${id}/void`, {}),

	hold: (id: string) => api.put<ApiResponse<null>>(`/sales/${id}/hold`, {})
};
