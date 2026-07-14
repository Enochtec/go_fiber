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

type SaleHandler struct {
	sales    *repositories.SaleRepo
	service  *services.SaleService
	validate *validator.Validate
}

func NewSaleHandler(sales *repositories.SaleRepo, service *services.SaleService, v *validator.Validate) *SaleHandler {
	return &SaleHandler{sales: sales, service: service, validate: v}
}

func (h *SaleHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}

	filter := models.SaleFilter{
		Status:    c.Query("status"),
		CashierID: c.Query("cashier_id"),
		DateFrom:  c.Query("date_from"),
		DateTo:    c.Query("date_to"),
		Page:      page,
		Limit:     limit,
	}

	sales, total, err := h.sales.List(filter)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Paginated(c, sales, total, page, limit)
}

func (h *SaleHandler) GetByID(c *fiber.Ctx) error {
	sale, err := h.sales.FindByID(c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "sale")
	}
	return utils.OK(c, sale)
}

func (h *SaleHandler) Create(c *fiber.Ctx) error {
	var input models.CreateSaleInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	cashierID := c.Locals("userID").(string)
	sale, err := h.service.Create(cashierID, &input)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, sale)
}

func (h *SaleHandler) Void(c *fiber.Ctx) error {
	if err := h.service.Void(c.Params("id")); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	return utils.OKMessage(c, "sale voided")
}

func (h *SaleHandler) Hold(c *fiber.Ctx) error {
	if err := h.sales.UpdateStatus(c.Params("id"), models.SaleHeld); err != nil {
		return utils.Internal(c, err)
	}
	return utils.OKMessage(c, "sale held")
}
