package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func Connect() (*sqlx.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "maestro.db"
	}

	db, err := sqlx.Connect("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func migrate(db *sqlx.DB) error {
	_, err := db.Exec(schema)
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	role TEXT NOT NULL DEFAULT 'cashier',
	is_active INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS categories (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS products (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	barcode TEXT UNIQUE,
	sku TEXT UNIQUE,
	category_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
	buying_price REAL NOT NULL DEFAULT 0,
	selling_price REAL NOT NULL DEFAULT 0,
	stock_qty INTEGER NOT NULL DEFAULT 0,
	reorder_level INTEGER NOT NULL DEFAULT 10,
	image_url TEXT,
	is_active INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_products_barcode ON products(barcode);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);

CREATE TABLE IF NOT EXISTS customers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT,
	phone TEXT,
	address TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);

CREATE TABLE IF NOT EXISTS suppliers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT,
	phone TEXT,
	address TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS sales (
	id TEXT PRIMARY KEY,
	cashier_id TEXT NOT NULL REFERENCES users(id),
	customer_id TEXT REFERENCES customers(id),
	subtotal REAL NOT NULL DEFAULT 0,
	discount REAL NOT NULL DEFAULT 0,
	tax REAL NOT NULL DEFAULT 0,
	total REAL NOT NULL DEFAULT 0,
	payment_method TEXT NOT NULL DEFAULT 'cash',
	status TEXT NOT NULL DEFAULT 'completed',
	note TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_sales_cashier ON sales(cashier_id);
CREATE INDEX IF NOT EXISTS idx_sales_created_at ON sales(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sales_status ON sales(status);

CREATE TABLE IF NOT EXISTS sale_items (
	id TEXT PRIMARY KEY,
	sale_id TEXT NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
	product_id TEXT NOT NULL REFERENCES products(id),
	quantity INTEGER NOT NULL,
	unit_price REAL NOT NULL,
	total REAL NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sale_items_sale ON sale_items(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_items_product ON sale_items(product_id);

CREATE TABLE IF NOT EXISTS purchases (
	id TEXT PRIMARY KEY,
	supplier_id TEXT REFERENCES suppliers(id) ON DELETE SET NULL,
	user_id TEXT NOT NULL REFERENCES users(id),
	total REAL NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'received',
	note TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_purchases_created_at ON purchases(created_at DESC);

CREATE TABLE IF NOT EXISTS purchase_items (
	id TEXT PRIMARY KEY,
	purchase_id TEXT NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
	product_id TEXT NOT NULL REFERENCES products(id),
	quantity INTEGER NOT NULL,
	unit_price REAL NOT NULL,
	total REAL NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_purchase_items_purchase ON purchase_items(purchase_id);

CREATE TABLE IF NOT EXISTS stock_adjustments (
	id TEXT PRIMARY KEY,
	product_id TEXT NOT NULL REFERENCES products(id),
	user_id TEXT NOT NULL REFERENCES users(id),
	quantity INTEGER NOT NULL,
	reason TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_stock_adj_product ON stock_adjustments(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_adj_created_at ON stock_adjustments(created_at DESC);
`
