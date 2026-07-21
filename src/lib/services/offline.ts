import { db, type OfflineSale, type OfflineSaleItem } from './db';
import { api } from './api';
import { enqueueSync } from './sync';
import { connectionStatus, isOnline } from './connection';
import type { Product, Category, Customer, Supplier } from '$lib/types';
import type { ProductFilter } from './products';
import type { CustomerInput } from './customers';
import type { CreateSaleInput } from './sales';

// ─── Local ID generation ──────────────────────────────────────────
function generateLocalId(): string {
	return 'local_' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 9);
}

function isLocalId(id: string): boolean {
	return id.startsWith('local_');
}

// ─── Helpers ───────────────────────────────────────────────────────
function getConnectionStatus(): 'online' | 'offline' {
	let status: 'online' | 'offline' = 'online';
	connectionStatus.subscribe((v) => {
		if (v === 'online' || v === 'syncing') status = 'online';
		else status = 'offline';
	})();
	return status;
}

// ─── Products ──────────────────────────────────────────────────────
async function cacheProducts() {
	try {
		const res = await api.get<{ success: boolean; data: Product[]; total: number }>('/products?limit=10000');
		if (res?.data) {
			await db.products.clear();
			await db.products.bulkPut(res.data);
		}
	} catch {
		// silently fail
	}
}

async function cacheCategories() {
	try {
		const res = await api.get<{ success: boolean; data: Category[] }>('/categories');
		if (res?.data) {
			await db.categories.clear();
			await db.categories.bulkPut(res.data);
		}
	} catch {
		// silently fail
	}
}

export async function refreshLocalCache() {
	await Promise.all([cacheProducts(), cacheCategories()]);
}

export const offlineProducts = {
	list: async (filter: ProductFilter = {}) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.get<any>('/products' + buildQuery(filter));
				await cacheProducts();
				await cacheCategories();
				return res;
			} catch {
				// fall through to offline
			}
		}

		let collection = db.products.where('is_active').equals(1);

		if (filter.search) {
			const q = filter.search.toLowerCase();
			const all = await db.products.toArray();
			const filtered = all.filter(
				(p) =>
					p.name.toLowerCase().includes(q) ||
					(p.barcode && p.barcode.toLowerCase().includes(q)) ||
					(p.sku && p.sku.toLowerCase().includes(q))
			);
			return {
				success: true,
				data: filtered,
				total: filtered.length,
				page: filter.page || 1,
				limit: filter.limit || filtered.length
			};
		}

		if (filter.category_id) {
			collection = collection.filter((p) => p.category_id === filter.category_id);
		}
		if (filter.low_stock) {
			collection = collection.filter((p) => p.stock_qty <= p.reorder_level);
		}

		const all = await collection.toArray();
		const total = all.length;
		const page = filter.page || 1;
		const limit = filter.limit || total || 20;
		const start = (page - 1) * limit;
		const data = all.slice(start, start + limit);

		return {
			success: true,
			data,
			total,
			page,
			limit
		};
	},

	getById: async (id: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.get<{ success: boolean; data: Product }>(`/products/${id}`);
				if (res?.data) await db.products.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		const product = await db.products.get(id);
		return { success: !!product, data: product };
	},

	getByBarcode: async (barcode: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.get<{ success: boolean; data: Product }>(`/products/barcode/${barcode}`);
				if (res?.data) await db.products.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		const product = await db.products.where('barcode').equals(barcode).first();
		return { success: !!product, data: product };
	},

	create: async (data: any) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.post<{ success: boolean; data: Product }>('/products', data);
				if (res?.data) await db.products.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		const localId = generateLocalId();
		const product: Product = {
			id: localId,
			name: data.name,
			barcode: data.barcode || null,
			sku: data.sku || null,
			category_id: data.category_id || null,
			category_name: null,
			buying_price: data.buying_price || 0,
			selling_price: data.selling_price || 0,
			stock_qty: data.stock_qty || 0,
			reorder_level: data.reorder_level || 0,
			image_url: null,
			is_active: true,
			created_at: new Date().toISOString(),
			updated_at: new Date().toISOString()
		};
		await db.products.put(product);
		await enqueueSync('product', 'create', localId, data);
		return { success: true, data: product };
	},

	update: async (id: string, data: any) => {
		const online = getConnectionStatus() === 'online';
		if (online && !isLocalId(id)) {
			try {
				const res = await api.put<{ success: boolean; data: Product }>(`/products/${id}`, data);
				if (res?.data) await db.products.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		await db.products.update(id, { ...data, updated_at: new Date().toISOString() });
		if (!isLocalId(id)) {
			await enqueueSync('product', 'update', id, data);
		}
		return { success: true, data: await db.products.get(id) };
	},

	delete: async (id: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.delete<{ success: boolean }>(`/products/${id}`);
				await db.products.delete(id);
				return res;
			} catch {
				// fall through
			}
		}
		await db.products.delete(id);
		await enqueueSync('product', 'delete', id, null);
		return { success: true };
	},

	listCategories: async () => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.get<{ success: boolean; data: Category[] }>('/categories');
				if (res?.data) await db.categories.bulkPut(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		const data = await db.categories.toArray();
		return { success: true, data };
	},

	createCategory: async (name: string, description?: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.post<{ success: boolean; data: Category }>('/categories', { name, description });
				if (res?.data) await db.categories.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		const localId = generateLocalId();
		const cat: Category = { id: localId, name, description: description || null, created_at: new Date().toISOString() };
		await db.categories.put(cat);
		await enqueueSync('category', 'create', localId, { name, description });
		return { success: true, data: cat };
	},

	updateCategory: async (id: string, name: string, description?: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.put<{ success: boolean }>(`/categories/${id}`, { name, description });
				await db.categories.update(id, { name, description: description || null });
				return res;
			} catch {
				// fall through
			}
		}
		await db.categories.update(id, { name, description: description || null });
		await enqueueSync('category', 'update', id, { name, description });
		return { success: true };
	},

	deleteCategory: async (id: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.delete<{ success: boolean }>(`/categories/${id}`);
				await db.categories.delete(id);
				return res;
			} catch {
				// fall through
			}
		}
		await db.categories.delete(id);
		await enqueueSync('category', 'delete', id, null);
		return { success: true };
	}
};

