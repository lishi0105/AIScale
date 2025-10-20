package inquiry

import (
	"context"
	"time"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/inquiry"
)

type InquiryRepository interface {
	// 创建询价记录
	CreateInquiry(ctx context.Context, inquiry *domain.BasePriceInquiry) error
	
	// 根据ID获取询价记录
	GetInquiry(ctx context.Context, id string) (*domain.BasePriceInquiry, error)
	
	// 查询询价记录列表
	ListInquiries(ctx context.Context, params domain.InquiryQueryParams) ([]domain.BasePriceInquiry, int64, error)
	
	// 更新询价记录
	UpdateInquiry(ctx context.Context, params domain.UpdateInquiryParams) error
	
	// 软删除询价记录
	SoftDeleteInquiry(ctx context.Context, id string) error
	
	// 硬删除询价记录
	HardDeleteInquiry(ctx context.Context, id string) error
	
	// 检查标题和日期的唯一性（用于创建和更新时的验证）
	CheckTitleDateUnique(ctx context.Context, orgID, title string, inquiryDate time.Time, excludeID *string) (bool, error)
}

func NewRepository(db *gorm.DB) InquiryRepository {
	return &inquiryRepo{db: db}
}