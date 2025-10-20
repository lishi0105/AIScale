package goods

import (
	"context"
	"errors"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/goods"
)

type goodsRepo struct{ db *gorm.DB }

func (r *goodsRepo) CreateGoods(ctx context.Context, m *domain.Goods) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *goodsRepo) GetGoods(ctx context.Context, id string) (*domain.Goods, error) {
	var out domain.Goods
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *goodsRepo) ListGoods(ctx context.Context, keyword string, orgID string, categoryID, specID *string, page, pageSize int) ([]domain.Goods, int64, error) {
	var list []domain.Goods
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Goods{}).
		Where("is_deleted = 0 AND org_id = ?", orgID)

	if categoryID != nil && *categoryID != "" {
		q = q.Where("category_id = ?", *categoryID)
	}
	if specID != nil && *specID != "" {
		q = q.Where("spec_id = ?", *specID)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(name LIKE ? OR code LIKE ? OR pinyin LIKE ?)", pattern, pattern, pattern)
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

func (r *goodsRepo) UpdateGoods(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	if params.Name != nil {
		updates["name"] = *params.Name
	}
	if params.Code != nil {
		updates["code"] = *params.Code
	}
	if params.UpdatePinyin {
		if params.Pinyin != nil {
			updates["pinyin"] = *params.Pinyin
		} else {
			updates["pinyin"] = nil
		}
	}
	if params.Sort != nil {
		updates["sort"] = *params.Sort
	}
	if params.SpecID != nil {
		updates["spec_id"] = *params.SpecID
	}
	if params.CategoryID != nil {
		updates["category_id"] = *params.CategoryID
	}
	if params.UpdateImageURL {
		if params.ImageURL != nil {
			updates["image_url"] = *params.ImageURL
		} else {
			updates["image_url"] = nil
		}
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.Goods{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *goodsRepo) SoftDeleteGoods(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Goods{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *goodsRepo) HardDeleteGoods(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.Goods{}).Error
}
