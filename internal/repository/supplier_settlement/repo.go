package supplier_settlement

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/supplier_settlement"
)

type UpdateParams struct {
	ID               string
	SupplierID       *string
	SupplierNameSnap *string
	FloatRatioSnap   *float64
	SettlementPrice  *float64
}

type ListParams struct {
	InquiryID  string
	ItemID     *string
	SupplierID *string
	Page       int
	PageSize   int
}

type SupplierSettlementRepository interface {
	CreateSupplierSettlement(ctx context.Context, m *domain.SupplierSettlement) error
	GetSupplierSettlement(ctx context.Context, id string) (*domain.SupplierSettlement, error)
	ListSupplierSettlements(ctx context.Context, params ListParams) ([]domain.SupplierSettlement, int64, error)
	UpdateSupplierSettlement(ctx context.Context, params UpdateParams) error
	SoftDeleteSupplierSettlement(ctx context.Context, id string) error
	HardDeleteSupplierSettlement(ctx context.Context, id string) error
	BatchCreateSupplierSettlements(ctx context.Context, settlements []domain.SupplierSettlement) error
	GetByItemAndSupplier(ctx context.Context, itemID, supplierName string) (*domain.SupplierSettlement, error)
}

func NewRepository(db *gorm.DB) SupplierSettlementRepository { return &supplierSettlementRepo{db: db} }