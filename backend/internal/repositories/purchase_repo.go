package repositories

import (
	"pos/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PurchaseRepo struct {
	db *sqlx.DB
}

func NewPurchaseRepo(db *sqlx.DB) *PurchaseRepo {
	return &PurchaseRepo{db: db}
}

func (r *PurchaseRepo) List(shopID string, page, limit int) ([]models.Purchase, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM purchases WHERE shop_id = $1`, shopID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var purchases []models.Purchase
	err = r.db.Select(&purchases, `
		SELECT p.*, s.name AS supplier_name, u.name AS user_name
		FROM purchases p
		LEFT JOIN suppliers s ON p.supplier_id = s.id
		JOIN users u ON p.user_id = u.id
		WHERE p.shop_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`, shopID, limit, offset)
	return purchases, total, err
}

func (r *PurchaseRepo) FindByID(shopID string, id string) (*models.Purchase, error) {
	p := &models.Purchase{}
	err := r.db.Get(p, `
		SELECT p.*, s.name AS supplier_name, u.name AS user_name
		FROM purchases p
		LEFT JOIN suppliers s ON p.supplier_id = s.id
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1 AND p.shop_id = $2`, id, shopID)
	if err != nil {
		return nil, err
	}

	err = r.db.Select(&p.Items, `
		SELECT pi.*, pr.name AS product_name
		FROM purchase_items pi
		JOIN products pr ON pi.product_id = pr.id
		WHERE pi.purchase_id = $1`, id)
	return p, err
}

func (r *PurchaseRepo) Create(tx *sqlx.Tx, shopID string, p *models.Purchase) error {
	p.ID = uuid.New().String()
	return tx.QueryRowx(`
		INSERT INTO purchases (id, supplier_id, user_id, total, status, note, shop_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`,
		p.ID, p.SupplierID, p.UserID, p.Total, p.Status, p.Note, shopID,
	).Scan(&p.ID, &p.CreatedAt)
}

func (r *PurchaseRepo) CreateItem(tx *sqlx.Tx, shopID string, item *models.PurchaseItem) error {
	item.ID = uuid.New().String()
	return tx.QueryRowx(`
		INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, total, shop_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		item.ID, item.PurchaseID, item.ProductID, item.Quantity, item.UnitPrice, item.Total, shopID,
	).Scan(&item.ID)
}

func (r *PurchaseRepo) DB() *sqlx.DB {
	return r.db
}
