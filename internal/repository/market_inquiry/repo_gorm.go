package market_inquiry

import (
	"context"
	"errors"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/market_inquiry"
)

type marketInquiryRepo struct{ db *gorm.DB }

func (r *marketInquiryRepo) CreateMarketInquiry(ctx context.Context, m *domain.MarketInquiry) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *marketInquiryRepo) GetMarketInquiry(ctx context.Context, id string) (*domain.MarketInquiry, error) {
	var out domain.MarketInquiry
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *marketInquiryRepo) ListMarketInquiries(ctx context.Context, params ListParams) ([]domain.MarketInquiry, int64, error) {
	var list []domain.MarketInquiry
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.MarketInquiry{}).
		Where("is_deleted = 0 AND inquiry_id = ?", params.InquiryID)

	if params.ItemID != nil && *params.ItemID != "" {
		q = q.Where("item_id = ?", *params.ItemID)
	}
	if params.MarketID != nil && *params.MarketID != "" {
		q = q.Where("market_id = ?", *params.MarketID)
	}

	q.Count(&total)
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 1000 {
		params.PageSize = 20
	}

	err := q.
		Order("market_name_snap ASC").
		Order("created_at ASC").
		Limit(params.PageSize).Offset((params.Page - 1) * params.PageSize).
		Find(&list).Error
	return list, total, err
}

func (r *marketInquiryRepo) UpdateMarketInquiry(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	if params.MarketID != nil {
		updates["market_id"] = *params.MarketID
	}
	if params.MarketNameSnap != nil {
		updates["market_name_snap"] = *params.MarketNameSnap
	}
	if params.Price != nil {
		updates["price"] = *params.Price
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.MarketInquiry{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *marketInquiryRepo) SoftDeleteMarketInquiry(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.MarketInquiry{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *marketInquiryRepo) HardDeleteMarketInquiry(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.MarketInquiry{}).Error
}

func (r *marketInquiryRepo) BatchCreateMarketInquiries(ctx context.Context, items []domain.MarketInquiry) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(items, 100).Error
}