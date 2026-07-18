package services

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type ReportService struct {
	db *sqlx.DB
}

func NewReportService(db *sqlx.DB) *ReportService {
	return &ReportService{db: db}
}

type DailySalesRow struct {
	Date   string  `db:"date" json:"date"`
	Orders int     `db:"orders" json:"orders"`
	Total  float64 `db:"total" json:"total"`
}

type TopProductRow struct {
	ProductID    string  `db:"product_id" json:"product_id"`
	ProductName  string  `db:"product_name" json:"product_name"`
	Quantity     int     `db:"quantity" json:"quantity"`
	QuantitySold int     `db:"quantity_sold" json:"quantity_sold"`
	Revenue      float64 `db:"revenue" json:"revenue"`
}

type ProfitRow struct {
	Date   string  `db:"date" json:"date"`
	Sales  float64 `db:"sales" json:"sales"`
	Cost   float64 `db:"cost" json:"cost"`
	Profit float64 `db:"profit" json:"profit"`
}

type InventoryValueRow struct {
	CategoryName string  `db:"category_name" json:"category_name"`
	ProductCount int     `db:"product_count" json:"product_count"`
	TotalCost    float64 `db:"total_cost" json:"total_cost"`
	TotalValue   float64 `db:"total_value" json:"total_value"`
}

func (s *ReportService) DailySales(shopID string, days int) ([]DailySalesRow, error) {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	var rows []DailySalesRow
	err := s.db.Select(&rows, `
		SELECT
			DATE(created_at) AS date,
			COUNT(*) AS orders,
			COALESCE(SUM(total), 0) AS total
		FROM sales
		WHERE status = 'completed'
			AND shop_id = $1
			AND created_at >= $2
		GROUP BY DATE(created_at)
		ORDER BY date DESC`, shopID, since)
	return rows, err
}

func (s *ReportService) MonthlySales(shopID string, months int) ([]DailySalesRow, error) {
	since := time.Now().AddDate(0, -months, 0).Format("2006-01-02 15:04:05")
	var rows []DailySalesRow
	err := s.db.Select(&rows, `
		SELECT
			TO_CHAR(created_at, 'YYYY-MM') AS date,
			COUNT(*) AS orders,
			COALESCE(SUM(total), 0) AS total
		FROM sales
		WHERE status = 'completed'
			AND shop_id = $1
			AND created_at >= $2
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY TO_CHAR(created_at, 'YYYY-MM') DESC`, shopID, since)
	return rows, err
}

func (s *ReportService) TopProducts(shopID string, limit int) ([]TopProductRow, error) {
	since := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	var rows []TopProductRow
	err := s.db.Select(&rows, `
		SELECT
			si.product_id,
			p.name AS product_name,
			SUM(si.quantity) AS quantity,
			SUM(si.quantity) AS quantity_sold,
			SUM(si.total) AS revenue
		FROM sale_items si
		JOIN products p ON si.product_id = p.id
		JOIN sales s ON si.sale_id = s.id
		WHERE s.status = 'completed'
			AND s.shop_id = $1
			AND s.created_at >= $2
		GROUP BY si.product_id, p.name
		ORDER BY quantity DESC
		LIMIT $3`, shopID, since, limit)
	return rows, err
}

func (s *ReportService) InventoryValue(shopID string) ([]InventoryValueRow, error) {
	var rows []InventoryValueRow
	err := s.db.Select(&rows, `
		SELECT
			COALESCE(c.name, 'Uncategorized') AS category_name,
			COUNT(p.id) AS product_count,
			COALESCE(SUM(p.buying_price * p.stock_qty), 0) AS total_cost,
			COALESCE(SUM(p.selling_price * p.stock_qty), 0) AS total_value
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.is_active = TRUE
			AND p.shop_id = $1
		GROUP BY c.name
		ORDER BY total_value DESC`, shopID)
	return rows, err
}
