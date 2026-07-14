package handlers

import (
	"pos/internal/services"
	"pos/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	reports *services.ReportService
}

func NewReportHandler(reports *services.ReportService) *ReportHandler {
	return &ReportHandler{reports: reports}
}

func (h *ReportHandler) DailySales(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}
	rows, err := h.reports.DailySales(days)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, rows)
}

func (h *ReportHandler) MonthlySales(c *fiber.Ctx) error {
	months, _ := strconv.Atoi(c.Query("months", "12"))
	if months < 1 || months > 24 {
		months = 12
	}
	rows, err := h.reports.MonthlySales(months)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, rows)
}

func (h *ReportHandler) TopProducts(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}
	rows, err := h.reports.TopProducts(limit)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, rows)
}

func (h *ReportHandler) InventoryValue(c *fiber.Ctx) error {
	rows, err := h.reports.InventoryValue()
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, rows)
}
