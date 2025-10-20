package supplier

import (
	"context"
	"time"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/supplier"
)

type UpdateParams struct {
	ID                   string
	Name                 *string
	Code                 *string
	Pinyin               *string
	Sort                 *int
	Status               *int
	Description          *string
	FloatRatio           *float64
	ContactName          *string
	ContactPhone         *string
	ContactEmail         *string
	ContactAddress       *string
	StartTime            *time.Time
	EndTime              *time.Time
	UpdateCode           bool
	UpdatePinyin         bool
	UpdateSort           bool
	UpdateContactName    bool
	UpdateContactPhone   bool
	UpdateContactEmail   bool
	UpdateContactAddress bool
	UpdateStartTime      bool
	UpdateEndTime        bool
}

type SupplierRepository interface {
	CreateSupplier(ctx context.Context, m *domain.Supplier) error
	GetSupplier(ctx context.Context, id string) (*domain.Supplier, error)
	ListSuppliers(ctx context.Context, keyword string, orgID *string, status *int, contactName, contactPhone, contactEmail, contactAddress *string, page, pageSize int) ([]domain.Supplier, int64, error)
	UpdateSupplier(ctx context.Context, params UpdateParams) error
	SoftDeleteSupplier(ctx context.Context, id string) error
	HardDeleteSupplier(ctx context.Context, id string) error
	FindByName(ctx context.Context, name string, orgID string) (*domain.Supplier, error)
}

func NewRepository(db *gorm.DB) SupplierRepository { return &supplierRepo{db: db} }
