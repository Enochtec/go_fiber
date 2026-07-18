package services

import (
	"pos/internal/models"
	"pos/internal/repositories"
)

type InventoryService struct {
	inventory *repositories.InventoryRepo
	products  *repositories.ProductRepo
}

func NewInventoryService(inventory *repositories.InventoryRepo, products *repositories.ProductRepo) *InventoryService {
	return &InventoryService{inventory: inventory, products: products}
}

func (s *InventoryService) Adjust(shopID string, userID string, in *models.StockAdjustmentInput) (*models.StockAdjustment, error) {
	adj := &models.StockAdjustment{
		ProductID: in.ProductID,
		UserID:    userID,
		Quantity:  in.Quantity,
		Reason:    in.Reason,
	}

	db := s.inventory.DB()
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.products.UpdateStock(tx, in.ProductID, in.Quantity); err != nil {
		return nil, err
	}

	if err := s.inventory.CreateAdjustment(tx, shopID, adj); err != nil {
		return nil, err
	}

	return adj, tx.Commit()
}
