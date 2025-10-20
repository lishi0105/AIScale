package inquiry

import (
    "context"
    "time"

    "gorm.io/gorm"
    domain "hdzk.cn/foodapp/internal/domain/inquiry"
)

type UpdateParams struct {
    ID               string
    InquiryTitle     *string
    InquiryDate      *time.Time
    Market1          *string
    Market2          *string
    Market3          *string
}

type Repository interface {
    Create(ctx context.Context, m *domain.PriceInquiry) error
    Get(ctx context.Context, id string) (*domain.PriceInquiry, error)
    List(ctx context.Context, orgID string, keyword string, dateFrom, dateTo *time.Time, page, pageSize int) ([]domain.PriceInquiry, int64, error)
    Update(ctx context.Context, params UpdateParams) error
    SoftDelete(ctx context.Context, id string) error
    HardDelete(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) Repository { return &repo{db: db} }
