import { api } from './api';
import type { ApiResponse } from '$lib/types';

export interface SendReceiptInput {
	sale_id: string;
	email: string;
	tendered?: number;
	change?: number;
}

export const emailService = {
	sendReceipt: (input: SendReceiptInput) =>
		api.post<ApiResponse<null>>('/email/receipt', input)
};

/** Basic email validation */
export function validateEmail(email: string): string | null {
	const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
	if (!re.test(email.trim())) return 'Enter a valid email address';
	return null;
}
