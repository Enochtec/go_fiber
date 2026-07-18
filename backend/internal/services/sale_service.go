package services

import (
	"errors"
	"fmt"
	"pos/internal/models"
	"pos/internal/repositories"
)

type SaleService struct {
	sales    *repositories.SaleRepo
	products *repositories.ProductRepo
}

func NewSaleService(sales *repositories.SaleRepo, products *repositories.ProductRepo) *SaleService {
	return &SaleService{sales: sales, products: products}
}

func (s *SaleService) Create(shopID string, cashierID string, in *models.CreateSaleInput) (*models.Sale, error) {
	status := models.SaleCompleted
	if in.Status != "" {
		status = in.Status
	}

	var subtotal float64
	for _, item := range in.Items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}
	taxAmount := subtotal * (in.TaxRate / 100)
	total := subtotal - in.Discount + taxAmount

	sale := &models.Sale{
		CashierID:     cashierID,
		CustomerID:    in.CustomerID,
		Subtotal:      subtotal,
		Discount:      in.Discount,
		Tax:           taxAmount,
		Total:         total,
		PaymentMethod: in.PaymentMethod,
		Status:        status,
		Note:          in.Note,
	}

	db := s.sales.DB()
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.sales.Create(tx, shopID, sale); err != nil {
		return nil, err
	}

	for _, inp := range in.Items {
		if status == models.SaleCompleted {
			name, stock, err := s.products.GetStock(shopID, tx, inp.ProductID)
			if err != nil {
				return nil, fmt.Errorf("product lookup failed: %w", err)
			}
			if stock < inp.Quantity {
				return nil, fmt.Errorf("insufficient stock for '%s': have %d, need %d", name, stock, inp.Quantity)
			}
		}

		item := &models.SaleItem{
			SaleID:    sale.ID,
			ProductID: inp.ProductID,
			Quantity:  inp.Quantity,
			UnitPrice: inp.UnitPrice,
			Total:     inp.UnitPrice * float64(inp.Quantity),
		}
		if err := s.sales.CreateItem(tx, shopID, item); err != nil {
			return nil, err
		}

		if status == models.SaleCompleted {
			if err := s.products.UpdateStock(tx, inp.ProductID, -inp.Quantity); err != nil {
				return nil, err
			}
		}

		sale.Items = append(sale.Items, *item)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return sale, nil
}

func (s *SaleService) Void(shopID string, id string) error {
	sale, err := s.sales.FindByID(shopID, id)
	if err != nil {
		return err
	}
	if sale.Status == models.SaleVoided {
		return errors.New("sale already voided")
	}

	db := s.sales.DB()
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if sale.Status == models.SaleCompleted {
		for _, item := range sale.Items {
			if err := s.products.UpdateStock(tx, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
	}

	if err := s.sales.UpdateStatus(shopID, id, models.SaleVoided); err != nil {
		return err
	}

	return tx.Commit()
}
