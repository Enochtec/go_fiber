package repositories

import (
	"fmt"
	"pos/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CustomerRepo struct {
	db *sqlx.DB
}

func NewCustomerRepo(db *sqlx.DB) *CustomerRepo {
	return &CustomerRepo{db: db}
}

func (r *CustomerRepo) List(search string, page, limit int) ([]models.Customer, int, error) {
	var total int
	args := []interface{}{}
	where := "1=1"

	if search != "" {
		where = "name LIKE $1 OR phone LIKE $1 OR email LIKE $1"
		args = append(args, "%"+search+"%")
	}

	err := r.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM customers WHERE %s", where), args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	args = append(args, limit, offset)
	i := len(args) - 1

	var customers []models.Customer
	err = r.db.Select(&customers, fmt.Sprintf(
		`SELECT * FROM customers WHERE %s ORDER BY name ASC LIMIT $%d OFFSET $%d`,
		where, i, i+1,
	), args...)
	return customers, total, err
}

func (r *CustomerRepo) FindByID(id string) (*models.Customer, error) {
	c := &models.Customer{}
	err := r.db.Get(c, `SELECT * FROM customers WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CustomerRepo) Create(c *models.Customer) error {
	c.ID = uuid.New().String()
	return r.db.QueryRowx(
		`INSERT INTO customers (id, name, email, phone, address) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		c.ID, c.Name, c.Email, c.Phone, c.Address,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *CustomerRepo) Update(id string, in *models.CustomerInput) error {
	_, err := r.db.Exec(
		`UPDATE customers SET name = $1, email = $2, phone = $3, address = $4 WHERE id = $5`,
		in.Name, in.Email, in.Phone, in.Address, id,
	)
	return err
}

func (r *CustomerRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM customers WHERE id = $1`, id)
	return err
}
