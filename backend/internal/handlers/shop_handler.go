package handlers

import (
	"pos/internal/models"
	"pos/internal/repositories"
	"pos/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ShopHandler struct {
	shops    *repositories.ShopRepo
	validate *validator.Validate
}

func NewShopHandler(shops *repositories.ShopRepo, validate *validator.Validate) *ShopHandler {
	return &ShopHandler{shops: shops, validate: validate}
}

func (h *ShopHandler) Info(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	shop, err := h.shops.FindByID(shopID)
	if err != nil {
		return utils.NotFound(c, "shop")
	}
	settings, err := h.shops.FindSettingsByShopID(shopID)
	if err != nil {
		settings = nil
	}
	return utils.OK(c, fiber.Map{"shop": shop, "settings": settings})
}

func (h *ShopHandler) UpdateName(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	var input models.ShopNameInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}
	if err := h.shops.UpdateName(shopID, input.Name); err != nil {
		return utils.Internal(c, err)
	}
	shop, err := h.shops.FindByID(shopID)
	if err != nil {
		return utils.NotFound(c, "shop")
	}
	return utils.OK(c, shop)
}
