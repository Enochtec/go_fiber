import ExcelJS from 'exceljs';
import type { Product, Sale, Customer, Supplier, Purchase } from '$lib/types';

const BRAND     = '3F00FF';
const HEADER_FG = 'FFFFFF';
const ALT_ROW   = 'F2F4F7';
const BORDER    = 'D0D5DD';
const META_CLR  = '667085';

export function formatCurrency(n: number | null | undefined): string {
	if (n === null || n === undefined) return '0.00';
	return new Intl.NumberFormat('en-KE', {
		minimumFractionDigits: 2, maximumFractionDigits: 2
	}).format(n);
}

export function formatDate(d: string | null | undefined): string {
	if (!d) return '';
	return new Date(d).toLocaleDateString('en-GB', {
		day: '2-digit', month: 'short', year: 'numeric'
	});
}

export function formatPercent(n: number): string {
	return n.toFixed(1) + '%';
}

export function safeFilename(shopName: string, module: string, ext = 'xlsx'): string {
	const date = new Date().toISOString().slice(0, 10);
	const safe = (shopName || 'Export').replace(/[^a-zA-Z0-9]/g, '_').replace(/_+/g, '_');
	return `${safe}_${module}_${date}.${ext}`;
}

async function buildWorkbook(
	sheetName: string,
	columns: { header: string; key: string; width?: number; numFmt?: string }[],
	data: Record<string, unknown>[],
	shopName: string,
	exportedBy: string
): Promise<ArrayBuffer> {
	const wb = new ExcelJS.Workbook();
	wb.creator = exportedBy;
	wb.created = new Date();

	const safeSheet = sheetName.replace(/[\[\]:*?\/\\]/g, '_').slice(0, 31);
	const ws = wb.addWorksheet(safeSheet);

	ws.columns = columns.map(c => ({
		header: c.header,
		key: c.key,
		width: c.width ?? 22
	}));

	const colCount = columns.length;

	// ── Title row ────────────────────────────────────
	ws.mergeCells(1, 1, 1, colCount);
	const titleCell = ws.getCell('A1');
	titleCell.value = sheetName;
	titleCell.font = { name: 'Calibri', size: 18, bold: true, color: { argb: HEADER_FG } };
	titleCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: BRAND } };
	titleCell.alignment = { vertical: 'middle', horizontal: 'left' };
	ws.getRow(1).height = 40;

	// ── Meta row ─────────────────────────────────────
	const now = new Date();
	const dateStr = now.toLocaleDateString('en-GB', { day: '2-digit', month: 'long', year: 'numeric' });
	const timeStr = now.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });
	ws.mergeCells(2, 1, 2, colCount);
	const metaCell = ws.getCell('A2');
	metaCell.value = `Exported: ${dateStr} at ${timeStr}  |  Shop: ${shopName}  |  Exported by: ${exportedBy}`;
	metaCell.font = { name: 'Calibri', size: 10, color: { argb: META_CLR }, italic: true };
	metaCell.alignment = { vertical: 'middle' };
	ws.getRow(2).height = 24;

	// ── Header row (row 3) ──────────────────────────
	const hdrRow = ws.getRow(3);
	hdrRow.height = 32;
	for (let i = 0; i < colCount; i++) {
		const cell = hdrRow.getCell(i + 1);
		cell.font = { name: 'Calibri', size: 11, bold: true, color: { argb: HEADER_FG } };
		cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: BRAND } };
		cell.alignment = { vertical: 'middle', horizontal: 'center' };
		cell.border = {
			top: { style: 'thin', color: { argb: BORDER } },
			left: { style: 'thin', color: { argb: BORDER } },
			bottom: { style: 'thin', color: { argb: BORDER } },
			right: { style: 'thin', color: { argb: BORDER } },
		};
	}

	// ── Data rows (from row 4) ──────────────────────
	for (let r = 0; r < data.length; r++) {
		const row = ws.getRow(r + 4);
		const bgColor = r % 2 === 0 ? ALT_ROW : HEADER_FG;

		for (let c = 0; c < colCount; c++) {
			const cell = row.getCell(c + 1);
			const col = columns[c];
			const val = data[r][col.key];
			cell.value = val ?? null;

			if (col.numFmt && typeof val === 'number') {
				cell.numFmt = col.numFmt;
			}

			cell.font = { name: 'Calibri', size: 10, color: { argb: '101828' } };
			cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: bgColor } };
			cell.alignment = { vertical: 'middle', horizontal: col.numFmt ? 'right' : 'left' };
			cell.border = {
				top: { style: 'thin', color: { argb: BORDER } },
				left: { style: 'thin', color: { argb: BORDER } },
				bottom: { style: 'thin', color: { argb: BORDER } },
				right: { style: 'thin', color: { argb: BORDER } },
			};
		}
		row.height = 22;
	}

	// ── Auto-filter ─────────────────────────────────
	if (data.length > 0) {
		ws.autoFilter = {
			from: { row: 3, column: 1 },
			to:   { row: 3 + data.length, column: colCount }
		};
	}

	// ── Freeze panes ────────────────────────────────
	ws.views = [{ state: 'frozen', ySplit: 3 }];

	return wb.xlsx.writeBuffer() as Promise<ArrayBuffer>;
}

