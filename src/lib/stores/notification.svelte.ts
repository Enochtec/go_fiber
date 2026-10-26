export type NotifType = 'success' | 'error' | 'info';

interface Notification {
	id: number;
	type: NotifType;
	message: string;
}

export function toUserMessage(error: unknown, fallback = 'Something went wrong. Please try again.') {
	const message = error instanceof Error ? error.message : typeof error === 'string' ? error : '';
	const normalized = message.toLowerCase();
	if (!message || normalized.includes('failed to fetch') || normalized.includes('networkerror') || normalized.includes('load failed')) {
		return 'We could not connect to the server. Your work is saved on this device and will sync when you are back online.';
	}
	if (normalized.includes('unauthorized') || normalized.includes('401')) return 'Your session has expired. Please sign in again.';
	if (normalized.includes('forbidden') || normalized.includes('403')) return 'You do not have permission to perform this action.';
	if (normalized.includes('not found') || normalized.includes('404')) return 'The requested information could not be found.';
	if (normalized.includes('timeout') || normalized.includes('aborted')) return 'The request took too long. Please try again.';
	if (normalized.includes('quota') || normalized.includes('indexeddb') || normalized.includes('transaction')) return 'Local storage is unavailable. Please check your browser storage settings.';
	if (normalized.includes('already exists') || normalized.includes('required') || normalized.includes('cannot') || normalized.includes('invalid')) return message;
	return fallback;
}

function createNotificationStore() {
	let items = $state<Notification[]>([]);
	let next = 0;

	function add(type: NotifType, message: string, duration = 3500) {
		const id = ++next;
		items.push({ id, type, message });
		setTimeout(() => {
			const idx = items.findIndex((n) => n.id === id);
			if (idx !== -1) items.splice(idx, 1);
		}, duration);
	}

	return {
		get items() { return items; },
		success: (msg: string) => add('success', msg),
		error: (msg: string) => add('error', toUserMessage(msg), 5000),
		info: (msg: string) => add('info', msg),
		dismiss: (id: number) => {
			const idx = items.findIndex((n) => n.id === id);
			if (idx !== -1) items.splice(idx, 1);
		}
	};
}

export const notify = createNotificationStore();
