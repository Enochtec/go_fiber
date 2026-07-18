package services

import (
	"database/sql"
	"errors"
	"pos/internal/models"
	"pos/internal/repositories"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type RegistrationService struct {
	db     *sqlx.DB
	shops  *repositories.ShopRepo
	users  *repositories.UserRepo
	auth   *AuthService
}

func NewRegistrationService(db *sqlx.DB, shops *repositories.ShopRepo, users *repositories.UserRepo, auth *AuthService) *RegistrationService {
	return &RegistrationService{db: db, shops: shops, users: users, auth: auth}
}

func (s *RegistrationService) Register(input *models.RegisterInput) (*models.RegisterResult, error) {
	exists, err := s.users.FindByEmail(input.OwnerEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if exists != nil {
		return nil, errors.New("an account with this email already exists")
	}

	existingUser, err := s.users.FindByUsername(input.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("this username is already taken")
	}

	existingPhone, err := s.users.FindByPhone(input.OwnerPhone)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existingPhone != nil {
		return nil, errors.New("this phone number is already registered")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to secure password")
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, errors.New("failed to start transaction")
	}
	defer tx.Rollback()

	shop := &models.Shop{
		Name:         input.ShopName,
		BusinessType: input.BusinessType,
		Email:        input.BusinessEmail,
		Phone:        input.BusinessPhone,
		Address:      input.Address,
		Country:      input.Country,
		County:       input.County,
		Town:         input.Town,
		Currency:     input.Currency,
		Timezone:     input.Timezone,
		Logo:         input.ShopName[:1],
	}
	if err := s.shops.Create(shop); err != nil {
		return nil, errors.New("failed to create shop")
	}

	user := &models.User{
		ShopID:   sql.NullString{String: shop.ID, Valid: true},
		Name:     input.OwnerName,
		Username: input.Username,
		Email:    input.OwnerEmail,
		Phone:    input.OwnerPhone,
		Password: string(passwordHash),
		Role:     models.RoleAdmin,
		IsActive: true,
	}
	if err := s.users.Create(user); err != nil {
		return nil, errors.New("failed to create account")
	}

	settings := &models.ShopSettings{
		ShopID:               shop.ID,
		TaxRate:              0,
		ReceiptFooter:        "Thank you for your business!",
		InvoicePrefix:        "INV-",
		DefaultPaymentMethod: "cash",
		LowStockThreshold:    10,
		EnableNotifications:  true,
		Currency:             input.Currency,
	}
	if err := s.shops.CreateSettings(settings); err != nil {
		return nil, errors.New("failed to initialize shop settings")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("failed to complete registration")
	}

	token, err := generateToken(user)
	if err != nil {
		return nil, errors.New("failed to generate session")
	}

	return &models.RegisterResult{
		User:  user,
		Shop:  shop,
		Token: token,
	}, nil
}
