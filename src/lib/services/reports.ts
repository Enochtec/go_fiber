import { api } from './api';
import type { ApiResponse, DailySalesRow, InventoryValueRow, TopProductRow } from '$lib/types';

export const reportsService = {
	dailySales: (days = 30) =>
		api.get<ApiResponse<DailySalesRow[]>>(`/reports/sales/daily?days=${days}`),

	monthlySales: (months = 12) =>
		api.get<ApiResponse<DailySalesRow[]>>(`/reports/sales/monthly?months=${months}`),

	topProducts: (limit = 10) =>
		api.get<ApiResponse<TopProductRow[]>>(`/reports/products/top?limit=${limit}`),

	inventoryValue: () =>
		api.get<ApiResponse<InventoryValueRow[]>>('/reports/inventory/value')
};
