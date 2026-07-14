import { api } from './api';

export interface Shift {
	id: string;
	cashier_id: string;
	cashier_name: string;
	opening_float: number;
	opening_time: string;
	closing_time?: string;
	cash_sales: number;
	mpesa_sales: number;
	card_sales: number;
	other_sales: number;
	total_sales: number;
	transaction_count: number;
	closing_float?: number;
	expected_cash?: number;
	actual_cash?: number;
	variance?: number;
	status: 'open' | 'closed';
	notes: string;
	created_at: string;
}

interface ShiftCurrentResponse { data: Shift | null; open: boolean }
interface ShiftResponse { data: Shift; success: boolean }
interface ShiftListResponse { data: Shift[]; success: boolean }

export const shiftsService = {
	getCurrent: () =>
		api.get<ShiftCurrentResponse>('/shifts/current'),

	open: (data: { opening_float: number; notes?: string }) =>
		api.post<ShiftResponse>('/shifts/open', data),

	close: (id: string, data: { actual_cash: number; notes?: string }) =>
		api.post<ShiftResponse>(`/shifts/${id}/close`, data),

	list: () =>
		api.get<ShiftListResponse>('/shifts'),
};
