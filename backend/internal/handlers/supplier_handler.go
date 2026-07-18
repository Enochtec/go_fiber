package handlers

import (
	"pos/internal/models"
	"pos/internal/repositories"
	"pos/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type SupplierHandler struct {
	suppliers *repositories.SupplierRepo
	validate  *validator.Validate
}

func NewSupplierHandler(suppliers *repositories.SupplierRepo, v *validator.Validate) *SupplierHandler {
	return &SupplierHandler{suppliers: suppliers, validate: v}
}

func (h *SupplierHandler) List(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	suppliers, err := h.suppliers.List(shopID)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, suppliers)
}

func (h *SupplierHandler) GetByID(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	supplier, err := h.suppliers.FindByID(shopID, c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "supplier")
	}
	return utils.OK(c, supplier)
}

func (h *SupplierHandler) Create(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	var input models.SupplierInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	supplier := &models.Supplier{
		Name:    input.Name,
		Email:   input.Email,
		Phone:   input.Phone,
		Address: input.Address,
	}

	if err := h.suppliers.Create(shopID, supplier); err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, supplier)
}

func (h *SupplierHandler) Update(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	var input models.SupplierInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.suppliers.Update(shopID, c.Params("id"), &input); err != nil {
		return utils.Internal(c, err)
	}

	supplier, err := h.suppliers.FindByID(shopID, c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "supplier")
	}
	return utils.OK(c, supplier)
}

func (h *SupplierHandler) Delete(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	if err := h.suppliers.Delete(shopID, c.Params("id")); err != nil {
		return utils.Internal(c, err)
	}
	return utils.OKMessage(c, "supplier deleted")
}
