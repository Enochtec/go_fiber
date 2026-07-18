package main

import (
	"fmt"
	"log"
	"os"
	"pos/internal/database"
	"pos/internal/models"
	"pos/internal/repositories"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, using environment variables")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer db.Close()

	email := "admin@pos.local"
	password := "admin123"
	name := "Administrator"

	if len(os.Args) >= 3 {
		email = os.Args[1]
		password = os.Args[2]
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Hash error: %v", err)
	}

	repo := repositories.NewUserRepo(db)
	user := &models.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		Role:     models.RoleAdmin,
		IsActive: true,
	}

	if err := repo.Create("", user); err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Printf("Admin created:\n  Email:    %s\n  Password: %s\n", email, password)
}
