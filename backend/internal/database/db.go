package database

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

func Connect() (*sqlx.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/pos_db?sslmode=disable"
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

func migrate(db *sqlx.DB) error {
	schemas := []string{
		schemaV1,
		schemaShops,
		schemaShopSettings,
		schemaUsersV2,
	}

	for _, s := range schemas {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM users`); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	id := uuid.New().String()
	_, err = db.Exec(
		`INSERT INTO users (id, name, email, password, role, is_active) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, "Administrator", "admin@pos.local", string(hash), "admin", true,
	)
	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	fmt.Println("Seeded default admin user (admin@pos.local / admin123)")
	return nil
}

const schemaV1 = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	role TEXT NOT NULL DEFAULT 'cashier',
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	barcode TEXT UNIQUE,
	sku TEXT UNIQUE,
	category_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
	buying_price DOUBLE PRECISION NOT NULL DEFAULT 0,
	selling_price DOUBLE PRECISION NOT NULL DEFAULT 0,
	stock_qty INTEGER NOT NULL DEFAULT 0,
	reorder_level INTEGER NOT NULL DEFAULT 10,
	image_url TEXT,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
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
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);

CREATE TABLE IF NOT EXISTS suppliers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT,
	phone TEXT,
	address TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sales (
	id TEXT PRIMARY KEY,
	cashier_id TEXT NOT NULL REFERENCES users(id),
	customer_id TEXT REFERENCES customers(id),
	subtotal DOUBLE PRECISION NOT NULL DEFAULT 0,
	discount DOUBLE PRECISION NOT NULL DEFAULT 0,
	tax DOUBLE PRECISION NOT NULL DEFAULT 0,
	total DOUBLE PRECISION NOT NULL DEFAULT 0,
	payment_method TEXT NOT NULL DEFAULT 'cash',
	status TEXT NOT NULL DEFAULT 'completed',
	note TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sales_cashier ON sales(cashier_id);
CREATE INDEX IF NOT EXISTS idx_sales_created_at ON sales(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sales_status ON sales(status);

CREATE TABLE IF NOT EXISTS sale_items (
	id TEXT PRIMARY KEY,
	sale_id TEXT NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
	product_id TEXT NOT NULL REFERENCES products(id),
	quantity INTEGER NOT NULL,
	unit_price DOUBLE PRECISION NOT NULL,
	total DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sale_items_sale ON sale_items(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_items_product ON sale_items(product_id);

CREATE TABLE IF NOT EXISTS purchases (
	id TEXT PRIMARY KEY,
	supplier_id TEXT REFERENCES suppliers(id) ON DELETE SET NULL,
	user_id TEXT NOT NULL REFERENCES users(id),
	total DOUBLE PRECISION NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'received',
	note TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_purchases_created_at ON purchases(created_at DESC);

CREATE TABLE IF NOT EXISTS purchase_items (
	id TEXT PRIMARY KEY,
	purchase_id TEXT NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
	product_id TEXT NOT NULL REFERENCES products(id),
	quantity INTEGER NOT NULL,
	unit_price DOUBLE PRECISION NOT NULL,
	total DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_purchase_items_purchase ON purchase_items(purchase_id);

CREATE TABLE IF NOT EXISTS stock_adjustments (
	id TEXT PRIMARY KEY,
	product_id TEXT NOT NULL REFERENCES products(id),
	user_id TEXT NOT NULL REFERENCES users(id),
	quantity INTEGER NOT NULL,
	reason TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stock_adj_product ON stock_adjustments(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_adj_created_at ON stock_adjustments(created_at DESC);

DROP TABLE IF EXISTS shifts CASCADE;

CREATE TABLE IF NOT EXISTS shifts (
	id UUID PRIMARY KEY,
	cashier_id UUID NOT NULL,
	opening_float DOUBLE PRECISION NOT NULL DEFAULT 0,
	opening_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	closing_time TIMESTAMPTZ,
	cash_sales DOUBLE PRECISION NOT NULL DEFAULT 0,
	mpesa_sales DOUBLE PRECISION NOT NULL DEFAULT 0,
	card_sales DOUBLE PRECISION NOT NULL DEFAULT 0,
	other_sales DOUBLE PRECISION NOT NULL DEFAULT 0,
	total_sales DOUBLE PRECISION NOT NULL DEFAULT 0,
	transaction_count INTEGER NOT NULL DEFAULT 0,
	closing_float DOUBLE PRECISION,
	expected_cash DOUBLE PRECISION,
	actual_cash DOUBLE PRECISION,
	variance DOUBLE PRECISION,
	status TEXT NOT NULL DEFAULT 'open',
	notes TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shifts_cashier ON shifts(cashier_id);
CREATE INDEX IF NOT EXISTS idx_shifts_status ON shifts(status);
CREATE INDEX IF NOT EXISTS idx_shifts_created_at ON shifts(created_at DESC);
`

const schemaShops = `
CREATE TABLE IF NOT EXISTS shops (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	business_type TEXT NOT NULL DEFAULT 'retail',
	email TEXT DEFAULT '',
	phone TEXT DEFAULT '',
	address TEXT DEFAULT '',
	country TEXT NOT NULL DEFAULT '',
	county TEXT NOT NULL DEFAULT '',
	town TEXT NOT NULL DEFAULT '',
	currency TEXT NOT NULL DEFAULT 'KES',
	timezone TEXT NOT NULL DEFAULT 'Africa/Nairobi',
	logo TEXT DEFAULT '',
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

const schemaShopSettings = `
CREATE TABLE IF NOT EXISTS shop_settings (
	id TEXT PRIMARY KEY,
	shop_id TEXT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
	tax_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
	receipt_footer TEXT DEFAULT 'Thank you for your business!',
	invoice_prefix TEXT DEFAULT 'INV-',
	default_payment_method TEXT DEFAULT 'cash',
	low_stock_threshold INTEGER NOT NULL DEFAULT 10,
	enable_notifications BOOLEAN NOT NULL DEFAULT TRUE,
	currency TEXT NOT NULL DEFAULT 'KES',
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_settings_shop ON shop_settings(shop_id);
`

const schemaUsersV2 = `
ALTER TABLE users ADD COLUMN IF NOT EXISTS shop_id TEXT REFERENCES shops(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone TEXT DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_shop ON users(shop_id);
`
