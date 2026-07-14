import { api } from './api';
import type { ApiResponse, Supplier } from '$lib/types';

export interface SupplierInput {
	name: string;
	email?: string | null;
	phone?: string | null;
	address?: string | null;
}

export const suppliersService = {
	list: () => api.get<ApiResponse<Supplier[]>>('/suppliers'),

	getById: (id: string) => api.get<ApiResponse<Supplier>>(`/suppliers/${id}`),

	create: (data: SupplierInput) => api.post<ApiResponse<Supplier>>('/suppliers', data),

	update: (id: string, data: SupplierInput) =>
		api.put<ApiResponse<Supplier>>(`/suppliers/${id}`, data),

	delete: (id: string) => api.delete<ApiResponse<null>>(`/suppliers/${id}`)
};