// ─── Customers ─────────────────────────────────────────────────────
export const offlineCustomers = {
	list: async (search = '', page = 1, limit = 20) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.get<any>(`/customers?search=${encodeURIComponent(search)}&page=${page}&limit=${limit}`);
				if (res?.data) {
					for (const c of res.data) await db.customers.put(c);
				}
				return res;
			} catch {
				// fall through
			}
		}
		let all = await db.customers.toArray();
		if (search) {
			const q = search.toLowerCase();
			all = all.filter((c) => c.name.toLowerCase().includes(q) || (c.phone && c.phone.includes(q)));
		}
		const total = all.length;
		const start = (page - 1) * limit;
		const data = all.slice(start, start + limit);
		return { success: true, data, total, page, limit };
	},

	getById: async (id: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.get<{ success: boolean; data: Customer }>(`/customers/${id}`);
				if (res?.data) await db.customers.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		const customer = await db.customers.get(id);
		return { success: !!customer, data: customer };
	},

	create: async (data: CustomerInput) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.post<{ success: boolean; data: Customer }>('/customers', data);
				if (res?.data) await db.customers.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		const localId = generateLocalId();
		const customer: Customer = {
			id: localId,
			name: data.name,
			email: data.email || null,
			phone: data.phone || null,
			address: data.address || null,
			created_at: new Date().toISOString()
		};
		await db.customers.put(customer);
		await enqueueSync('customer', 'create', localId, data);
		return { success: true, data: customer };
	},

	update: async (id: string, data: CustomerInput) => {
		const online = getConnectionStatus() === 'online';
		if (online && !isLocalId(id)) {
			try {
				const res = await api.put<{ success: boolean; data: Customer }>(`/customers/${id}`, data);
				if (res?.data) await db.customers.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		await db.customers.update(id, data as any);
		if (!isLocalId(id)) {
			await enqueueSync('customer', 'update', id, data);
		}
		return { success: true, data: await db.customers.get(id) };
	},

	delete: async (id: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.delete<{ success: boolean }>(`/customers/${id}`);
				await db.customers.delete(id);
				return res;
			} catch {
				// fall through
			}
		}
		await db.customers.delete(id);
		await enqueueSync('customer', 'delete', id, null);
		return { success: true };
	}
};

