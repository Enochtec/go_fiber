package services

import (
	"pos/internal/models"
	"pos/internal/repositories"
)

type PurchaseService struct {
	purchases *repositories.PurchaseRepo
	products  *repositories.ProductRepo
}

func NewPurchaseService(purchases *repositories.PurchaseRepo, products *repositories.ProductRepo) *PurchaseService {
	return &PurchaseService{purchases: purchases, products: products}
}

func (s *PurchaseService) Create(userID string, in *models.CreatePurchaseInput) (*models.Purchase, error) {
	status := models.PurchaseReceived
	if in.Status != "" {
		status = in.Status
	}

	var total float64
	for _, item := range in.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}

	purchase := &models.Purchase{
		SupplierID: in.SupplierID,
		UserID:     userID,
		Total:      total,
		Status:     status,
		Note:       in.Note,
	}

	db := s.purchases.DB()
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.purchases.Create(tx, purchase); err != nil {
		return nil, err
	}

	for _, inp := range in.Items {
		item := &models.PurchaseItem{
			PurchaseID: purchase.ID,
			ProductID:  inp.ProductID,
			Quantity:   inp.Quantity,
			UnitPrice:  inp.UnitPrice,
			Total:      inp.UnitPrice * float64(inp.Quantity),
		}
		if err := s.purchases.CreateItem(tx, item); err != nil {
			return nil, err
		}

		if status == models.PurchaseReceived {
			if err := s.products.UpdateStock(tx, inp.ProductID, inp.Quantity); err != nil {
				return nil, err
			}
		}

		purchase.Items = append(purchase.Items, *item)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return purchase, nil
}