async function downloadXLSX(buffer: ArrayBuffer, filename: string): Promise<void> {
	const blob = new Blob([buffer], {
		type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
	});
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = filename;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}

// ─── Per-module exporters ─────────────────────────────────────────────────

export async function exportProducts(products: Product[], shopName: string, exportedBy: string): Promise<void> {
	const columns = [
		{ header: 'Product Name',       key: 'name',        width: 30 },
		{ header: 'Barcode',            key: 'barcode',     width: 18 },
		{ header: 'SKU',                key: 'sku',         width: 16 },
		{ header: 'Category',           key: 'category',    width: 18 },
		{ header: 'Buying Price (KES)', key: 'buying',      width: 18, numFmt: '#,##0.00' },
		{ header: 'Selling Price (KES)',key: 'selling',     width: 18, numFmt: '#,##0.00' },
		{ header: 'Profit Margin (%)',  key: 'margin',      width: 16, numFmt: '0.0"%"' },
		{ header: 'Current Stock',      key: 'stock',       width: 14, numFmt: '#,##0' },
		{ header: 'Reorder Level',      key: 'reorder',     width: 14, numFmt: '#,##0' },
		{ header: 'Status',             key: 'status',      width: 12 },
		{ header: 'Created Date',       key: 'created',     width: 16 },
	];

	const data = products.map(p => {
		const margin = p.buying_price > 0
			? ((p.selling_price - p.buying_price) / p.buying_price * 100)
			: 0;
		return {
			name:     p.name ?? '',
			barcode:  p.barcode ?? '',
			sku:      p.sku ?? '',
			category: p.category_name ?? '',
			buying:   p.buying_price ?? 0,
			selling:  p.selling_price ?? 0,
			margin:   Math.round(margin * 10) / 10,
			stock:    p.stock_qty ?? 0,
			reorder:  p.reorder_level ?? 0,
			status:   p.is_active ? 'Active' : 'Inactive',
			created:  formatDate(p.created_at),
		};
	});

	const buffer = await buildWorkbook('Products', columns, data, shopName, exportedBy);
	await downloadXLSX(buffer, safeFilename(shopName, 'Products'));
}

export async function exportSales(sales: Sale[], shopName: string, exportedBy: string): Promise<void> {
	const columns = [
		{ header: 'Receipt No',         key: 'receipt',     width: 16 },
		{ header: 'Customer',           key: 'customer',    width: 22 },
		{ header: 'Cashier',            key: 'cashier',     width: 20 },
		{ header: 'Payment Method',     key: 'method',      width: 16 },
		{ header: 'Date',               key: 'date',        width: 14 },
		{ header: 'Time',               key: 'time',        width: 10 },
		{ header: 'Subtotal (KES)',     key: 'subtotal',    width: 16, numFmt: '#,##0.00' },
		{ header: 'Discount (KES)',     key: 'discount',    width: 16, numFmt: '#,##0.00' },
		{ header: 'Tax (KES)',          key: 'tax',         width: 14, numFmt: '#,##0.00' },
		{ header: 'Grand Total (KES)',  key: 'total',       width: 18, numFmt: '#,##0.00' },
		{ header: 'Status',             key: 'status',      width: 12 },
	];

	const data = sales.map(s => ({
		receipt:  s.id.slice(0, 10).toUpperCase(),
		customer: s.customer_name ?? 'Walk-in',
		cashier:  s.cashier_name ?? '',
		method:   s.payment_method.toUpperCase(),
		date:     formatDate(s.created_at),
		time:     new Date(s.created_at).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }),
		subtotal: s.subtotal ?? 0,
		discount: s.discount ?? 0,
		tax:      s.tax ?? 0,
		total:    s.total ?? 0,
		status:   s.status,
	}));

	const buffer = await buildWorkbook('Sales', columns, data, shopName, exportedBy);
	await downloadXLSX(buffer, safeFilename(shopName, 'Sales'));
}

