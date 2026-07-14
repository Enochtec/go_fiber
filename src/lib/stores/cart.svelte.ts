import type { CartItem, Product } from '$lib/types';

function createCartStore() {
	let items = $state<CartItem[]>([]);
	let discount = $state(0);
	let taxRate = $state(0);
	let customerId = $state<string | null>(null);
	let customerName = $state<string | null>(null);
	let note = $state('');

	return {
		get items() { return items; },
		get discount() { return discount; },
		get taxRate() { return taxRate; },
		get customerId() { return customerId; },
		get customerName() { return customerName; },
		get note() { return note; },

		get subtotal() {
			return items.reduce((sum, i) => sum + i.unit_price * i.quantity, 0);
		},
		get taxAmount() {
			const sub = items.reduce((sum, i) => sum + i.unit_price * i.quantity, 0);
			return sub * (taxRate / 100);
		},
		get total() {
			const sub = items.reduce((sum, i) => sum + i.unit_price * i.quantity, 0);
			const tax = sub * (taxRate / 100);
			return sub - discount + tax;
		},
		get count() {
			return items.reduce((sum, i) => sum + i.quantity, 0);
		},

		addProduct(product: Product) {
			const existing = items.find((i) => i.product.id === product.id);
			if (existing) {
				existing.quantity += 1;
			} else {
				items.push({ product, quantity: 1, unit_price: product.selling_price });
			}
		},

		removeItem(productId: string) {
			const idx = items.findIndex((i) => i.product.id === productId);
			if (idx !== -1) items.splice(idx, 1);
		},

		updateQuantity(productId: string, qty: number) {
			if (qty <= 0) {
				const idx = items.findIndex((i) => i.product.id === productId);
				if (idx !== -1) items.splice(idx, 1);
				return;
			}
			const item = items.find((i) => i.product.id === productId);
			if (item) item.quantity = qty;
		},

		updatePrice(productId: string, price: number) {
			const item = items.find((i) => i.product.id === productId);
			if (item) item.unit_price = price;
		},

		setDiscount(v: number) { discount = v; },
		setTaxRate(v: number) { taxRate = v; },
		setCustomer(id: string | null, name: string | null) {
			customerId = id;
			customerName = name;
		},
		setNote(v: string) { note = v; },

		clear() {
			items = [];
			discount = 0;
			taxRate = 0;
			customerId = null;
			customerName = null;
			note = '';
		}
	};
}

export const cart = createCartStore();
