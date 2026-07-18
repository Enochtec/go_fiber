package repositories

import (
	"pos/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Get(user, `SELECT * FROM users WHERE email = $1 AND is_active = TRUE`, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) FindByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Get(user, `SELECT * FROM users WHERE username = $1 AND is_active = TRUE`, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) FindByPhone(phone string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Get(user, `SELECT * FROM users WHERE phone = $1 AND is_active = TRUE`, phone)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) FindByID(id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Get(user, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) List() ([]models.User, error) {
	var users []models.User
	err := r.db.Select(&users, `SELECT * FROM users ORDER BY created_at DESC`)
	return users, err
}

func (r *UserRepo) Create(u *models.User) error {
	u.ID = uuid.New().String()
	return r.db.QueryRowx(
		`INSERT INTO users (id, shop_id, name, username, email, phone, password, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		u.ID, u.ShopID, u.Name, u.Username, u.Email, u.Phone, u.Password, u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) Update(id string, fields map[string]interface{}) error {
	fields["id"] = id
	_, err := r.db.NamedExec(
		`UPDATE users SET
			name = COALESCE(:name, name),
			email = COALESCE(:email, email),
			password = COALESCE(NULLIF(:password, ''), password),
			role = COALESCE(NULLIF(:role, ''), role),
			is_active = COALESCE(:is_active, is_active),
			updated_at = NOW()
		WHERE id = :id`,
		fields,
	)
	return err
}

func (r *UserRepo) Delete(id string) error {
	_, err := r.db.Exec(`UPDATE users SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
	return err
}
