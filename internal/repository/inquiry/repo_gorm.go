package inquiry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/inquiry"
)

type inquiryRepo struct {
	db *gorm.DB
}

func (r *inquiryRepo) CreateInquiry(ctx context.Context, inquiry *domain.BasePriceInquiry) error {
	return r.db.WithContext(ctx).Create(inquiry).Error
}

func (r *inquiryRepo) GetInquiry(ctx context.Context, id string) (*domain.BasePriceInquiry, error) {
	var inquiry domain.BasePriceInquiry
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&inquiry).Error
	if err != nil {
		return nil, err
	}
	return &inquiry, nil
}

func (r *inquiryRepo) ListInquiries(ctx context.Context, params domain.InquiryQueryParams) ([]domain.BasePriceInquiry, int64, error) {
	var inquiries []domain.BasePriceInquiry
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.BasePriceInquiry{}).
		Where("is_deleted = 0")

	// 组织ID过滤
	if params.OrgID != "" {
		query = query.Where("org_id = ?", params.OrgID)
	}

	// 关键词搜索（标题）
	if params.Keyword != "" {
		keyword := "%" + strings.TrimSpace(params.Keyword) + "%"
		query = query.Where("inquiry_title LIKE ?", keyword)
	}

	// 日期范围过滤
	if params.StartDate != nil {
		query = query.Where("inquiry_date >= ?", *params.StartDate)
	}
	if params.EndDate != nil {
		query = query.Where("inquiry_date <= ?", *params.EndDate)
	}

	// 市场过滤
	if params.Market1 != nil && *params.Market1 != "" {
		query = query.Where("market_1 = ?", *params.Market1)
	}
	if params.Market2 != nil && *params.Market2 != "" {
		query = query.Where("market_2 = ?", *params.Market2)
	}
	if params.Market3 != nil && *params.Market3 != "" {
		query = query.Where("market_3 = ?", *params.Market3)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页和排序
	offset := (params.Page - 1) * params.PageSize
	err := query.Order("inquiry_date DESC, created_at DESC").
		Offset(offset).
		Limit(params.PageSize).
		Find(&inquiries).Error

	return inquiries, total, err
}

func (r *inquiryRepo) UpdateInquiry(ctx context.Context, params domain.UpdateInquiryParams) error {
	updates := make(map[string]interface{})

	if params.InquiryTitle != nil {
		updates["inquiry_title"] = *params.InquiryTitle
		// 同时更新 active_title
		updates["active_title"] = *params.InquiryTitle
	}
	if params.InquiryDate != nil {
		updates["inquiry_date"] = *params.InquiryDate
	}
	if params.Market1 != nil {
		updates["market_1"] = params.Market1
	}
	if params.Market2 != nil {
		updates["market_2"] = params.Market2
	}
	if params.Market3 != nil {
		updates["market_3"] = params.Market3
	}
	if params.InquiryStartDate != nil {
		updates["inquiry_start_date"] = *params.InquiryStartDate
	}
	if params.InquiryEndDate != nil {
		updates["inquiry_end_date"] = *params.InquiryEndDate
	}

	if len(updates) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}

	// 更新时间戳
	updates["updated_at"] = time.Now()

	return r.db.WithContext(ctx).
		Model(&domain.BasePriceInquiry{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *inquiryRepo) SoftDeleteInquiry(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&domain.BasePriceInquiry{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(map[string]interface{}{
			"is_deleted":   1,
			"active_title": nil, // 软删除时清空 active_title
			"updated_at":   time.Now(),
		}).Error
}

func (r *inquiryRepo) HardDeleteInquiry(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&domain.BasePriceInquiry{}).Error
}

func (r *inquiryRepo) CheckTitleDateUnique(ctx context.Context, orgID, title string, inquiryDate time.Time, excludeID *string) (bool, error) {
	query := r.db.WithContext(ctx).
		Model(&domain.BasePriceInquiry{}).
		Where("org_id = ? AND inquiry_title = ? AND inquiry_date = ? AND is_deleted = 0", 
			orgID, title, inquiryDate)

	// 排除指定ID（用于更新时检查）
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	// 返回 true 表示存在重复，false 表示不重复
	return count > 0, nil
}