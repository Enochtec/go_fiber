package models

import "database/sql"

type Customer struct {
	ID        string         `db:"id" json:"id"`
	ShopID    sql.NullString `db:"shop_id" json:"shop_id"`
	Name      string         `db:"name" json:"name"`
	Email     *string        `db:"email" json:"email"`
	Phone     *string        `db:"phone" json:"phone"`
	Address   *string        `db:"address" json:"address"`
	CreatedAt Time           `db:"created_at" json:"created_at"`
}

type CustomerInput struct {
	Name    string  `json:"name" validate:"required,min=1,max=100"`
	Email   *string `json:"email" validate:"omitempty,email"`
	Phone   *string `json:"phone"`
	Address *string `json:"address"`
}
