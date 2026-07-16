package repositories

import (
	"fmt"
	"pos/internal/models"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ProductRepo struct {
	db *sqlx.DB
}

func NewProductRepo(db *sqlx.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) List(f models.ProductFilter) ([]models.Product, int, error) {
	where := []string{"p.is_active = TRUE"}
	args := []interface{}{}
	i := 1

	if f.Search != "" {
		where = append(where, fmt.Sprintf("(p.name ILIKE $%d OR p.barcode = $%d OR p.sku ILIKE $%d)", i, i+1, i+2))
		like := "%" + f.Search + "%"
		args = append(args, like, f.Search, like)
		i += 3
	}
	if f.CategoryID != "" {
		where = append(where, fmt.Sprintf("p.category_id = $%d", i))
		args = append(args, f.CategoryID)
		i++
	}
	if f.LowStock {
		where = append(where, "p.stock_qty <= p.reorder_level")
	}

	whereClause := strings.Join(where, " AND ")
	base := fmt.Sprintf(`
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE %s`, whereClause)

	var total int
	err := r.db.QueryRow("SELECT COUNT(*) "+base, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)
	query := fmt.Sprintf(`SELECT p.*, c.name AS category_name %s ORDER BY p.name ASC LIMIT $%d OFFSET $%d`, base, i, i+1)

	var products []models.Product
	err = r.db.Select(&products, query, args...)
	return products, total, err
}

func (r *ProductRepo) FindByID(id string) (*models.Product, error) {
	p := &models.Product{}
	err := r.db.Get(p, `
		SELECT p.*, c.name AS category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1`, id)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProductRepo) FindByBarcode(barcode string) (*models.Product, error) {
	p := &models.Product{}
	err := r.db.Get(p, `
		SELECT p.*, c.name AS category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.barcode = $1 AND p.is_active = TRUE`, barcode)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProductRepo) Create(p *models.Product) error {
	p.ID = uuid.New().String()
	return r.db.QueryRowx(`
		INSERT INTO products (id, name, barcode, sku, category_id, buying_price, selling_price, stock_qty, reorder_level, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`,
		p.ID, p.Name, p.Barcode, p.SKU, p.CategoryID, p.BuyingPrice, p.SellingPrice, p.StockQty, p.ReorderLevel, p.ImageURL,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *ProductRepo) Update(id string, in *models.ProductInput) error {
	_, err := r.db.Exec(`
		UPDATE products SET
			name = $1, barcode = $2, sku = $3, category_id = $4,
			buying_price = $5, selling_price = $6, stock_qty = $7,
			reorder_level = $8, image_url = $9, updated_at = NOW()
		WHERE id = $10`,
		in.Name, in.Barcode, in.SKU, in.CategoryID,
		in.BuyingPrice, in.SellingPrice, in.StockQty,
		in.ReorderLevel, in.ImageURL, id,
	)
	return err
}

func (r *ProductRepo) Delete(id string) error {
	_, err := r.db.Exec(`UPDATE products SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *ProductRepo) GetStock(tx *sqlx.Tx, productID string) (string, int, error) {
	var name string
	var stock int
	err := tx.QueryRow(`SELECT name, stock_qty FROM products WHERE id = $1`, productID).Scan(&name, &stock)
	return name, stock, err
}

func (r *ProductRepo) UpdateStock(tx *sqlx.Tx, productID string, delta int) error {
	_, err := tx.Exec(
		`UPDATE products SET stock_qty = stock_qty + $1, updated_at = NOW() WHERE id = $2`,
		delta, productID,
	)
	return err
}

func (r *ProductRepo) ListCategories() ([]models.Category, error) {
	var cats []models.Category
	err := r.db.Select(&cats, `SELECT * FROM categories ORDER BY name ASC`)
	return cats, err
}

func (r *ProductRepo) CreateCategory(c *models.Category) error {
	c.ID = uuid.New().String()
	return r.db.QueryRowx(
		`INSERT INTO categories (id, name, description) VALUES ($1, $2, $3) RETURNING id, created_at`,
		c.ID, c.Name, c.Description,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *ProductRepo) UpdateCategory(id string, in *models.CategoryInput) error {
	_, err := r.db.Exec(
		`UPDATE categories SET name = $1, description = $2 WHERE id = $3`,
		in.Name, in.Description, id,
	)
	return err
}

func (r *ProductRepo) DeleteCategory(id string) error {
	_, err := r.db.Exec(`DELETE FROM categories WHERE id = $1`, id)
	return err
}
