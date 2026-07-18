package services

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"pos/internal/middleware"
	"pos/internal/models"
	"pos/internal/repositories"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users *repositories.UserRepo
}

func NewAuthService(users *repositories.UserRepo) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
	user, err := s.users.FindByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[auth] login failed: user not found: %s", email)
			return nil, "", errors.New("invalid credentials")
		}
		log.Printf("[auth] login error for %s: %v", email, err)
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Printf("[auth] login failed: password mismatch for %s", email)
		return nil, "", errors.New("invalid credentials")
	}

	log.Printf("[auth] login success: %s (%s)", email, user.ID[:8])

	token, err := generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func generateToken(user *models.User) (string, error) {
	hours := 24
	if h, err := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS")); err == nil {
		hours = h
	}

	shopID := ""
	if user.ShopID.Valid {
		shopID = user.ShopID.String
	}

	claims := &middleware.Claims{
		UserID: user.ID,
		Role:   string(user.Role),
		ShopID: shopID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(hours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
