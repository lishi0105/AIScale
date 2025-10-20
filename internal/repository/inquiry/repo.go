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
	InquiryStartDate *time.Time
	InquiryEndDate   *time.Time
	UpdateMarket1    bool
	UpdateMarket2    bool
	UpdateMarket3    bool
}

type InquiryRepository interface {
	CreateInquiry(ctx context.Context, m *domain.Inquiry) error
	GetInquiry(ctx context.Context, id string) (*domain.Inquiry, error)
	ListInquiries(ctx context.Context, keyword string, orgID string, startDate, endDate *time.Time, page, pageSize int) ([]domain.Inquiry, int64, error)
	UpdateInquiry(ctx context.Context, params UpdateParams) error
	SoftDeleteInquiry(ctx context.Context, id string) error
	HardDeleteInquiry(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) InquiryRepository { return &inquiryRepo{db: db} }
