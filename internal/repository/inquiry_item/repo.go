package inquiry_item

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/inquiry_item"
)

type UpdateParams struct {
	ID                string
	InquiryID         *string
	GoodsID           *string
	CategoryID        *string
	SpecID            *string
	UnitID            *string
	GoodsNameSnap     *string
	CategoryNameSnap  *string
	SpecNameSnap      *string
	UnitNameSnap      *string
	GuidePrice        *float64
	LastMonthAvgPrice *float64
	CurrentAvgPrice   *float64
	Sort              *int
	UpdateSpecID      bool
	UpdateUnitID      bool
	UpdateSpecName    bool
	UpdateUnitName    bool
	UpdateGuidePrice  bool
	UpdateLastMonth   bool
	UpdateCurrentAvg  bool
}

type InquiryItemRepository interface {
	CreateInquiryItem(ctx context.Context, m *domain.InquiryItem) error
	GetInquiryItem(ctx context.Context, id string) (*domain.InquiryItem, error)
	ListInquiryItems(ctx context.Context, inquiryID string, categoryID *string, page, pageSize int) ([]domain.InquiryItem, int64, error)
	UpdateInquiryItem(ctx context.Context, params UpdateParams) error
	SoftDeleteInquiryItem(ctx context.Context, id string) error
	HardDeleteInquiryItem(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) InquiryItemRepository { return &inquiryItemRepo{db: db} }
