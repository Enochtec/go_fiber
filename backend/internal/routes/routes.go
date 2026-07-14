package routes

import (
	"pos/internal/handlers"
	"pos/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	Auth      *handlers.AuthHandler
	User      *handlers.UserHandler
	Product   *handlers.ProductHandler
	Sale      *handlers.SaleHandler
	Customer  *handlers.CustomerHandler
	Supplier  *handlers.SupplierHandler
	Purchase  *handlers.PurchaseHandler
	Inventory *handlers.InventoryHandler
	Report    *handlers.ReportHandler
	Shift     *handlers.ShiftHandler
}

func Setup(app *fiber.App, h *Handlers) {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "data": "ok"})
	})

	api.Post("/auth/login", h.Auth.Login)

	protected := api.Group("", middleware.Auth())

	protected.Get("/auth/me", h.Auth.Me)

	protected.Get("/dashboard", h.Inventory.Dashboard)

	users := protected.Group("/users", middleware.RequireRole("admin", "manager"))
	users.Get("/", h.User.List)
	users.Get("/:id", h.User.GetByID)
	users.Post("/", middleware.RequireRole("admin"), h.User.Create)
	users.Put("/:id", middleware.RequireRole("admin"), h.User.Update)
	users.Delete("/:id", middleware.RequireRole("admin"), h.User.Delete)

	cats := protected.Group("/categories")
	cats.Get("/", h.Product.ListCategories)
	cats.Post("/", middleware.RequireRole("admin", "manager"), h.Product.CreateCategory)
	cats.Put("/:id", middleware.RequireRole("admin", "manager"), h.Product.UpdateCategory)
	cats.Delete("/:id", middleware.RequireRole("admin", "manager"), h.Product.DeleteCategory)

	products := protected.Group("/products")
	products.Get("/", h.Product.List)
	products.Get("/barcode/:barcode", h.Product.GetByBarcode)
	products.Get("/:id", h.Product.GetByID)
	products.Post("/", middleware.RequireRole("admin", "manager"), h.Product.Create)
	products.Put("/:id", middleware.RequireRole("admin", "manager"), h.Product.Update)
	products.Delete("/:id", middleware.RequireRole("admin", "manager"), h.Product.Delete)

	customers := protected.Group("/customers")
	customers.Get("/", h.Customer.List)
	customers.Get("/:id/stats", h.Customer.Stats)
	customers.Get("/:id/history", h.Customer.History)
	customers.Get("/:id", h.Customer.GetByID)
	customers.Post("/", h.Customer.Create)
	customers.Put("/:id", h.Customer.Update)
	customers.Delete("/:id", middleware.RequireRole("admin", "manager"), h.Customer.Delete)

	suppliers := protected.Group("/suppliers")
	suppliers.Get("/", h.Supplier.List)
	suppliers.Get("/:id", h.Supplier.GetByID)
	suppliers.Post("/", middleware.RequireRole("admin", "manager"), h.Supplier.Create)
	suppliers.Put("/:id", middleware.RequireRole("admin", "manager"), h.Supplier.Update)
	suppliers.Delete("/:id", middleware.RequireRole("admin", "manager"), h.Supplier.Delete)

	sales := protected.Group("/sales")
	sales.Get("/", h.Sale.List)
	sales.Get("/:id", h.Sale.GetByID)
	sales.Post("/", h.Sale.Create)
	sales.Put("/:id/void", middleware.RequireRole("admin", "manager"), h.Sale.Void)
	sales.Put("/:id/hold", h.Sale.Hold)

	purchases := protected.Group("/purchases", middleware.RequireRole("admin", "manager"))
	purchases.Get("/", h.Purchase.List)
	purchases.Get("/:id", h.Purchase.GetByID)
	purchases.Post("/", h.Purchase.Create)

	inventory := protected.Group("/inventory")
	inventory.Get("/adjustments", h.Inventory.ListAdjustments)
	inventory.Post("/adjust", middleware.RequireRole("admin", "manager"), h.Inventory.Adjust)

	reports := protected.Group("/reports", middleware.RequireRole("admin", "manager"))
	reports.Get("/sales/daily", h.Report.DailySales)
	reports.Get("/sales/monthly", h.Report.MonthlySales)
	reports.Get("/products/top", h.Report.TopProducts)
	reports.Get("/inventory/value", h.Report.InventoryValue)

	shifts := protected.Group("/shifts")
	shifts.Get("/current", h.Shift.Current)
	shifts.Get("/", h.Shift.List)
	shifts.Post("/open", h.Shift.Open)
	shifts.Post("/:id/close", h.Shift.Close)
}
