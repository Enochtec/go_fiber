package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"pos/internal/database"
	"pos/internal/handlers"
	"pos/internal/repositories"
	"pos/internal/routes"
	"pos/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer db.Close()

	validate := validator.New()

	userRepo := repositories.NewUserRepo(db)
	productRepo := repositories.NewProductRepo(db)
	saleRepo := repositories.NewSaleRepo(db)
	customerRepo := repositories.NewCustomerRepo(db)
	supplierRepo := repositories.NewSupplierRepo(db)
	purchaseRepo := repositories.NewPurchaseRepo(db)
	inventoryRepo := repositories.NewInventoryRepo(db)

	authSvc := services.NewAuthService(userRepo)
	saleSvc := services.NewSaleService(saleRepo, productRepo)
	purchaseSvc := services.NewPurchaseService(purchaseRepo, productRepo)
	inventorySvc := services.NewInventoryService(inventoryRepo, productRepo)
	reportSvc := services.NewReportService(db)

	h := &routes.Handlers{
		Auth:      handlers.NewAuthHandler(authSvc, userRepo, validate),
		User:      handlers.NewUserHandler(userRepo, authSvc, validate),
		Product:   handlers.NewProductHandler(productRepo, validate),
		Sale:      handlers.NewSaleHandler(saleRepo, saleSvc, validate),
		Customer:  handlers.NewCustomerHandler(customerRepo, validate),
		Supplier:  handlers.NewSupplierHandler(supplierRepo, validate),
		Purchase:  handlers.NewPurchaseHandler(purchaseRepo, purchaseSvc, validate),
		Inventory: handlers.NewInventoryHandler(inventoryRepo, inventorySvc, validate),
		Report:    handlers.NewReportHandler(reportSvc),
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"success": false, "error": err.Error()})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	routes.Setup(app, h)

	webDir := filepath.Join(".", "web")
	app.Use(func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/api") {
			return c.Status(404).JSON(fiber.Map{"success": false, "error": "Not found"})
		}
		if _, err := os.Stat(webDir); err != nil {
			return c.Next()
		}
		p := c.Path()
		if p == "/" {
			p = "/index.html"
		}
		file := webDir + p
		if _, err := os.Stat(file); err == nil {
			return c.SendFile(file)
		}
		return c.SendFile(webDir + "/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