// ─── Sales ─────────────────────────────────────────────────────────
export const offlineSales = {
	list: async (filter: any = {}) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.get<any>('/sales' + buildQuery(filter));
				return res;
			} catch {
				// fall through
			}
		}

		let all = await db.sales.where('status').notEqual('voided').reverse().sortBy('created_at');

		if (filter.status) {
			all = all.filter((s) => s.status === filter.status);
		}
		if (filter.date_from) {
			all = all.filter((s) => s.created_at >= filter.date_from!);
		}
		if (filter.date_to) {
			all = all.filter((s) => s.created_at <= filter.date_to!);
		}

		const page = filter.page || 1;
		const limit = filter.limit || 20;
		const total = all.length;
		const start = (page - 1) * limit;
		const data = all.slice(start, start + limit);

		return { success: true, data, total, page, limit };
	},

	getById: async (id: string) => {
		const online = getConnectionStatus() === 'online';
		if (online && !isLocalId(id)) {
			try {
				return await api.get<{ success: boolean; data: any }>(`/sales/${id}`);
			} catch {
				// fall through
			}
		}
		const sale = await db.sales.get(id);
		return { success: !!sale, data: sale };
	},

	create: async (data: CreateSaleInput) => {
		const online = getConnectionStatus() === 'online';
		const localId = generateLocalId();
		const now = new Date().toISOString();

		const itemsTotal = data.items.reduce((sum, item) => sum + item.unit_price * item.quantity, 0);
		const subtotal = itemsTotal;
		const discount = data.discount || 0;
		const taxRate = data.tax_rate || 0;
		const taxableAmount = subtotal - discount;
		const tax = taxableAmount * taxRate;
		const total = taxableAmount + tax;

		if (online) {
			try {
				const res = await api.post<{ success: boolean; data: any }>('/sales', data);
				if (res?.data) {
					api.invalidate('/dashboard');
					return res;
				}
			} catch {
				// fall through to offline
			}
		}

		const offlineItems: OfflineSaleItem[] = data.items.map((item, idx) => ({
			id: localId + '_item_' + idx,
			product_id: item.product_id,
			product_name: null,
			quantity: item.quantity,
			unit_price: item.unit_price,
			total: item.unit_price * item.quantity
		}));

		const offlineSale: OfflineSale = {
			id: localId,
			local_id: localId,
			cashier_id: '',
			customer_id: data.customer_id || null,
			customer_name: null,
			subtotal,
			discount,
			tax,
			total,
			payment_method: data.payment_method,
			status: 'completed',
			note: data.note || null,
			items: offlineItems,
			created_at: now,
			synced: false,
			synced_at: null,
			server_id: null
		};

		await db.sales.put(offlineSale);
		await enqueueSync('sale', 'create', localId, { ...data, local_id: localId, created_at: now });

		const saleResponse = {
			...offlineSale,
			cashier_name: null,
			id: localId,
			items: offlineItems
		};

		return { success: true, data: saleResponse };
	},

	void: async (id: string) => {
		const online = getConnectionStatus() === 'online';
		if (online && !isLocalId(id)) {
			try {
				const res = await api.put<{ success: boolean }>(`/sales/${id}/void`, {});
				api.invalidate('/dashboard');
				return res;
			} catch {
				// fall through
			}
		}
		await db.sales.update(id, { status: 'voided' });
		if (!isLocalId(id)) {
			await enqueueSync('sale', 'update', id, { status: 'voided' });
		}
		return { success: true };
	}
};

// ─── Suppliers ─────────────────────────────────────────────────────
export const offlineSuppliers = {
	list: async () => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.get<{ success: boolean; data: Supplier[] }>('/suppliers');
				if (res?.data) {
					await db.suppliers.clear();
					await db.suppliers.bulkPut(res.data);
				}
				return res;
			} catch {
				// fall through
			}
		}
		const data = await db.suppliers.toArray();
		return { success: true, data };
	},

	getById: async (id: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.get<{ success: boolean; data: Supplier }>(`/suppliers/${id}`);
				if (res?.data) await db.suppliers.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		const supplier = await db.suppliers.get(id);
		return { success: !!supplier, data: supplier };
	},

	create: async (data: any) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.post<{ success: boolean; data: Supplier }>('/suppliers', data);
				if (res?.data) await db.suppliers.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		const localId = generateLocalId();
		const supplier: Supplier = {
			id: localId,
			name: data.name,
			email: data.email || null,
			phone: data.phone || null,
			address: data.address || null,
			created_at: new Date().toISOString()
		};
		await db.suppliers.put(supplier);
		await enqueueSync('supplier', 'create', localId, data);
		return { success: true, data: supplier };
	},

	update: async (id: string, data: any) => {
		const online = getConnectionStatus() === 'online';
		if (online && !isLocalId(id)) {
			try {
				const res = await api.put<{ success: boolean; data: Supplier }>(`/suppliers/${id}`, data);
				if (res?.data) await db.suppliers.put(res.data);
				return res;
			} catch {
				// fall through
			}
		}
		await db.suppliers.update(id, data as any);
		if (!isLocalId(id)) {
			await enqueueSync('supplier', 'update', id, data);
		}
		return { success: true, data: await db.suppliers.get(id) };
	},

	delete: async (id: string) => {
		const online = getConnectionStatus() === 'online';
		if (online) {
			try {
				const res = await api.delete<{ success: boolean }>(`/suppliers/${id}`);
				await db.suppliers.delete(id);
				return res;
			} catch {
				// fall through
			}
		}
		await db.suppliers.delete(id);
		await enqueueSync('supplier', 'delete', id, null);
		return { success: true };
	}
};

// ─── Query builder ─────────────────────────────────────────────────
function buildQuery(params: Record<string, string | number | boolean | undefined>) {
	const q = new URLSearchParams();
	for (const [k, v] of Object.entries(params)) {
		if (v !== undefined && v !== '' && v !== false) q.set(k, String(v));
	}
	const s = q.toString();
	return s ? '?' + s : '';
}
