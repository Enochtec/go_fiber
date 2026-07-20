import { api } from './api';
import type { ApiResponse } from '$lib/types';

export interface STKPushInput {
	phone: string;
	amount: number;
	reference?: string;
}

export interface STKPushResponse {
	checkout_request_id: string;
	customer_message: string;
}

export interface MpesaStatus {
	status: 'pending' | 'completed' | 'failed' | 'cancelled';
	mpesa_receipt: string | null;
	result_code: string | null;
	result_desc: string | null;
	phone: string;
	amount: number;
}

export const mpesaService = {
	initiateSTKPush: (input: STKPushInput) =>
		api.post<ApiResponse<STKPushResponse>>('/mpesa/stk-push', input),

	getStatus: (checkoutRequestID: string) =>
		api.get<ApiResponse<MpesaStatus>>(`/mpesa/status/${checkoutRequestID}`)
};

/**
 * Poll M-Pesa payment status until completed/failed/cancelled or timeout.
 * @param checkoutRequestID  The ID returned by STK push
 * @param onUpdate           Called each poll with current status
 * @param timeoutMs          Max wait time in ms (default 90s)
 * @param intervalMs         Poll interval in ms (default 3s)
 */
export async function pollMpesaStatus(
	checkoutRequestID: string,
	onUpdate: (status: MpesaStatus) => void,
	timeoutMs = 90_000,
	intervalMs = 3_000
): Promise<MpesaStatus> {
	const deadline = Date.now() + timeoutMs;

	while (Date.now() < deadline) {
		await new Promise(r => setTimeout(r, intervalMs));

		try {
			const res = await mpesaService.getStatus(checkoutRequestID);
			const s = res.data!;
			onUpdate(s);

			if (s.status === 'completed' || s.status === 'failed' || s.status === 'cancelled') {
				return s;
			}
		} catch {
			// Network error — keep trying
		}
	}

	// Timeout
	const timeout: MpesaStatus = {
		status: 'failed',
		mpesa_receipt: null,
		result_code: 'TIMEOUT',
		result_desc: 'Payment timed out. Please check your M-Pesa messages.',
		phone: '',
		amount: 0
	};
	onUpdate(timeout);
	return timeout;
}

/** Normalize phone: 07xx → 254 7xx, +254 → 254 */
export function normalizePhone(p: string): string {
	const d = p.replace(/\D/g, '');
	if (d.length === 9) return '254' + d;
	if (d.length === 10 && d[0] === '0') return '254' + d.slice(1);
	if (d.length === 12 && d.startsWith('254')) return d;
	if (d.startsWith('254')) return d;
	return d;
}

/** Validate Kenyan mobile phone */
export function validatePhone(p: string): string | null {
	const d = normalizePhone(p);
	if (!/^254[71]\d{8}$/.test(d)) {
		return 'Enter a valid Kenyan mobile number (e.g. 0712 345678)';
	}
	return null;
}
