package market

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/market"
)

type UpdateParams struct {
	ID     string
	Name   *string
	Code   *string
	Sort   *int
}

type MarketRepository interface {
	CreateMarket(ctx context.Context, m *domain.Market) error
	GetMarket(ctx context.Context, id string) (*domain.Market, error)
	ListMarkets(ctx context.Context, keyword string, orgID string, page, pageSize int) ([]domain.Market, int64, error)
	UpdateMarket(ctx context.Context, params UpdateParams) error
	SoftDeleteMarket(ctx context.Context, id string) error
	HardDeleteMarket(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) MarketRepository { return &marketRepo{db: db} }