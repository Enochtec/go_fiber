package repositories

import (
	"pos/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SupplierRepo struct {
	db *sqlx.DB
}

func NewSupplierRepo(db *sqlx.DB) *SupplierRepo {
	return &SupplierRepo{db: db}
}

func (r *SupplierRepo) List(shopID string) ([]models.Supplier, error) {
	var suppliers []models.Supplier
	err := r.db.Select(&suppliers, `SELECT * FROM suppliers WHERE shop_id = $1 ORDER BY name ASC`, shopID)
	return suppliers, err
}

func (r *SupplierRepo) FindByID(shopID string, id string) (*models.Supplier, error) {
	s := &models.Supplier{}
	err := r.db.Get(s, `SELECT * FROM suppliers WHERE id = $1 AND shop_id = $2`, id, shopID)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SupplierRepo) Create(shopID string, s *models.Supplier) error {
	s.ID = uuid.New().String()
	return r.db.QueryRowx(
		`INSERT INTO suppliers (id, name, email, phone, address, shop_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`,
		s.ID, s.Name, s.Email, s.Phone, s.Address, shopID,
	).Scan(&s.ID, &s.CreatedAt)
}

func (r *SupplierRepo) Update(shopID string, id string, in *models.SupplierInput) error {
	_, err := r.db.Exec(
		`UPDATE suppliers SET name = $1, email = $2, phone = $3, address = $4 WHERE id = $5 AND shop_id = $6`,
		in.Name, in.Email, in.Phone, in.Address, id, shopID,
	)
	return err
}

func (r *SupplierRepo) Delete(shopID string, id string) error {
	_, err := r.db.Exec(`DELETE FROM suppliers WHERE id = $1 AND shop_id = $2`, id, shopID)
	return err
}
