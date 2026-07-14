export type Role = 'admin' | 'manager' | 'cashier';

export interface User {
	id: string;
	name: string;
	email: string;
	role: Role;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface Category {
	id: string;
	name: string;
	description: string | null;
	created_at: string;
}

export interface Product {
	id: string;
	name: string;
	barcode: string | null;
	sku: string | null;
	category_id: string | null;
	category_name: string | null;
	buying_price: number;
	selling_price: number;
	stock_qty: number;
	reorder_level: number;
	image_url: string | null;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface Customer {
	id: string;
	name: string;
	email: string | null;
	phone: string | null;
	address: string | null;
	created_at: string;
}

export interface Supplier {
	id: string;
	name: string;
	email: string | null;
	phone: string | null;
	address: string | null;
	created_at: string;
}

export type PaymentMethod = 'cash' | 'mpesa' | 'bank' | 'card' | 'credit';
export type SaleStatus = 'completed' | 'held' | 'voided';

export interface SaleItem {
	id: string;
	sale_id: string;
	product_id: string;
	product_name: string | null;
	quantity: number;
	unit_price: number;
	total: number;
}

export interface Sale {
	id: string;
	cashier_id: string;
	cashier_name: string | null;
	customer_id: string | null;
	customer_name: string | null;
	subtotal: number;
	discount: number;
	tax: number;
	total: number;
	payment_method: PaymentMethod;
	status: SaleStatus;
	note: string | null;
	created_at: string;
	items?: SaleItem[];
}

export type PurchaseStatus = 'received' | 'pending';

export interface PurchaseItem {
	id: string;
	purchase_id: string;
	product_id: string;
	product_name: string | null;
	quantity: number;
	unit_price: number;
	total: number;
}

export interface Purchase {
	id: string;
	supplier_id: string | null;
	supplier_name: string | null;
	user_id: string;
	user_name: string | null;
	total: number;
	status: PurchaseStatus;
	note: string | null;
	created_at: string;
	items?: PurchaseItem[];
}

export interface StockAdjustment {
	id: string;
	product_id: string;
	product_name: string | null;
	user_id: string;
	user_name: string | null;
	quantity: number;
	reason: string;
	created_at: string;
}

export interface DashboardStats {
	today_sales: number;
	today_orders: number;
	total_products: number;
	low_stock_count: number;
	total_customers: number;
	month_sales: number;
}

export interface CartItem {
	product: Product;
	quantity: number;
	unit_price: number;
}

export interface ApiResponse<T> {
	success: boolean;
	data?: T;
	message?: string;
	error?: string;
}

export interface PaginatedResponse<T> {
	success: boolean;
	data: T[];
	total: number;
	page: number;
	limit: number;
}

export interface DailySalesRow {
	date: string;
	orders: number;
	total: number;
}

export interface TopProductRow {
	product_id: string;
	product_name: string;
	quantity: number;
	revenue: number;
}

export interface InventoryValueRow {
	category_name: string;
	product_count: number;
	total_cost: number;
	total_value: number;
}
