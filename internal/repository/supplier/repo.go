package supplier

import (
	"context"
	"time"

	"gorm.io/gorm"
	supplier "hdzk.cn/foodapp/internal/domain/supplier"
)

type SupplierRepository interface {
	CreateSupplier(ctx context.Context, m *supplier.Supplier) error
	GetSupplier(ctx context.Context, id string) (*supplier.Supplier, error)
	ListSuppliers(ctx context.Context, keyword string, orgID *string, status *int, page, pageSize int) ([]supplier.Supplier, int64, error)
	UpdateSupplier(ctx context.Context, id string, name string, code *string, pinyin *string, sort *int, status *int, description *string, floatRatio *float64, orgID *string, startTime *time.Time, endTime *time.Time, updateCode bool, updatePinyin bool, updateSort bool, updateStatus bool, updateDescription bool, updateFloatRatio bool, updateOrgID bool, updateStartTime bool, updateEndTime bool) error
	SoftDeleteSupplier(ctx context.Context, id string) error
	HardDeleteSupplier(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) SupplierRepository { return &supplierRepo{db: db} }
