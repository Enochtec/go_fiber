package handlers

import (
	"pos/internal/models"
	"pos/internal/repositories"
	"pos/internal/services"
	"pos/internal/utils"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type PurchaseHandler struct {
	purchases *repositories.PurchaseRepo
	service   *services.PurchaseService
	validate  *validator.Validate
}

func NewPurchaseHandler(purchases *repositories.PurchaseRepo, service *services.PurchaseService, v *validator.Validate) *PurchaseHandler {
	return &PurchaseHandler{purchases: purchases, service: service, validate: v}
}

func (h *PurchaseHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}

	purchases, total, err := h.purchases.List(page, limit)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Paginated(c, purchases, total, page, limit)
}

func (h *PurchaseHandler) GetByID(c *fiber.Ctx) error {
	purchase, err := h.purchases.FindByID(c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "purchase")
	}
	return utils.OK(c, purchase)
}

func (h *PurchaseHandler) Create(c *fiber.Ctx) error {
	var input models.CreatePurchaseInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	userID := c.Locals("userID").(string)
	purchase, err := h.service.Create(userID, &input)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, purchase)
}
