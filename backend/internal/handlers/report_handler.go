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
	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}
	key := "daily_sales:" + strconv.Itoa(days)
	if v, ok := h.cache.Get(key); ok {
		return utils.OK(c, v)
	}
	rows, err := h.reports.DailySales(days)
	if err != nil {
		return utils.Internal(c, err)
	}
	h.cache.SetWithTTL(key, rows, 60*time.Second)
	return utils.OK(c, rows)
}

func (h *ReportHandler) MonthlySales(c *fiber.Ctx) error {
	months, _ := strconv.Atoi(c.Query("months", "12"))
	if months < 1 || months > 24 {
		months = 12
	}
	key := "monthly_sales:" + strconv.Itoa(months)
	if v, ok := h.cache.Get(key); ok {
		return utils.OK(c, v)
	}
	rows, err := h.reports.MonthlySales(months)
	if err != nil {
		return utils.Internal(c, err)
	}
	h.cache.SetWithTTL(key, rows, 60*time.Second)
	return utils.OK(c, rows)
}

func (h *ReportHandler) TopProducts(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}
	key := "top_products:" + strconv.Itoa(limit)
	if v, ok := h.cache.Get(key); ok {
		return utils.OK(c, v)
	}
	rows, err := h.reports.TopProducts(limit)
	if err != nil {
		return utils.Internal(c, err)
	}
	h.cache.SetWithTTL(key, rows, 60*time.Second)
	return utils.OK(c, rows)
}

func (h *ReportHandler) InventoryValue(c *fiber.Ctx) error {
	if v, ok := h.cache.Get("inventory_value"); ok {
		return utils.OK(c, v)
	}
	rows, err := h.reports.InventoryValue()
	if err != nil {
		return utils.Internal(c, err)
	}
	h.cache.SetWithTTL("inventory_value", rows, 60*time.Second)
	return utils.OK(c, rows)
}
