package repositories

import (
	"pos/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type InventoryRepo struct {
	db *sqlx.DB
}

func NewInventoryRepo(db *sqlx.DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

func (r *InventoryRepo) ListAdjustments(page, limit int) ([]models.StockAdjustment, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM stock_adjustments`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var adjs []models.StockAdjustment
	err = r.db.Select(&adjs, `
		SELECT sa.*, p.name AS product_name, u.name AS user_name
		FROM stock_adjustments sa
		JOIN products p ON sa.product_id = p.id
		JOIN users u ON sa.user_id = u.id
		ORDER BY sa.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	return adjs, total, err
}

func (r *InventoryRepo) CreateAdjustment(tx *sqlx.Tx, adj *models.StockAdjustment) error {
	adj.ID = uuid.New().String()
	return tx.QueryRowx(`
		INSERT INTO stock_adjustments (id, product_id, user_id, quantity, reason)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		adj.ID, adj.ProductID, adj.UserID, adj.Quantity, adj.Reason,
	).Scan(&adj.ID, &adj.CreatedAt)
}

func (r *InventoryRepo) GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}
	err := r.db.Get(stats, `
		SELECT
			COALESCE((SELECT SUM(total) FROM sales WHERE status = 'completed' AND DATE(created_at) = CURRENT_DATE), 0) AS today_sales,
			COALESCE((SELECT COUNT(*) FROM sales WHERE status = 'completed' AND DATE(created_at) = CURRENT_DATE), 0) AS today_orders,
			COALESCE((SELECT COUNT(*) FROM products WHERE is_active = TRUE), 0) AS total_products,
			COALESCE((SELECT COUNT(*) FROM products WHERE is_active = TRUE AND stock_qty <= reorder_level), 0) AS low_stock_count,
			COALESCE((SELECT COUNT(*) FROM customers), 0) AS total_customers,
			COALESCE((SELECT SUM(total) FROM sales WHERE status = 'completed' AND TO_CHAR(created_at, 'YYYY-MM') = TO_CHAR(NOW(), 'YYYY-MM')), 0) AS month_sales
	`)
	return stats, err
}

func (r *InventoryRepo) DB() *sqlx.DB {
	return r.db
}
