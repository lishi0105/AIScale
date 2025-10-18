package supplier

import (
	"context"
	"time"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/supplier"
)

type ListParams struct {
	Keyword  string
	OrgID    string
	Status   *int
	Page     int
	PageSize int
}

type UpdateParams struct {
	ID              string
	Name            *string
	Code            *string
	Pinyin          *string
	Sort            *int
	Status          *int
	Description     *string
	FloatRatio      *float64
	StartTime       *time.Time
	EndTime         *time.Time
	UpdateCode      bool
	UpdatePinyin    bool
	UpdateSort      bool
	UpdateStartTime bool
	UpdateEndTime   bool
}

type SupplierRepository interface {
	CreateSupplier(ctx context.Context, m *domain.Supplier) error
	GetSupplier(ctx context.Context, id string) (*domain.Supplier, error)
	ListSuppliers(ctx context.Context, params ListParams) ([]domain.Supplier, int64, error)
	UpdateSupplier(ctx context.Context, params UpdateParams) error
	SoftDeleteSupplier(ctx context.Context, id string) error
	HardDeleteSupplier(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) SupplierRepository { return &supplierRepo{db: db} }
