package models

type StockAdjustment struct {
	ID          string  `db:"id" json:"id"`
	ProductID   string  `db:"product_id" json:"product_id"`
	ProductName *string `db:"product_name" json:"product_name"`
	UserID      string  `db:"user_id" json:"user_id"`
	UserName    *string `db:"user_name" json:"user_name"`
	Quantity    int     `db:"quantity" json:"quantity"`
	Reason      string  `db:"reason" json:"reason"`
	CreatedAt   Time    `db:"created_at" json:"created_at"`
}

type StockAdjustmentInput struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required"`
	Reason    string `json:"reason" validate:"required,min=1,max=200"`
}

type DashboardStats struct {
	TodaySales    float64 `json:"today_sales" db:"today_sales"`
	TodayOrders   int     `json:"today_orders" db:"today_orders"`
	TotalProducts int     `json:"total_products" db:"total_products"`
	LowStockCount int     `json:"low_stock_count" db:"low_stock_count"`
	TotalCustomers int    `json:"total_customers" db:"total_customers"`
	MonthSales    float64 `json:"month_sales" db:"month_sales"`
}
