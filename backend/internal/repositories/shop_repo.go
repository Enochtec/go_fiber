package repositories

import (
	"pos/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ShopRepo struct {
	db *sqlx.DB
}

func NewShopRepo(db *sqlx.DB) *ShopRepo {
	return &ShopRepo{db: db}
}

func (r *ShopRepo) Create(shop *models.Shop) error {
	shop.ID = uuid.New().String()
	return r.db.QueryRowx(`
		INSERT INTO shops (id, name, business_type, email, phone, address, country, county, town, currency, timezone, logo)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`,
		shop.ID, shop.Name, shop.BusinessType, shop.Email, shop.Phone,
		shop.Address, shop.Country, shop.County, shop.Town,
		shop.Currency, shop.Timezone, shop.Logo,
	).Scan(&shop.ID, &shop.CreatedAt, &shop.UpdatedAt)
}

func (r *ShopRepo) FindByID(id string) (*models.Shop, error) {
	shop := &models.Shop{}
	err := r.db.Get(shop, `SELECT * FROM shops WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return shop, nil
}

func (r *ShopRepo) FindByEmail(email string) (*models.Shop, error) {
	shop := &models.Shop{}
	err := r.db.Get(shop, `SELECT * FROM shops WHERE email = $1`, email)
	if err != nil {
		return nil, err
	}
	return shop, nil
}

func (r *ShopRepo) CreateSettings(s *models.ShopSettings) error {
	s.ID = uuid.New().String()
	return r.db.QueryRowx(`
		INSERT INTO shop_settings (id, shop_id, tax_rate, receipt_footer, invoice_prefix, default_payment_method, low_stock_threshold, enable_notifications, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`,
		s.ID, s.ShopID, s.TaxRate, s.ReceiptFooter, s.InvoicePrefix,
		s.DefaultPaymentMethod, s.LowStockThreshold, s.EnableNotifications, s.Currency,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *ShopRepo) FindSettingsByShopID(shopID string) (*models.ShopSettings, error) {
	s := &models.ShopSettings{}
	err := r.db.Get(s, `SELECT * FROM shop_settings WHERE shop_id = $1`, shopID)
	if err != nil {
		return nil, err
	}
	return s, nil
}
