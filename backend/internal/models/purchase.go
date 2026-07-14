package models

type PurchaseStatus string

const (
	PurchaseReceived PurchaseStatus = "received"
	PurchasePending  PurchaseStatus = "pending"
)

type Purchase struct {
	ID           string         `db:"id" json:"id"`
	SupplierID   *string        `db:"supplier_id" json:"supplier_id"`
	SupplierName *string        `db:"supplier_name" json:"supplier_name"`
	UserID       string         `db:"user_id" json:"user_id"`
	UserName     *string        `db:"user_name" json:"user_name"`
	Total        float64        `db:"total" json:"total"`
	Status       PurchaseStatus `db:"status" json:"status"`
	Note         *string        `db:"note" json:"note"`
	CreatedAt    Time           `db:"created_at" json:"created_at"`
	Items        []PurchaseItem `db:"-" json:"items,omitempty"`
}

type PurchaseItem struct {
	ID          string  `db:"id" json:"id"`
	PurchaseID  string  `db:"purchase_id" json:"purchase_id"`
	ProductID   string  `db:"product_id" json:"product_id"`
	ProductName *string `db:"product_name" json:"product_name"`
	Quantity    int     `db:"quantity" json:"quantity"`
	UnitPrice   float64 `db:"unit_price" json:"unit_price"`
	Total       float64 `db:"total" json:"total"`
}

type PurchaseItemInput struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	Quantity  int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" validate:"required,gt=0"`
}

type CreatePurchaseInput struct {
	SupplierID *string             `json:"supplier_id"`
	Items      []PurchaseItemInput `json:"items" validate:"required,min=1,dive"`
	Status     PurchaseStatus      `json:"status" validate:"omitempty,oneof=received pending"`
	Note       *string             `json:"note"`
}
