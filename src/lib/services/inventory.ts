import { api } from './api';
import type { ApiResponse, DashboardStats, PaginatedResponse, StockAdjustment } from '$lib/types';

export interface AdjustInput {
	product_id: string;
	quantity: number;
	reason: string;
}

export const inventoryService = {
	dashboard: () => api.get<ApiResponse<DashboardStats>>('/dashboard'),

	listAdjustments: (page = 1, limit = 20) =>
		api.get<PaginatedResponse<StockAdjustment>>(
			`/inventory/adjustments?page=${page}&limit=${limit}`
		),

	adjust: (data: AdjustInput) =>
		api.post<ApiResponse<StockAdjustment>>('/inventory/adjust', data)
};