export async function exportCustomers(customers: Customer[], shopName: string, exportedBy: string): Promise<void> {
	const columns = [
		{ header: 'Customer Name', key: 'name',    width: 28 },
		{ header: 'Phone',         key: 'phone',   width: 20 },
		{ header: 'Email',         key: 'email',   width: 30 },
		{ header: 'Address',       key: 'address', width: 30 },
		{ header: 'Date Created',  key: 'created', width: 16 },
	];

	const data = customers.map(c => ({
		name:    c.name ?? '',
		phone:   c.phone ?? '',
		email:   c.email ?? '',
		address: c.address ?? '',
		created: formatDate(c.created_at),
	}));

	const buffer = await buildWorkbook('Customers', columns, data, shopName, exportedBy);
	await downloadXLSX(buffer, safeFilename(shopName, 'Customers'));
}

export async function exportSuppliers(suppliers: Supplier[], shopName: string, exportedBy: string): Promise<void> {
	const columns = [
		{ header: 'Supplier Name', key: 'name',    width: 28 },
		{ header: 'Phone',         key: 'phone',   width: 20 },
		{ header: 'Email',         key: 'email',   width: 30 },
		{ header: 'Address',       key: 'address', width: 30 },
		{ header: 'Date Created',  key: 'created', width: 16 },
	];

	const data = suppliers.map(s => ({
		name:    s.name ?? '',
		phone:   s.phone ?? '',
		email:   s.email ?? '',
		address: s.address ?? '',
		created: formatDate(s.created_at),
	}));

	const buffer = await buildWorkbook('Suppliers', columns, data, shopName, exportedBy);
	await downloadXLSX(buffer, safeFilename(shopName, 'Suppliers'));
}

export async function exportPurchases(purchases: Purchase[], shopName: string, exportedBy: string): Promise<void> {
	const columns = [
		{ header: 'Purchase No',      key: 'purchaseNo', width: 16 },
		{ header: 'Supplier',         key: 'supplier',   width: 22 },
		{ header: 'Ordered By',       key: 'orderedBy',  width: 20 },
		{ header: 'Date',             key: 'date',       width: 14 },
		{ header: 'Items Count',      key: 'items',      width: 14, numFmt: '#,##0' },
		{ header: 'Total Cost (KES)', key: 'total',      width: 18, numFmt: '#,##0.00' },
		{ header: 'Status',           key: 'status',     width: 14 },
		{ header: 'Notes',            key: 'notes',      width: 28 },
	];

	const data = purchases.map(p => ({
		purchaseNo: p.id.slice(0, 10).toUpperCase(),
		supplier:   p.supplier_name ?? '',
		orderedBy:  p.user_name ?? '',
		date:       formatDate(p.created_at),
		items:      p.items?.length ?? 0,
		total:      p.total ?? 0,
		status:     p.status,
		notes:      p.note ?? '',
	}));

	const buffer = await buildWorkbook('Purchases', columns, data, shopName, exportedBy);
	await downloadXLSX(buffer, safeFilename(shopName, 'Purchases'));
}

export interface InventoryRow {
	product_name: string;
	stock_qty: number;
	reorder_level: number;
	buying_price: number;
	selling_price: number;
	category_name?: string | null;
}

export async function exportInventory(rows: InventoryRow[], shopName: string, exportedBy: string): Promise<void> {
	const columns = [
		{ header: 'Product',               key: 'product',  width: 28 },
		{ header: 'Category',              key: 'category', width: 18 },
		{ header: 'Current Stock',         key: 'stock',    width: 14, numFmt: '#,##0' },
		{ header: 'Reorder Level',         key: 'reorder',  width: 14, numFmt: '#,##0' },
		{ header: 'Stock Status',          key: 'status',   width: 16 },
		{ header: 'Inventory Value (KES)', key: 'invValue', width: 20, numFmt: '#,##0.00' },
		{ header: 'Retail Value (KES)',    key: 'retValue', width: 20, numFmt: '#,##0.00' },
	];

	const data = rows.map(r => {
		const status = r.stock_qty <= 0 ? 'Out of Stock'
			: r.stock_qty <= r.reorder_level ? 'Low Stock'
			: 'In Stock';
		return {
			product:  r.product_name ?? '',
			category: r.category_name ?? '',
			stock:    r.stock_qty ?? 0,
			reorder:  r.reorder_level ?? 0,
			status,
			invValue: (r.buying_price ?? 0) * (r.stock_qty ?? 0),
			retValue: (r.selling_price ?? 0) * (r.stock_qty ?? 0),
		};
	});

	const buffer = await buildWorkbook('Inventory', columns, data, shopName, exportedBy);
	await downloadXLSX(buffer, safeFilename(shopName, 'Inventory'));
}
