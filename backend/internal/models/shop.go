package models

type Shop struct {
	ID           string `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	BusinessType string `db:"business_type" json:"business_type"`
	Email        string `db:"email" json:"email"`
	Phone        string `db:"phone" json:"phone"`
	Address      string `db:"address" json:"address"`
	Country      string `db:"country" json:"country"`
	County       string `db:"county" json:"county"`
	Town         string `db:"town" json:"town"`
	Currency     string `db:"currency" json:"currency"`
	Timezone     string `db:"timezone" json:"timezone"`
	Logo         string `db:"logo" json:"logo"`
	CreatedAt    Time   `db:"created_at" json:"created_at"`
	UpdatedAt    Time   `db:"updated_at" json:"updated_at"`
}

type ShopSettings struct {
	ID                  string `db:"id" json:"id"`
	ShopID              string `db:"shop_id" json:"shop_id"`
	TaxRate             float64 `db:"tax_rate" json:"tax_rate"`
	ReceiptFooter       string `db:"receipt_footer" json:"receipt_footer"`
	InvoicePrefix       string `db:"invoice_prefix" json:"invoice_prefix"`
	DefaultPaymentMethod string `db:"default_payment_method" json:"default_payment_method"`
	LowStockThreshold   int    `db:"low_stock_threshold" json:"low_stock_threshold"`
	EnableNotifications bool   `db:"enable_notifications" json:"enable_notifications"`
	Currency            string `db:"currency" json:"currency"`
	CreatedAt           Time   `db:"created_at" json:"created_at"`
	UpdatedAt           Time   `db:"updated_at" json:"updated_at"`
}

type RegisterInput struct {
	ShopName      string `json:"shop_name" validate:"required,min=2,max=200"`
	BusinessType  string `json:"business_type" validate:"required"`
	BusinessEmail string `json:"business_email" validate:"omitempty,email"`
	BusinessPhone string `json:"business_phone" validate:"omitempty"`
	Country       string `json:"country" validate:"required"`
	County        string `json:"county" validate:"required"`
	Town          string `json:"town" validate:"required"`
	Address       string `json:"address" validate:"omitempty"`
	Currency      string `json:"currency" validate:"required"`
	Timezone      string `json:"timezone" validate:"required"`
	OwnerName     string `json:"owner_name" validate:"required,min=2,max=100"`
	OwnerEmail    string `json:"owner_email" validate:"required,email"`
	OwnerPhone    string `json:"owner_phone" validate:"required"`
	Username      string `json:"username" validate:"required,min=3,max=50"`
	Password      string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type RegisterResult struct {
	User  *User  `json:"user"`
	Shop  *Shop  `json:"shop"`
	Token string `json:"token"`
}
