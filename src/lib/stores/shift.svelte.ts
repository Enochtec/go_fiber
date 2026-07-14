import { shiftsService, type Shift } from '$lib/services/shifts';

function createShiftStore() {
	let current = $state<Shift | null>(null);
	let loading = $state(true);
	let checked = $state(false);

	return {
		get current() { return current; },
		get loading() { return loading; },
		get isOpen() { return current !== null && current.status === 'open'; },
		get checked() { return checked; },

		async fetch() {
			loading = true;
			try {
				const res = await shiftsService.getCurrent();
				current = res.open && res.data ? res.data : null;
			} catch {
				current = null;
			} finally {
				loading = false;
				checked = true;
			}
		},

		async open(float: number, notes: string) {
			const res = await shiftsService.open({ opening_float: float, notes });
			current = res.data;
			return res.data;
		},

		async close(actualCash: number, notes: string) {
			if (!current) throw new Error('No open shift');
			const closed = await shiftsService.close(current.id, { actual_cash: actualCash, notes });
			current = null;
			return closed.data;
		},

		set(s: Shift | null) { current = s; }
	};
}

export const shiftStore = createShiftStore();
