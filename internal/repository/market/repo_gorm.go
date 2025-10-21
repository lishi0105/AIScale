package market

import (
	"context"
	"errors"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/market"
)

type marketRepo struct{ db *gorm.DB }

func (r *marketRepo) CreateMarket(ctx context.Context, m *domain.Market) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *marketRepo) GetMarket(ctx context.Context, id string) (*domain.Market, error) {
	var out domain.Market
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *marketRepo) ListMarkets(ctx context.Context, keyword string, orgID string, page, pageSize int) ([]domain.Market, int64, error) {
	var list []domain.Market
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Market{}).
		Where("is_deleted = 0 AND org_id = ?", orgID)

	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(name LIKE ? OR code LIKE ?)", pattern, pattern)
	}

	q.Count(&total)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20
	}

	err := q.
		Order("sort ASC").
		Order("name ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *marketRepo) UpdateMarket(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	if params.Name != nil {
		updates["name"] = *params.Name
	}
	if params.Code != nil {
		updates["code"] = *params.Code
	}
	if params.Sort != nil {
		updates["sort"] = *params.Sort
	}
	if params.OrgID != nil {
		updates["org_id"] = *params.OrgID
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.Market{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *marketRepo) SoftDeleteMarket(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Market{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *marketRepo) HardDeleteMarket(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.Market{}).Error
}
