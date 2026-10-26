import { api } from './api';
import type { ApiResponse, DailySalesRow, InventoryValueRow, TopProductRow } from '$lib/types';
import { offlineReports } from './offline';

export const reportsService = {
	dailySales: async (days = 30) => {
		try { return await api.get<ApiResponse<DailySalesRow[]>>(`/reports/sales/daily?days=${days}`); }
		catch { return offlineReports.dailySales(days); }
	},

	monthlySales: async (months = 12) => {
		try { return await api.get<ApiResponse<DailySalesRow[]>>(`/reports/sales/monthly?months=${months}`); }
		catch { return offlineReports.monthlySales(months); }
	},

	topProducts: async (limit = 10) => {
		try { return await api.get<ApiResponse<TopProductRow[]>>(`/reports/products/top?limit=${limit}`); }
		catch { return offlineReports.topProducts(limit); }
	},

	inventoryValue: async () => {
		try { return await api.get<ApiResponse<InventoryValueRow[]>>('/reports/inventory/value'); }
		catch { return offlineReports.inventoryValue(); }
	}
};
