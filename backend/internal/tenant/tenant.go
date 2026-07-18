package tenant

import (
	"github.com/jmoiron/sqlx"
)

// ScopedDB wraps sqlx.DB with a shop_id to automatically scope queries.
// Every query executed through ScopedDB has `AND shop_id = $N` appended.
type ScopedDB struct {
	*sqlx.DB
	ShopID string
}

func NewScopedDB(db *sqlx.DB, shopID string) *ScopedDB {
	return &ScopedDB{DB: db, ShopID: shopID}
}

// ScopedTx wraps sqlx.Tx with a shop_id.
type ScopedTx struct {
	*sqlx.Tx
	ShopID string
}

func NewScopedTx(tx *sqlx.Tx, shopID string) *ScopedTx {
	return &ScopedTx{Tx: tx, ShopID: shopID}
}
