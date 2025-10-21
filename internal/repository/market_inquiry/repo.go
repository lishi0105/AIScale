package market_inquiry

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/market_inquiry"
)

type UpdateParams struct {
	ID             string
	InquiryID      *string
	ItemID         *string
	MarketID       *string
	MarketNameSnap *string
	Price          *float64
	UpdateMarketID bool
}

type MarketInquiryRepository interface {
	CreateMarketInquiry(ctx context.Context, m *domain.MarketInquiry) error
	GetMarketInquiry(ctx context.Context, id string) (*domain.MarketInquiry, error)
	ListMarketInquiries(ctx context.Context, inquiryID *string, itemID *string, page, pageSize int) ([]domain.MarketInquiry, int64, error)
	UpdateMarketInquiry(ctx context.Context, params UpdateParams) error
	DeleteMarketInquiry(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) MarketInquiryRepository { return &marketInquiryRepo{db: db} }
