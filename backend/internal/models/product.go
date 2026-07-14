package models

type Category struct {
	ID          string  `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`
	Description *string `db:"description" json:"description"`
	CreatedAt   Time    `db:"created_at" json:"created_at"`
}

type CategoryInput struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description"`
}

type Product struct {
	ID           string  `db:"id" json:"id"`
	Name         string  `db:"name" json:"name"`
	Barcode      *string `db:"barcode" json:"barcode"`
	SKU          *string `db:"sku" json:"sku"`
	CategoryID   *string `db:"category_id" json:"category_id"`
	CategoryName *string `db:"category_name" json:"category_name"`
	BuyingPrice  float64 `db:"buying_price" json:"buying_price"`
	SellingPrice float64 `db:"selling_price" json:"selling_price"`
	StockQty     int     `db:"stock_qty" json:"stock_qty"`
	ReorderLevel int     `db:"reorder_level" json:"reorder_level"`
	ImageURL     *string `db:"image_url" json:"image_url"`
	IsActive     bool    `db:"is_active" json:"is_active"`
	CreatedAt    Time    `db:"created_at" json:"created_at"`
	UpdatedAt    Time    `db:"updated_at" json:"updated_at"`
}

type ProductInput struct {
	Name         string  `json:"name" validate:"required,min=1,max=200"`
	Barcode      *string `json:"barcode"`
	SKU          *string `json:"sku"`
	CategoryID   *string `json:"category_id"`
	BuyingPrice  float64 `json:"buying_price" validate:"gte=0"`
	SellingPrice float64 `json:"selling_price" validate:"gte=0"`
	StockQty     int     `json:"stock_qty" validate:"gte=0"`
	ReorderLevel int     `json:"reorder_level" validate:"gte=0"`
	ImageURL     *string `json:"image_url"`
}

type ProductFilter struct {
	Search     string
	CategoryID string
	LowStock   bool
	Page       int
	Limit      int
}
