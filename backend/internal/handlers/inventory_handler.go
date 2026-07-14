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

type InventoryHandler struct {
	inventory *repositories.InventoryRepo
	service   *services.InventoryService
	validate  *validator.Validate
}

func NewInventoryHandler(inventory *repositories.InventoryRepo, service *services.InventoryService, v *validator.Validate) *InventoryHandler {
	return &InventoryHandler{inventory: inventory, service: service, validate: v}
}

func (h *InventoryHandler) Dashboard(c *fiber.Ctx) error {
	stats, err := h.inventory.GetDashboardStats()
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, stats)
}

func (h *InventoryHandler) ListAdjustments(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}

	adjs, total, err := h.inventory.ListAdjustments(page, limit)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Paginated(c, adjs, total, page, limit)
}

func (h *InventoryHandler) Adjust(c *fiber.Ctx) error {
	var input models.StockAdjustmentInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	userID := c.Locals("userID").(string)
	adj, err := h.service.Adjust(userID, &input)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, adj)
}
