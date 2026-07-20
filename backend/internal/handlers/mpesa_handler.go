package handlers

import (
	"pos/internal/services"
	"pos/internal/utils"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/gofiber/fiber/v2"
)

type MpesaHandler struct {
	db      *sqlx.DB
	service *services.MpesaService
}

func NewMpesaHandler(db *sqlx.DB, svc *services.MpesaService) *MpesaHandler {
	return &MpesaHandler{db: db, service: svc}
}

// POST /api/mpesa/stk-push
// Initiates an STK push and stores the pending transaction.
func (h *MpesaHandler) STKPush(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)

	var body struct {
		Phone     string  `json:"phone"`
		Amount    float64 `json:"amount"`
		Reference string  `json:"reference"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.BadRequest(c, "invalid request body")
	}
	if body.Phone == "" {
		return utils.BadRequest(c, "phone is required")
	}
	if body.Amount <= 0 {
		return utils.BadRequest(c, "amount must be greater than 0")
	}

	result, err := h.service.InitiateSTKPush(services.STKPushInput{
		ShopID:    shopID,
		Phone:     body.Phone,
		Amount:    body.Amount,
		Reference: body.Reference,
	})
	if err != nil {
		return utils.BadRequest(c, err.Error())
	}

	// Store pending transaction
	id := uuid.New().String()
	_, err = h.db.Exec(`
		INSERT INTO mpesa_transactions
		  (id, shop_id, merchant_request_id, checkout_request_id, phone, amount, account_reference, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,'pending')`,
		id, shopID, result.MerchantRequestID, result.CheckoutRequestID,
		body.Phone, body.Amount, body.Reference,
	)
	if err != nil {
		return utils.Internal(c, err)
	}

	return utils.OK(c, fiber.Map{
		"checkout_request_id": result.CheckoutRequestID,
		"customer_message":    result.CustomerMessage,
	})
}

// POST /api/mpesa/callback   (no auth — called by Safaricom)
func (h *MpesaHandler) Callback(c *fiber.Ctx) error {
	var body services.MpesaCallbackBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"ResultCode": 1, "ResultDesc": "Bad request"})
	}

	res := h.service.ParseCallback(body)

	status := "failed"
	if res.ResultCode == 0 {
		status = "completed"
	} else if res.ResultCode == 1032 {
		status = "cancelled"
	}

	_, _ = h.db.Exec(`
		UPDATE mpesa_transactions SET
			status       = $1,
			mpesa_receipt = $2,
			result_code   = $3,
			result_desc   = $4,
			updated_at    = $5
		WHERE checkout_request_id = $6`,
		status, res.MpesaReceipt,
		res.ResultCode, res.ResultDesc,
		time.Now(),
		res.CheckoutRequestID,
	)

	return c.JSON(fiber.Map{"ResultCode": 0, "ResultDesc": "Accepted"})
}

// GET /api/mpesa/status/:checkoutRequestID
func (h *MpesaHandler) Status(c *fiber.Ctx) error {
	shopID := c.Locals("shopID").(string)
	crID   := c.Params("checkoutRequestID")

	var tx struct {
		Status       string  `db:"status"        json:"status"`
		MpesaReceipt *string `db:"mpesa_receipt" json:"mpesa_receipt"`
		ResultCode   *string `db:"result_code"   json:"result_code"`
		ResultDesc   *string `db:"result_desc"   json:"result_desc"`
		Phone        string  `db:"phone"         json:"phone"`
		Amount       float64 `db:"amount"        json:"amount"`
	}
	err := h.db.Get(&tx, `
		SELECT status, mpesa_receipt, result_code, result_desc, phone, amount
		FROM   mpesa_transactions
		WHERE  checkout_request_id = $1 AND shop_id = $2`,
		crID, shopID,
	)
	if err != nil {
		return utils.NotFound(c, "mpesa transaction")
	}
	return utils.OK(c, tx)
}
