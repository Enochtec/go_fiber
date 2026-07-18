package handlers

import (
	"pos/internal/models"
	"pos/internal/services"
	"pos/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type RegisterHandler struct {
	registration *services.RegistrationService
	validate     *validator.Validate
}

func NewRegisterHandler(registration *services.RegistrationService, v *validator.Validate) *RegisterHandler {
	return &RegisterHandler{registration: registration, validate: v}
}

func (h *RegisterHandler) Register(c *fiber.Ctx) error {
	var input models.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}

	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	result, err := h.registration.Register(&input)
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	return utils.Created(c, result)
}
