package handlers

import (
	"pos/internal/models"
	"pos/internal/repositories"
	"pos/internal/services"
	"pos/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	auth     *services.AuthService
	users    *repositories.UserRepo
	validate *validator.Validate
}

func NewAuthHandler(auth *services.AuthService, users *repositories.UserRepo, v *validator.Validate) *AuthHandler {
	return &AuthHandler{auth: auth, users: users, validate: v}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input models.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	user, token, err := h.auth.Login(input.Email, input.Password)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.OK(c, fiber.Map{"user": user, "token": token})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	shopID := c.Locals("shopID").(string)
	user, err := h.users.FindByID(shopID, userID)
	if err != nil {
		return utils.NotFound(c, "user")
	}
	return utils.OK(c, user)
}
