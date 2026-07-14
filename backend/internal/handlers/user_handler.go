package handlers

import (
	"pos/internal/models"
	"pos/internal/repositories"
	"pos/internal/services"
	"pos/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	users    *repositories.UserRepo
	auth     *services.AuthService
	validate *validator.Validate
}

func NewUserHandler(users *repositories.UserRepo, auth *services.AuthService, v *validator.Validate) *UserHandler {
	return &UserHandler{users: users, auth: auth, validate: v}
}

func (h *UserHandler) List(c *fiber.Ctx) error {
	users, err := h.users.List()
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, users)
}

func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	user, err := h.users.FindByID(c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "user")
	}
	return utils.OK(c, user)
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var input models.CreateUserInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	hash, err := h.auth.HashPassword(input.Password)
	if err != nil {
		return utils.Internal(c, err)
	}

	user := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hash,
		Role:     input.Role,
		IsActive: true,
	}

	if err := h.users.Create(user); err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, user)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	var input models.UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	fields := map[string]interface{}{
		"name":      nilIfEmpty(input.Name),
		"email":     nilIfEmpty(input.Email),
		"role":      nilIfEmpty(string(input.Role)),
		"is_active": input.IsActive,
		"password":  "",
	}

	if input.Password != "" {
		hash, err := h.auth.HashPassword(input.Password)
		if err != nil {
			return utils.Internal(c, err)
		}
		fields["password"] = hash
	}

	if err := h.users.Update(c.Params("id"), fields); err != nil {
		return utils.Internal(c, err)
	}

	user, err := h.users.FindByID(c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "user")
	}
	return utils.OK(c, user)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	if err := h.users.Delete(c.Params("id")); err != nil {
		return utils.Internal(c, err)
	}
	return utils.OKMessage(c, "user deactivated")
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
