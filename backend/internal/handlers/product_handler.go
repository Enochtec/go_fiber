package handlers

import (
	"pos/internal/models"
	"pos/internal/repositories"
	"pos/internal/utils"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	products *repositories.ProductRepo
	validate *validator.Validate
}

func NewProductHandler(products *repositories.ProductRepo, v *validator.Validate) *ProductHandler {
	return &ProductHandler{products: products, validate: v}
}

func (h *ProductHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := models.ProductFilter{
		Search:     c.Query("search"),
		CategoryID: c.Query("category_id"),
		LowStock:   c.Query("low_stock") == "true",
		Page:       page,
		Limit:      limit,
	}

	products, total, err := h.products.List(filter)
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.Paginated(c, products, total, page, limit)
}

func (h *ProductHandler) GetByID(c *fiber.Ctx) error {
	product, err := h.products.FindByID(c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "product")
	}
	return utils.OK(c, product)
}

func (h *ProductHandler) GetByBarcode(c *fiber.Ctx) error {
	product, err := h.products.FindByBarcode(c.Params("barcode"))
	if err != nil {
		return utils.NotFound(c, "product")
	}
	return utils.OK(c, product)
}

func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var input models.ProductInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	product := &models.Product{
		Name:         input.Name,
		Barcode:      input.Barcode,
		SKU:          input.SKU,
		CategoryID:   input.CategoryID,
		BuyingPrice:  input.BuyingPrice,
		SellingPrice: input.SellingPrice,
		StockQty:     input.StockQty,
		ReorderLevel: input.ReorderLevel,
		ImageURL:     input.ImageURL,
		IsActive:     true,
	}

	if err := h.products.Create(product); err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, product)
}

func (h *ProductHandler) Update(c *fiber.Ctx) error {
	var input models.ProductInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	if err := h.products.Update(c.Params("id"), &input); err != nil {
		return utils.Internal(c, err)
	}

	product, err := h.products.FindByID(c.Params("id"))
	if err != nil {
		return utils.NotFound(c, "product")
	}
	return utils.OK(c, product)
}

func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	if err := h.products.Delete(c.Params("id")); err != nil {
		return utils.Internal(c, err)
	}
	return utils.OKMessage(c, "product deleted")
}

func (h *ProductHandler) ListCategories(c *fiber.Ctx) error {
	cats, err := h.products.ListCategories()
	if err != nil {
		return utils.Internal(c, err)
	}
	return utils.OK(c, cats)
}

func (h *ProductHandler) CreateCategory(c *fiber.Ctx) error {
	var input models.CategoryInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.validate.Struct(input); err != nil {
		return utils.BadRequest(c, err.Error())
	}

	cat := &models.Category{Name: input.Name, Description: input.Description}
	if err := h.products.CreateCategory(cat); err != nil {
		return utils.Internal(c, err)
	}
	return utils.Created(c, cat)
}

func (h *ProductHandler) UpdateCategory(c *fiber.Ctx) error {
	var input models.CategoryInput
	if err := c.BodyParser(&input); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if err := h.products.UpdateCategory(c.Params("id"), &input); err != nil {
		return utils.Internal(c, err)
	}
	return utils.OKMessage(c, "category updated")
}

func (h *ProductHandler) DeleteCategory(c *fiber.Ctx) error {
	if err := h.products.DeleteCategory(c.Params("id")); err != nil {
		return utils.Internal(c, err)
	}
	return utils.OKMessage(c, "category deleted")
}
