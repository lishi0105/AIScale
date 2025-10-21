package price_inquiry

import (
	"context"
	"errors"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/price_inquiry"
)

type priceInquiryRepo struct{ db *gorm.DB }

func (r *priceInquiryRepo) CreatePriceInquiry(ctx context.Context, m *domain.PriceInquiry) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *priceInquiryRepo) GetPriceInquiry(ctx context.Context, id string) (*domain.PriceInquiry, error) {
	var out domain.PriceInquiry
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *priceInquiryRepo) ListPriceInquiries(ctx context.Context, params ListParams) ([]domain.PriceInquiry, int64, error) {
	var list []domain.PriceInquiry
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.PriceInquiry{}).
		Where("is_deleted = 0 AND org_id = ?", params.OrgID)

	if params.Keyword != "" {
		pattern := "%" + params.Keyword + "%"
		q = q.Where("inquiry_title LIKE ?", pattern)
	}
	if params.Year != nil {
		q = q.Where("inquiry_year = ?", *params.Year)
	}
	if params.Month != nil {
		q = q.Where("inquiry_month = ?", *params.Month)
	}
	if params.TenDay != nil {
		q = q.Where("inquiry_ten_day = ?", *params.TenDay)
	}

	q.Count(&total)
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 1000 {
		params.PageSize = 20
	}

	err := q.
		Order("inquiry_date DESC").
		Order("created_at DESC").
		Limit(params.PageSize).Offset((params.Page - 1) * params.PageSize).
		Find(&list).Error
	return list, total, err
}

func (r *priceInquiryRepo) UpdatePriceInquiry(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	if params.InquiryTitle != nil {
		updates["inquiry_title"] = *params.InquiryTitle
	}
	if params.InquiryDate != nil {
		updates["inquiry_date"] = *params.InquiryDate
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.PriceInquiry{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *priceInquiryRepo) SoftDeletePriceInquiry(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.PriceInquiry{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *priceInquiryRepo) HardDeletePriceInquiry(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.PriceInquiry{}).Error
}