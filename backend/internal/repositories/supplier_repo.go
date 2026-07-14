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

func (r *SupplierRepo) List() ([]models.Supplier, error) {
	var suppliers []models.Supplier
	err := r.db.Select(&suppliers, `SELECT * FROM suppliers ORDER BY name ASC`)
	return suppliers, err
}

func (r *SupplierRepo) FindByID(id string) (*models.Supplier, error) {
	s := &models.Supplier{}
	err := r.db.Get(s, `SELECT * FROM suppliers WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SupplierRepo) Create(s *models.Supplier) error {
	s.ID = uuid.New().String()
	return r.db.QueryRowx(
		`INSERT INTO suppliers (id, name, email, phone, address) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		s.ID, s.Name, s.Email, s.Phone, s.Address,
	).Scan(&s.ID, &s.CreatedAt)
}

func (r *SupplierRepo) Update(id string, in *models.SupplierInput) error {
	_, err := r.db.Exec(
		`UPDATE suppliers SET name = $1, email = $2, phone = $3, address = $4 WHERE id = $5`,
		in.Name, in.Email, in.Phone, in.Address, id,
	)
	return err
}

func (r *SupplierRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM suppliers WHERE id = $1`, id)
	return err
}
