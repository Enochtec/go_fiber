package repositories

import (
	"fmt"
	"pos/internal/models"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SaleRepo struct {
	db *sqlx.DB
}

func NewSaleRepo(db *sqlx.DB) *SaleRepo {
	return &SaleRepo{db: db}
}

func (r *SaleRepo) List(shopID string, f models.SaleFilter) ([]models.Sale, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	i := 1

	where = append(where, fmt.Sprintf("s.shop_id = $%d", i))
	args = append(args, shopID)
	i++

	if f.Status != "" {
		where = append(where, fmt.Sprintf("s.status = $%d", i))
		args = append(args, f.Status)
		i++
	}
	if f.CashierID != "" {
		where = append(where, fmt.Sprintf("s.cashier_id = $%d", i))
		args = append(args, f.CashierID)
		i++
	}
	if f.DateFrom != "" {
		where = append(where, fmt.Sprintf("s.created_at >= $%d", i))
		args = append(args, f.DateFrom)
		i++
	}
	if f.DateTo != "" {
		where = append(where, fmt.Sprintf("s.created_at < $%d", i))
		args = append(args, f.DateTo)
		i++
	}

	whereClause := strings.Join(where, " AND ")
	base := fmt.Sprintf(`
		FROM sales s
		LEFT JOIN users u ON s.cashier_id = u.id
		LEFT JOIN customers c ON s.customer_id = c.id
		WHERE %s`, whereClause)

	var total int
	err := r.db.QueryRow("SELECT COUNT(*) "+base, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)
	query := fmt.Sprintf(`
		SELECT s.*, u.name AS cashier_name, c.name AS customer_name
		%s ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d`, base, i, i+1)

	var sales []models.Sale
	err = r.db.Select(&sales, query, args...)
	return sales, total, err
}

func (r *SaleRepo) FindByID(shopID string, id string) (*models.Sale, error) {
	sale := &models.Sale{}
	err := r.db.Get(sale, `
		SELECT s.*, u.name AS cashier_name, c.name AS customer_name
		FROM sales s
		LEFT JOIN users u ON s.cashier_id = u.id
		LEFT JOIN customers c ON s.customer_id = c.id
		WHERE s.id = $1 AND s.shop_id = $2`, id, shopID)
	if err != nil {
		return nil, err
	}

	err = r.db.Select(&sale.Items, `
		SELECT si.*, p.name AS product_name
		FROM sale_items si
		JOIN products p ON si.product_id = p.id
		WHERE si.sale_id = $1`, id)
	return sale, err
}

func (r *SaleRepo) Create(tx *sqlx.Tx, shopID string, sale *models.Sale) error {
	sale.ID = uuid.New().String()
	return tx.QueryRowx(`
		INSERT INTO sales (id, cashier_id, customer_id, subtotal, discount, tax, total, payment_method, status, note, shop_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at`,
		sale.ID, sale.CashierID, sale.CustomerID, sale.Subtotal, sale.Discount,
		sale.Tax, sale.Total, sale.PaymentMethod, sale.Status, sale.Note, shopID,
	).Scan(&sale.ID, &sale.CreatedAt)
}

func (r *SaleRepo) CreateItem(tx *sqlx.Tx, shopID string, item *models.SaleItem) error {
	item.ID = uuid.New().String()
	return tx.QueryRowx(`
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_price, total, shop_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		item.ID, item.SaleID, item.ProductID, item.Quantity, item.UnitPrice, item.Total, shopID,
	).Scan(&item.ID)
}

func (r *SaleRepo) UpdateStatus(shopID string, id string, status models.SaleStatus) error {
	_, err := r.db.Exec(`UPDATE sales SET status = $1 WHERE id = $2 AND shop_id = $3`, status, id, shopID)
	return err
}

func (r *SaleRepo) DB() *sqlx.DB {
	return r.db
}
