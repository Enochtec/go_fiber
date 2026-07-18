package models

import (
	"database/sql"
	"time"
)

type Shift struct {
	ID               string         `db:"id" json:"id"`
	ShopID           sql.NullString `db:"shop_id" json:"shop_id"`
	CashierID        string         `db:"cashier_id" json:"cashier_id"`
	CashierName      string         `db:"cashier_name" json:"cashier_name"`
	OpeningFloat     float64        `db:"opening_float" json:"opening_float"`
	OpeningTime      time.Time      `db:"opening_time" json:"opening_time"`
	ClosingTime      *time.Time     `db:"closing_time" json:"closing_time,omitempty"`
	CashSales        float64        `db:"cash_sales" json:"cash_sales"`
	MpesaSales       float64        `db:"mpesa_sales" json:"mpesa_sales"`
	CardSales        float64        `db:"card_sales" json:"card_sales"`
	OtherSales       float64        `db:"other_sales" json:"other_sales"`
	TotalSales       float64        `db:"total_sales" json:"total_sales"`
	TransactionCount int            `db:"transaction_count" json:"transaction_count"`
	ClosingFloat     *float64       `db:"closing_float" json:"closing_float,omitempty"`
	ExpectedCash     *float64       `db:"expected_cash" json:"expected_cash,omitempty"`
	ActualCash       *float64       `db:"actual_cash" json:"actual_cash,omitempty"`
	Variance         *float64       `db:"variance" json:"variance,omitempty"`
	Status           string         `db:"status" json:"status"`
	Notes            string         `db:"notes" json:"notes"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
}

type OpenShiftInput struct {
	OpeningFloat float64 `json:"opening_float" validate:"min=0"`
	Notes        string  `json:"notes"`
}

type CloseShiftInput struct {
	ActualCash float64 `json:"actual_cash" validate:"min=0"`
	Notes      string  `json:"notes"`
}
