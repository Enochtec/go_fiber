package handlers

import (
	"pos/internal/models"
	"pos/internal/repositories"
	"pos/internal/utils"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type CustomerHandler struct {
	customers *repositories.CustomerRepo
	validate  *validator.Validate
}

func NewCustomerHandler(customers *repositories.CustomerRepo, v *validator.Validate) *CustomerHandler {
	return &CustomerHandler{customers: customers, validate: v}
}

func (h *CustomerHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}

	customers, total, err := h.customers.List(c.Query("search"), page, limit)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Paginated(c, customers, total, page, limit)
}

func (h *CustomerHandler) GetByID(c *fiber.Ctx) error {
	customer, err := h.customers.FindByID(c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "customer")
	}
	return utils.OK(c, customer)
}

func (h *CustomerHandler) Create(c *fiber.Ctx) error {
	var input models.CustomerInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	customer := &models.Customer{
		Name:    input.Name,
		Email:   input.Email,
		Phone:   input.Phone,
		Address: input.Address,
	}

	if err := h.customers.Create(customer); err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, customer)
}

func (h *CustomerHandler) Update(c *fiber.Ctx) error {
	var input models.CustomerInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.customers.Update(c.Params("id"), &input); err != nil {
		return utils.Internal(c, err)
	}

	customer, err := h.customers.FindByID(c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "customer")
	}
	return utils.OK(c, customer)
}

func (h *CustomerHandler) Delete(c *fiber.Ctx) error {
	if err := h.customers.Delete(c.Params("id")); err != nil {
		return utils.Internal(c, err)
	}
	return utils.OKMessage(c, "customer deleted")
}

func (h *CustomerHandler) Stats(c *fiber.Ctx) error {
	stats, err := h.customers.GetStats(c.Params("id"))
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, stats)
}

func (h *CustomerHandler) History(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	sales, err := h.customers.ListPurchaseHistory(c.Params("id"), limit)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, sales)
}
