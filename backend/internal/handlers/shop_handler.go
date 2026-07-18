package handlers

import (
	"pos/internal/repositories"
	"pos/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type ShopHandler struct {
	shops *repositories.ShopRepo
}

func NewShopHandler(shops *repositories.ShopRepo) *ShopHandler {
	return &ShopHandler{shops: shops}
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
