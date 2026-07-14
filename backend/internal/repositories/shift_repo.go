package repositories

import (
	"fmt"
	"pos/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ShiftRepo struct{ db *sqlx.DB }

func NewShiftRepo(db *sqlx.DB) *ShiftRepo { return &ShiftRepo{db: db} }

func (r *ShiftRepo) GetCurrent(cashierID string) (*models.Shift, error) {
	var s models.Shift
	err := r.db.Get(&s, `
		SELECT s.*, u.name AS cashier_name
		FROM shifts s JOIN users u ON u.id = s.cashier_id
		WHERE s.cashier_id = $1 AND s.status = 'open'
		ORDER BY s.created_at DESC LIMIT 1`, cashierID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShiftRepo) GetAnyOpen() (*models.Shift, error) {
	var s models.Shift
	err := r.db.Get(&s, `
		SELECT s.*, u.name AS cashier_name
		FROM shifts s JOIN users u ON u.id = s.cashier_id
		WHERE s.status = 'open'
		ORDER BY s.created_at DESC LIMIT 1`)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShiftRepo) Open(cashierID string, input *models.OpenShiftInput) (*models.Shift, error) {
	id := uuid.New().String()
	now := time.Now()
	_, err := r.db.Exec(`
		INSERT INTO shifts (id, cashier_id, opening_float, opening_time, status, notes)
		VALUES ($1,$2,$3,$4,'open',$5)`,
		id, cashierID, input.OpeningFloat, now, input.Notes)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *ShiftRepo) Close(id string, cashierID string, input *models.CloseShiftInput) (*models.Shift, error) {
	var s models.Shift
	err := r.db.Get(&s, `SELECT * FROM shifts WHERE id=$1 AND cashier_id=$2`, id, cashierID)
	if err != nil {
		return nil, fmt.Errorf("shift not found")
	}
	if s.Status != "open" {
		return nil, fmt.Errorf("shift is already closed")
	}

	expectedCash := s.OpeningFloat + s.CashSales
	variance := input.ActualCash - expectedCash
	now := time.Now()

	_, err = r.db.Exec(`
		UPDATE shifts SET
			status='closed', closing_time=$1, actual_cash=$2,
			expected_cash=$3, variance=$4, notes=CASE WHEN $5='' THEN notes ELSE $5 END
		WHERE id=$6`,
		now, input.ActualCash, expectedCash, variance, input.Notes, id)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *ShiftRepo) FindByID(id string) (*models.Shift, error) {
	var s models.Shift
	err := r.db.Get(&s, `
		SELECT s.*, u.name AS cashier_name
		FROM shifts s JOIN users u ON u.id = s.cashier_id
		WHERE s.id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShiftRepo) AddSaleTotals(cashierID string, paymentMethod string, amount float64) {
	colMap := map[string]string{
		"cash":  "cash_sales",
		"mpesa": "mpesa_sales",
		"card":  "card_sales",
	}
	col, ok := colMap[paymentMethod]
	if !ok {
		col = "other_sales"
	}
	r.db.Exec(fmt.Sprintf(`
		UPDATE shifts SET %s=%s+$1, total_sales=total_sales+$1, transaction_count=transaction_count+1
		WHERE cashier_id=$2 AND status='open'`, col, col),
		amount, cashierID)
}

func (r *ShiftRepo) ForceClose(id string) {
	r.db.Exec(`UPDATE shifts SET status='closed', closing_time=NOW(), notes=CASE WHEN notes='' OR notes IS NULL THEN 'Auto-closed' ELSE notes || '; Auto-closed' END WHERE id=$1 AND status='open'`, id)
}

func (r *ShiftRepo) List(cashierID string, limit int) ([]models.Shift, error) {
	var shifts []models.Shift
	err := r.db.Select(&shifts, `
		SELECT s.*, u.name AS cashier_name
		FROM shifts s JOIN users u ON u.id = s.cashier_id
		WHERE ($1='' OR s.cashier_id=$1)
		ORDER BY s.created_at DESC LIMIT $2`, cashierID, limit)
	return shifts, err
}
