package inquiry_item

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/inquiry_item"
)

type UpdateParams struct {
	ID                string
	GoodsNameSnap     *string
	CategoryNameSnap  *string
	SpecNameSnap      *string
	UnitNameSnap      *string
	GuidePrice        *float64
	LastMonthAvgPrice *float64
	CurrentAvgPrice   *float64
	Sort              *int
}

type ListParams struct {
	InquiryID string
	GoodsID   *string
	CategoryID *string
	Page      int
	PageSize  int
}

type InquiryItemRepository interface {
	CreateInquiryItem(ctx context.Context, m *domain.InquiryItem) error
	GetInquiryItem(ctx context.Context, id string) (*domain.InquiryItem, error)
	ListInquiryItems(ctx context.Context, params ListParams) ([]domain.InquiryItem, int64, error)
	UpdateInquiryItem(ctx context.Context, params UpdateParams) error
	SoftDeleteInquiryItem(ctx context.Context, id string) error
	HardDeleteInquiryItem(ctx context.Context, id string) error
	BatchCreateInquiryItems(ctx context.Context, items []domain.InquiryItem) error
}

func NewRepository(db *gorm.DB) InquiryItemRepository { return &inquiryItemRepo{db: db} }