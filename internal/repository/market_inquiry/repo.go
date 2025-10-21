package market_inquiry

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/market_inquiry"
)

type UpdateParams struct {
	ID             string
	MarketID       *string
	MarketNameSnap *string
	Price          *float64
}

type ListParams struct {
	InquiryID string
	ItemID    *string
	MarketID  *string
	Page      int
	PageSize  int
}

type MarketInquiryRepository interface {
	CreateMarketInquiry(ctx context.Context, m *domain.MarketInquiry) error
	GetMarketInquiry(ctx context.Context, id string) (*domain.MarketInquiry, error)
	ListMarketInquiries(ctx context.Context, params ListParams) ([]domain.MarketInquiry, int64, error)
	UpdateMarketInquiry(ctx context.Context, params UpdateParams) error
	SoftDeleteMarketInquiry(ctx context.Context, id string) error
	HardDeleteMarketInquiry(ctx context.Context, id string) error
	BatchCreateMarketInquiries(ctx context.Context, items []domain.MarketInquiry) error
}

func NewRepository(db *gorm.DB) MarketInquiryRepository { return &marketInquiryRepo{db: db} }