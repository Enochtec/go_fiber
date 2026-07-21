import Dexie, { type Table } from 'dexie';
import type { Product, Category, Customer, Supplier } from '$lib/types';

export interface OfflineSale {
	id: string;
	local_id: string;
	cashier_id: string;
	customer_id: string | null;
	customer_name: string | null;
	subtotal: number;
	discount: number;
	tax: number;
	total: number;
	payment_method: string;
	status: string;
	note: string | null;
	items: OfflineSaleItem[];
	created_at: string;
	synced: boolean;
	synced_at: string | null;
	server_id: string | null;
}

export interface OfflineSaleItem {
	id: string;
	product_id: string;
	product_name: string | null;
	quantity: number;
	unit_price: number;
	total: number;
}

export interface SyncQueueItem {
	id?: number;
	entity: 'sale' | 'customer' | 'product' | 'supplier' | 'category';
	action: 'create' | 'update' | 'delete';
	entityId: string;
	payload: unknown;
	created_at: string;
	status: 'pending' | 'syncing' | 'failed';
	retry_count: number;
	last_error: string | null;
}

export class MaestroDB extends Dexie {
	products!: Table<Product, string>;
	categories!: Table<Category, string>;
	customers!: Table<Customer, string>;
	suppliers!: Table<Supplier, string>;
	sales!: Table<OfflineSale, string>;
	syncQueue!: Table<SyncQueueItem, number>;
	meta!: Table<{ key: string; value: string }, string>;

	constructor() {
		super('MaestroPOS');
		this.version(1).stores({
			products: 'id, name, barcode, category_id, is_active, updated_at',
			categories: 'id, name',
			customers: 'id, name, phone, email',
			suppliers: 'id, name, phone',
			sales: 'id, created_at, synced, status',
			syncQueue: '++id, entity, action, status, created_at',
			meta: 'key'
		});
	}
}

export const db = new MaestroDB();
