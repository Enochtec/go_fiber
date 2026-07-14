package models

type SaleStatus string
type PaymentMethod string

const (
	SaleCompleted SaleStatus = "completed"
	SaleHeld      SaleStatus = "held"
	SaleVoided    SaleStatus = "voided"

	PaymentCash   PaymentMethod = "cash"
	PaymentMpesa  PaymentMethod = "mpesa"
	PaymentBank   PaymentMethod = "bank"
	PaymentCard   PaymentMethod = "card"
	PaymentCredit PaymentMethod = "credit"
)

type Sale struct {
	ID            string        `db:"id" json:"id"`
	CashierID     string        `db:"cashier_id" json:"cashier_id"`
	CashierName   *string       `db:"cashier_name" json:"cashier_name"`
	CustomerID    *string       `db:"customer_id" json:"customer_id"`
	CustomerName  *string       `db:"customer_name" json:"customer_name"`
	Subtotal      float64       `db:"subtotal" json:"subtotal"`
	Discount      float64       `db:"discount" json:"discount"`
	Tax           float64       `db:"tax" json:"tax"`
	Total         float64       `db:"total" json:"total"`
	PaymentMethod PaymentMethod `db:"payment_method" json:"payment_method"`
	Status        SaleStatus    `db:"status" json:"status"`
	Note          *string       `db:"note" json:"note"`
	CreatedAt     Time          `db:"created_at" json:"created_at"`
	Items         []SaleItem    `db:"-" json:"items,omitempty"`
}

type SaleItem struct {
	ID          string  `db:"id" json:"id"`
	SaleID      string  `db:"sale_id" json:"sale_id"`
	ProductID   string  `db:"product_id" json:"product_id"`
	ProductName *string `db:"product_name" json:"product_name"`
	Quantity    int     `db:"quantity" json:"quantity"`
	UnitPrice   float64 `db:"unit_price" json:"unit_price"`
	Total       float64 `db:"total" json:"total"`
}

type SaleItemInput struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Quantity  int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" validate:"required,gt=0"`
}

type CreateSaleInput struct {
	CustomerID    *string        `json:"customer_id"`
	Items         []SaleItemInput `json:"items" validate:"required,min=1,dive"`
	Discount      float64        `json:"discount" validate:"gte=0"`
	TaxRate       float64        `json:"tax_rate" validate:"gte=0"`
	PaymentMethod PaymentMethod  `json:"payment_method" validate:"required,oneof=cash mpesa bank card credit"`
	Status        SaleStatus     `json:"status" validate:"omitempty,oneof=completed held"`
	Note          *string        `json:"note"`
}

type SaleFilter struct {
	Status    string
	CashierID string
	DateFrom  string
	DateTo    string
	Page      int
	Limit     int
}
