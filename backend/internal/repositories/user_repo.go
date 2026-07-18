package repositories

import (
	"database/sql"
	"fmt"
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

// FindByEmail is global (used during login pre-shop context).
func (r *UserRepo) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Get(user, `SELECT * FROM users WHERE email = $1 AND is_active = TRUE`, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByUsername is global (used during registration pre-shop context).
func (r *UserRepo) FindByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Get(user, `SELECT * FROM users WHERE username = $1 AND is_active = TRUE`, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByPhone is global (used during registration pre-shop context).
func (r *UserRepo) FindByPhone(phone string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Get(user, `SELECT * FROM users WHERE phone = $1 AND is_active = TRUE`, phone)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID is shop-scoped. A user can only look up another user in the same shop.
func (r *UserRepo) FindByID(shopID, id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Get(user, `SELECT * FROM users WHERE id = $1 AND shop_id = $2`, id, shopID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// List is shop-scoped.
func (r *UserRepo) List(shopID string) ([]models.User, error) {
	var users []models.User
	err := r.db.Select(&users, `SELECT * FROM users WHERE shop_id = $1 ORDER BY created_at DESC`, shopID)
	return users, err
}

// Create is shop-scoped. shopID is set on the user record.
func (r *UserRepo) Create(shopID string, u *models.User) error {
	u.ID = uuid.New().String()
	var shopNull sql.NullString
	err := r.db.QueryRowx(
		`INSERT INTO users (id, shop_id, name, username, email, phone, password, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, shop_id, created_at, updated_at`,
		u.ID, shopID, u.Name, u.Username, u.Email, u.Phone, u.Password, u.Role,
	).Scan(&u.ID, &shopNull, &u.CreatedAt, &u.UpdatedAt)
	u.ShopID = shopNull
	return err
}

// Update is shop-scoped. Only users in the same shop can be updated.
func (r *UserRepo) Update(shopID, id string, fields map[string]interface{}) error {
	fields["id"] = id
	fields["shop_id"] = shopID
	_, err := r.db.NamedExec(
		`UPDATE users SET
			name = COALESCE(:name, name),
			email = COALESCE(:email, email),
			password = COALESCE(NULLIF(:password, ''), password),
			role = COALESCE(NULLIF(:role, ''), role),
			is_active = COALESCE(:is_active, is_active),
			updated_at = NOW()
		WHERE id = :id AND shop_id = :shop_id`,
		fields,
	)
	return err
}

// Delete is shop-scoped. Only users in the same shop can be deactivated.
func (r *UserRepo) Delete(shopID, id string) error {
	_, err := r.db.Exec(`UPDATE users SET is_active = FALSE, updated_at = NOW() WHERE id = $1 AND shop_id = $2`, id, shopID)
	return err
}

// ExistsInShop checks if a user (identified by ID) belongs to a shop.
func (r *UserRepo) ExistsInShop(shopID, userID string) (bool, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM users WHERE id = $1 AND shop_id = $2 AND is_active = TRUE`, userID, shopID)
	if err != nil {
		return false, fmt.Errorf("check user shop: %w", err)
	}
	return count > 0, nil
}
