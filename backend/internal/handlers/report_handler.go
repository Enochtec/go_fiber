package handlers

import (
	"pos/internal/cache"
	"pos/internal/services"
	"pos/internal/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	reports *services.ReportService
	cache   *cache.Cache
}

func NewReportHandler(reports *services.ReportService) *ReportHandler {
	return &ReportHandler{reports: reports, cache: cache.New(60 * time.Second)}
}

func (h *ReportHandler) DailySales(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}
	key := "daily_sales:" + shopID + ":" + strconv.Itoa(days)
	if v, ok := h.cache.Get(key); ok {
		return utils.OK(c, v)
	}
	rows, err := h.reports.DailySales(shopID, days)
	if err != nil {
		return utils.Internal(c, err)
	}
	h.cache.SetWithTTL(key, rows, 60*time.Second)
	return utils.OK(c, rows)
}

func (h *ReportHandler) MonthlySales(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	months, _ := strconv.Atoi(c.Query("months", "12"))
	if months < 1 || months > 24 {
		months = 12
	}
	key := "monthly_sales:" + shopID + ":" + strconv.Itoa(months)
	if v, ok := h.cache.Get(key); ok {
		return utils.OK(c, v)
	}
	rows, err := h.reports.MonthlySales(shopID, months)
	if err != nil {
		return utils.Internal(c, err)
	}
	h.cache.SetWithTTL(key, rows, 60*time.Second)
	return utils.OK(c, rows)
}

func (h *ReportHandler) TopProducts(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}
	key := "top_products:" + shopID + ":" + strconv.Itoa(limit)
	if v, ok := h.cache.Get(key); ok {
		return utils.OK(c, v)
	}
	rows, err := h.reports.TopProducts(shopID, limit)
	if err != nil {
		return utils.Internal(c, err)
	}
	h.cache.SetWithTTL(key, rows, 60*time.Second)
	return utils.OK(c, rows)
}

func (h *ReportHandler) InventoryValue(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	key := "inventory_value:" + shopID
	if v, ok := h.cache.Get(key); ok {
		return utils.OK(c, v)
	}
	rows, err := h.reports.InventoryValue(shopID)
	if err != nil {
		return utils.Internal(c, err)
	}
	h.cache.SetWithTTL(key, rows, 60*time.Second)
	return utils.OK(c, rows)
}
