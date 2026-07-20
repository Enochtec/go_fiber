package handlers

import (
	"fmt"
	"pos/internal/repositories"
	"pos/internal/services"
	"pos/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type EmailHandler struct {
	sales    *repositories.SaleRepo
	shops    *repositories.ShopRepo
	emailSvc *services.EmailService
}

func NewEmailHandler(sales *repositories.SaleRepo, shops *repositories.ShopRepo, emailSvc *services.EmailService) *EmailHandler {
	return &EmailHandler{sales: sales, shops: shops, emailSvc: emailSvc}
}

// POST /api/email/receipt
func (h *EmailHandler) SendReceipt(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)

	var body struct {
		SaleID   string  `json:"sale_id"`
		Email    string  `json:"email"`
		Tendered float64 `json:"tendered"`
		Change   float64 `json:"change"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if body.SaleID == "" {
		return utils.BadRequest(c, "sale_id is required")
	}
	if body.Email == "" || !strings.Contains(body.Email, "@") {
		return utils.BadRequest(c, "valid email address is required")
	}

	// Fetch sale
	sale, err := h.sales.FindByID(shopID, body.SaleID)
	if err != nil {
		return utils.NotFound(c, "sale")
	}

	// Fetch shop + settings
	shop, err := h.shops.FindByID(shopID)
	if err != nil {
		return utils.Internal(c, err)
	}
	settings, _ := h.shops.FindSettingsByShopID(shopID)

	shopName := "Maestro POS"
	shopAddr := ""
	shopPhone := ""
	shopEmail := ""
	shopLogo := ""
	footer := "Thank you for your business!"
	currency := "KES"

	if shop != nil {
		if shop.Name != "" {
			shopName = shop.Name
		}
		shopAddr = shop.Address
		shopPhone = shop.Phone
		shopEmail = shop.Email
		shopLogo = shop.Logo
	}
	if settings != nil {
		if settings.ReceiptFooter != "" {
			footer = settings.ReceiptFooter
		}
		if settings.Currency != "" {
			currency = settings.Currency
		}
	}

	// Build items
	var items []services.ReceiptEmailItem
	for _, item := range sale.Items {
		name := "Item"
		if item.ProductName != nil {
			name = *item.ProductName
		}
		items = append(items, services.ReceiptEmailItem{
			Name:      name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Total:     item.Total,
		})
	}

	cashier := "—"
	if sale.CashierName != nil {
		cashier = *sale.CashierName
	}
	customer := ""
	if sale.CustomerName != nil {
		customer = *sale.CustomerName
	}

	receiptNo := strings.ToUpper(sale.ID[:10])

	data := services.ReceiptEmailData{
		ToEmail:       body.Email,
		ToName:        customer,
		ShopName:      shopName,
		ShopAddress:   shopAddr,
		ShopPhone:     shopPhone,
		ShopEmail:     shopEmail,
		ShopLogo:      shopLogo,
		ReceiptNo:     receiptNo,
		Date:          sale.CreatedAt.Format("02 Jan 2006"),
		Time:          sale.CreatedAt.Format("15:04"),
		Cashier:       cashier,
		Customer:      customer,
		PaymentMethod: string(sale.PaymentMethod),
		Items:         items,
		Subtotal:      sale.Subtotal,
		Discount:      sale.Discount,
		Tax:           sale.Tax,
		Total:         sale.Total,
		Tendered:      body.Tendered,
		Change:        body.Change,
		Footer:        footer,
		Currency:      currency,
	}

	if err := h.emailSvc.SendReceipt(data); err != nil {
		return utils.Internal(c, fmt.Errorf("failed to send email: %w", err))
	}

	return utils.OKMessage(c, "Receipt sent to "+body.Email)
}
