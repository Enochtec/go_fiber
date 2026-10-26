import { api } from './api';
import type { ApiResponse, PaginatedResponse, Purchase, PurchaseStatus } from '$lib/types';
import { offlinePurchases } from './offline';

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
	list: async (page = 1, limit = 20) => {
		try { return await api.get<PaginatedResponse<Purchase>>(`/purchases?page=${page}&limit=${limit}`); }
		catch { return offlinePurchases.list(page, limit); }
	},

	getById: async (id: string) => {
		try { return await api.get<ApiResponse<Purchase>>(`/purchases/${id}`); }
		catch {
			const result = await offlinePurchases.list(1, 10000);
			return { success: !!result.data.find((purchase) => purchase.id === id), data: result.data.find((purchase) => purchase.id === id) };
		}
	},

	create: async (data: CreatePurchaseInput) => {
		try { return await api.post<ApiResponse<Purchase>>('/purchases', data); }
		catch { return offlinePurchases.create(data); }
	}
};
