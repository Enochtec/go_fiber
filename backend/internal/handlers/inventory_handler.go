package handlers

import (
	"pos/internal/cache"
	"pos/internal/models"
	"pos/internal/repositories"
	"pos/internal/services"
	"pos/internal/utils"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type InventoryHandler struct {
	inventory *repositories.InventoryRepo
	service   *services.InventoryService
	validate  *validator.Validate
	cache     *cache.Cache
}

func NewInventoryHandler(inventory *repositories.InventoryRepo, service *services.InventoryService, v *validator.Validate) *InventoryHandler {
	return &InventoryHandler{inventory: inventory, service: service, validate: v, cache: cache.New(30 * time.Second)}
}

func (h *InventoryHandler) Dashboard(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	if v, ok := h.cache.Get("dashboard:" + shopID); ok {
		return utils.OK(c, v)
	}
	stats, err := h.inventory.GetDashboardStats(shopID)
	if err != nil {
		return utils.Internal(c, err)
	}
	h.cache.Set("dashboard:"+shopID, stats)
	return utils.OK(c, stats)
}

func (h *InventoryHandler) ListAdjustments(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}

	adjs, total, err := h.inventory.ListAdjustments(shopID, page, limit)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Paginated(c, adjs, total, page, limit)
}

func (h *InventoryHandler) Adjust(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	var input models.StockAdjustmentInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	userID := c.Locals("userID").(string)
	adj, err := h.service.Adjust(shopID, userID, &input)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, adj)
}
