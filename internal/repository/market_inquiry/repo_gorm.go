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
		Where("id = ?", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *marketInquiryRepo) ListMarketInquiries(ctx context.Context, inquiryID *string, itemID *string, page, pageSize int) ([]domain.MarketInquiry, int64, error) {
	var list []domain.MarketInquiry
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.MarketInquiry{})

	if inquiryID != nil && *inquiryID != "" {
		q = q.Where("inquiry_id = ?", *inquiryID)
	}
	if itemID != nil && *itemID != "" {
		q = q.Where("item_id = ?", *itemID)
	}

	q.Count(&total)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20
	}

	err := q.
		Order("market_name_snap ASC").
		Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *marketInquiryRepo) UpdateMarketInquiry(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	
	if params.InquiryID != nil {
		updates["inquiry_id"] = *params.InquiryID
	}
	if params.ItemID != nil {
		updates["item_id"] = *params.ItemID
	}
	if params.UpdateMarketID {
		if params.MarketID != nil {
			updates["market_id"] = *params.MarketID
		} else {
			updates["market_id"] = nil
		}
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
		Where("id = ?", params.ID).
		Updates(updates).Error
}

func (r *marketInquiryRepo) DeleteMarketInquiry(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&domain.MarketInquiry{}).Error
}
