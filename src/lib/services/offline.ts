import { db, type OfflineSale, type OfflineSaleItem } from './db';
import { api } from './api';
import { enqueueSync } from './sync.svelte';
import { connectionStatus, isOnline } from './connection';
import type { Product, Category, Customer, Supplier, DashboardStats, DailySalesRow, InventoryValueRow, TopProductRow, Purchase, StockAdjustment } from '$lib/types';
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
				const res = await api.get<any>('/products' + buildQuery(filter as Record<string, string | number | boolean | undefined>));
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
			} catch (err) {
				if (err instanceof Error && err.message.toLowerCase().includes('already exists')) throw err;
				// fall through
			}
		}
		const duplicate = await db.products.filter((product) => product.is_active && product.name.trim().toLowerCase() === String(data.name).trim().toLowerCase()).count();
		if (duplicate > 0) throw new Error('A product with this name already exists');
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
			} catch (err) {
				if (err instanceof Error && err.message.toLowerCase().includes('already exists')) throw err;
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
			} catch (err) {
				if (err instanceof Error && err.message.toLowerCase().includes('already exists')) throw err;
				// fall through
			}
		}
		const duplicate = await db.categories.filter((category) => category.name.trim().toLowerCase() === name.trim().toLowerCase()).count();
		if (duplicate > 0) throw new Error('A category with this name already exists');
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
			} catch (err) {
				if (err instanceof Error && err.message.toLowerCase().includes('already exists')) throw err;
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
	},

	stats: async (id: string) => {
		const sales = (await localSales()).filter((sale) => sale.customer_id === id);
		const lifetime = sales.reduce((sum, sale) => sum + sale.total, 0);
		return { success: true, data: { total_orders: sales.length, lifetime_spend: lifetime, avg_order: sales.length ? lifetime / sales.length : 0, last_visit: sales.length ? sales[sales.length - 1].created_at : null } };
	},

	history: async (id: string, limit = 20) => {
		const sales = (await localSales()).filter((sale) => sale.customer_id === id).slice(-limit).reverse();
		return { success: true, data: sales };
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

// ─── Inventory, purchases, reports and dashboard ─────────────────
		function localDate(date = new Date()): string {
			return date.toISOString().slice(0, 10);
		}

		async function localSales(): Promise<OfflineSale[]> {
			return db.sales.where('status').notEqual('voided').toArray();
		}

		export const offlineInventory = {
			dashboard: async () => {
				const products = await db.products.toArray();
				const customers = await db.customers.count();
				const sales = await localSales();
				const today = localDate();
				const month = today.slice(0, 7);
				const todaySales = sales.filter((s) => s.created_at.startsWith(today));
				const monthSales = sales.filter((s) => s.created_at.startsWith(month));
				const total = todaySales.reduce((sum, s) => sum + s.total, 0);
				const stats: DashboardStats = {
					today_sales: total,
					today_orders: todaySales.length,
					today_avg_sale: todaySales.length ? total / todaySales.length : 0,
					today_cash_sales: todaySales.filter((s) => s.payment_method === 'cash').reduce((sum, s) => sum + s.total, 0),
					today_mpesa: todaySales.filter((s) => s.payment_method === 'mpesa').reduce((sum, s) => sum + s.total, 0),
					today_card: todaySales.filter((s) => s.payment_method === 'card').reduce((sum, s) => sum + s.total, 0),
					total_products: products.length,
					low_stock_count: products.filter((p) => p.stock_qty > 0 && p.stock_qty <= p.reorder_level).length,
					out_of_stock: products.filter((p) => p.stock_qty <= 0).length,
					total_customers: customers,
					month_sales: monthSales.reduce((sum, s) => sum + s.total, 0),
					yesterday_sales: sales.filter((s) => s.created_at.startsWith(localDate(new Date(Date.now() - 86400000)))).reduce((sum, s) => sum + s.total, 0)
				};
				return { success: true, data: stats };
			},

			listAdjustments: async (page = 1, limit = 20) => {
				const all = (await db.adjustments.toArray()).sort((a, b) => b.created_at.localeCompare(a.created_at));
				const start = (page - 1) * limit;
				return { success: true, data: all.slice(start, start + limit), total: all.length, page, limit };
			},

			adjust: async (data: { product_id: string; quantity: number; reason: string }) => {
				const product = await db.products.get(data.product_id);
				if (!product) throw new Error('The selected product is not available offline');
				const nextStock = product.stock_qty + data.quantity;
				if (nextStock < 0) throw new Error('Stock cannot go below zero');
				await db.products.update(product.id, { stock_qty: nextStock, updated_at: new Date().toISOString() });
				const id = generateLocalId();
				const adjustment: StockAdjustment = {
					id, product_id: product.id, product_name: product.name, user_id: '', user_name: null,
					quantity: data.quantity, reason: data.reason, created_at: new Date().toISOString()
				};
				await db.adjustments.put(adjustment);
				await enqueueSync('adjustment', 'create', id, data);
				return { success: true, data: adjustment };
			}
		};

		export const offlinePurchases = {
			list: async (page = 1, limit = 20) => {
				const all = (await db.purchases.toArray()).sort((a, b) => b.created_at.localeCompare(a.created_at));
				const start = (page - 1) * limit;
				return { success: true, data: all.slice(start, start + limit), total: all.length, page, limit };
			},
			create: async (data: { supplier_id?: string | null; items: { product_id: string; quantity: number; unit_price: number }[]; status?: 'received' | 'pending'; note?: string | null }) => {
				const id = generateLocalId();
				const purchaseItems = [];
				for (const item of data.items) {
					const product = await db.products.get(item.product_id);
					if (!product) throw new Error('One of the selected products is not available offline');
					purchaseItems.push({ id: `${id}_${item.product_id}`, purchase_id: id, product_id: item.product_id, product_name: product.name, quantity: item.quantity, unit_price: item.unit_price, total: item.quantity * item.unit_price });
					if (data.status !== 'pending') await db.products.update(item.product_id, { stock_qty: product.stock_qty + item.quantity, updated_at: new Date().toISOString() });
				}
				const purchase: Purchase = {
					id, supplier_id: data.supplier_id || null, supplier_name: data.supplier_id ? (await db.suppliers.get(data.supplier_id))?.name ?? null : null,
					user_id: '', user_name: null, total: purchaseItems.reduce((sum, item) => sum + item.total, 0),
					status: data.status || 'received', note: data.note || null, created_at: new Date().toISOString(), items: purchaseItems
				};
				await db.purchases.put(purchase);
				await enqueueSync('purchase', 'create', id, data);
				return { success: true, data: purchase };
			}
		};

		export const offlineReports = {
			dailySales: async (days = 30) => {
				const sales = await localSales();
				const rows: DailySalesRow[] = [];
				for (let i = days - 1; i >= 0; i--) {
					const date = localDate(new Date(Date.now() - i * 86400000));
					const day = sales.filter((s) => s.created_at.startsWith(date));
					rows.push({ date, orders: day.length, total: day.reduce((sum, s) => sum + s.total, 0) });
				}
				return { success: true, data: rows };
			},
			monthlySales: async (months = 12) => {
				const sales = await localSales();
				const rows: DailySalesRow[] = [];
				for (let i = months - 1; i >= 0; i--) {
					const date = new Date();
					date.setMonth(date.getMonth() - i);
					const month = date.toISOString().slice(0, 7);
					const monthSales = sales.filter((s) => s.created_at.startsWith(month));
					rows.push({ date: month, orders: monthSales.length, total: monthSales.reduce((sum, s) => sum + s.total, 0) });
				}
				return { success: true, data: rows };
			},
			topProducts: async (limit = 10) => {
				const sales = await localSales();
				const totals = new Map<string, TopProductRow>();
				for (const sale of sales) for (const item of sale.items) {
					const current = totals.get(item.product_id) ?? { product_id: item.product_id, product_name: item.product_name || item.product_id, quantity: 0, quantity_sold: 0, revenue: 0 };
					current.quantity_sold += item.quantity;
					current.quantity += item.quantity;
					current.revenue += item.total;
					totals.set(item.product_id, current);
				}
				return { success: true, data: [...totals.values()].sort((a, b) => b.revenue - a.revenue).slice(0, limit) };
			},
			inventoryValue: async () => {
				const products = await db.products.toArray();
				const groups = new Map<string, InventoryValueRow>();
				for (const product of products) {
					const category = product.category_name || 'Uncategorized';
					const row = groups.get(category) ?? { category_name: category, product_count: 0, total_cost: 0, total_value: 0 };
					row.product_count++;
					row.total_cost += product.buying_price * product.stock_qty;
					row.total_value += product.selling_price * product.stock_qty;
					groups.set(category, row);
				}
				return { success: true, data: [...groups.values()] };
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
