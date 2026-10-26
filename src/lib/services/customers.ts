import { api } from './api';
import type { ApiResponse, Customer, PaginatedResponse, Sale } from '$lib/types';
import { offlineCustomers } from './offline';

export interface CustomerInput {
	name: string;
	email?: string | null;
	phone?: string | null;
	address?: string | null;
}

export interface CustomerStats {
	total_orders: number;
	lifetime_spend: number;
	avg_order: number;
	last_visit: string | null;
}

export const customersService = {
	list: (search = '', page = 1, limit = 20) =>
		api.get<PaginatedResponse<Customer>>(
			`/customers?search=${encodeURIComponent(search)}&page=${page}&limit=${limit}`
		),

	getById: (id: string) => api.get<ApiResponse<Customer>>(`/customers/${id}`),

	stats: async (id: string) => {
		try { return await api.get<ApiResponse<CustomerStats>>(`/customers/${id}/stats`); }
		catch { return offlineCustomers.stats(id); }
	},

	history: async (id: string, limit = 20) => {
		try { return await api.get<ApiResponse<Sale[]>>(`/customers/${id}/history?limit=${limit}`); }
		catch { return offlineCustomers.history(id, limit); }
	},

	create: (data: CustomerInput) => api.post<ApiResponse<Customer>>('/customers', data),

	update: (id: string, data: CustomerInput) =>
		api.put<ApiResponse<Customer>>(`/customers/${id}`, data),

	delete: (id: string) => api.delete<ApiResponse<null>>(`/customers/${id}`)
};
