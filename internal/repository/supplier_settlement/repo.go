package supplier_settlement

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/supplier_settlement"
)

type UpdateParams struct {
	ID               string
	InquiryID        *string
	ItemID           *string
	SupplierID       *string
	SupplierNameSnap *string
	FloatRatioSnap   *float64
	SettlementPrice  *float64
	UpdateSupplierID bool
}

type SupplierSettlementRepository interface {
	CreateSupplierSettlement(ctx context.Context, m *domain.SupplierSettlement) error
	GetSupplierSettlement(ctx context.Context, id string) (*domain.SupplierSettlement, error)
	ListSupplierSettlements(ctx context.Context, inquiryID *string, itemID *string, page, pageSize int) ([]domain.SupplierSettlement, int64, error)
	UpdateSupplierSettlement(ctx context.Context, params UpdateParams) error
	DeleteSupplierSettlement(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) SupplierSettlementRepository {
	return &supplierSettlementRepo{db: db}
}
