import { api } from './api';
import type { ApiResponse, Customer, PaginatedResponse } from '$lib/types';

export interface CustomerInput {
	name: string;
	email?: string | null;
	phone?: string | null;
	address?: string | null;
}

export const customersService = {
	list: (search = '', page = 1, limit = 20) =>
		api.get<PaginatedResponse<Customer>>(
			`/customers?search=${encodeURIComponent(search)}&page=${page}&limit=${limit}`
		),

	getById: (id: string) => api.get<ApiResponse<Customer>>(`/customers/${id}`),

	create: (data: CustomerInput) => api.post<ApiResponse<Customer>>('/customers', data),

	update: (id: string, data: CustomerInput) =>
		api.put<ApiResponse<Customer>>(`/customers/${id}`, data),

	delete: (id: string) => api.delete<ApiResponse<null>>(`/customers/${id}`)
};
