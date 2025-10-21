package price_inquiry

import (
	"context"
	"time"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/price_inquiry"
)

type UpdateParams struct {
	ID           string
	OrgID        *string
	InquiryTitle *string
	InquiryDate  *time.Time
}

type ListParams struct {
	OrgID         string
	Keyword       string
	Year          *int16
	Month         *int8
	TenDay        *int8
	Page          int
	PageSize      int
}

type PriceInquiryRepository interface {
	CreatePriceInquiry(ctx context.Context, m *domain.PriceInquiry) error
	GetPriceInquiry(ctx context.Context, id string) (*domain.PriceInquiry, error)
	ListPriceInquiries(ctx context.Context, params ListParams) ([]domain.PriceInquiry, int64, error)
	UpdatePriceInquiry(ctx context.Context, params UpdateParams) error
	SoftDeletePriceInquiry(ctx context.Context, id string) error
	HardDeletePriceInquiry(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) PriceInquiryRepository { return &priceInquiryRepo{db: db} }
