export type NotifType = 'success' | 'error' | 'info';

interface Notification {
	id: number;
	type: NotifType;
	message: string;
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
		error: (msg: string) => add('error', msg, 5000),
		info: (msg: string) => add('info', msg),
		dismiss: (id: number) => {
			const idx = items.findIndex((n) => n.id === id);
			if (idx !== -1) items.splice(idx, 1);
		}
	};
}

export const notify = createNotificationStore();
