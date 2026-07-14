package handlers

import (
	"pos/internal/models"
	"pos/internal/repositories"
	"pos/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ShiftHandler struct {
	shifts   *repositories.ShiftRepo
	validate *validator.Validate
}

func NewShiftHandler(shifts *repositories.ShiftRepo, v *validator.Validate) *ShiftHandler {
	return &ShiftHandler{shifts: shifts, validate: v}
}

func (h *ShiftHandler) Current(c *fiber.Ctx) error {
	cashierID := c.Locals("userID").(string)
	shift, err := h.shifts.GetCurrent(cashierID)
	if err != nil {
		return c.JSON(fiber.Map{"data": nil, "open": false})
	}
	return c.JSON(fiber.Map{"data": shift, "open": true})
}

func (h *ShiftHandler) Open(c *fiber.Ctx) error {
	cashierID := c.Locals("userID").(string)

	existing, _ := h.shifts.GetCurrent(cashierID)
	if existing != nil {
		h.shifts.ForceClose(existing.ID)
	}

	var input models.OpenShiftInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid body")
	}

	shift, err := h.shifts.Open(cashierID, &input)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, shift)
}

func (h *ShiftHandler) Close(c *fiber.Ctx) error {
	cashierID := c.Locals("userID").(string)
	id := c.Params("id")

	var input models.CloseShiftInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid body")
	}

	shift, err := h.shifts.Close(id, cashierID, &input)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}
	return utils.OK(c, shift)
}

func (h *ShiftHandler) List(c *fiber.Ctx) error {
	cashierID := c.Locals("userID").(string)
	role := c.Locals("role").(string)

	filterID := cashierID
	if role == "admin" || role == "manager" {
		filterID = c.Query("cashier_id", "")
	}

	shifts, err := h.shifts.List(filterID, 50)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, shifts)
}
