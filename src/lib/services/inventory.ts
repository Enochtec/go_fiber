import { api } from './api';
import type { ApiResponse, DashboardStats, PaginatedResponse, StockAdjustment } from '$lib/types';
import { offlineInventory } from './offline';

export interface AdjustInput {
	product_id: string;
	quantity: number;
	reason: string;
}

export const inventoryService = {
	dashboard: async () => {
		try { return await api.get<ApiResponse<DashboardStats>>('/dashboard'); }
		catch { return offlineInventory.dashboard(); }
	},

	listAdjustments: async (page = 1, limit = 20) => {
		try { return await api.get<PaginatedResponse<StockAdjustment>>(`/inventory/adjustments?page=${page}&limit=${limit}`); }
		catch { return offlineInventory.listAdjustments(page, limit); }
	},

	adjust: async (data: AdjustInput) => {
		try { return await api.post<ApiResponse<StockAdjustment>>('/inventory/adjust', data); }
		catch { return offlineInventory.adjust(data); }
	}
};
