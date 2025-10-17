package category

import (
	"context"

	"gorm.io/gorm"
	category "hdzk.cn/foodapp/internal/domain/category"
)

type categoryRepo struct{ db *gorm.DB }

func (r *categoryRepo) CreateCategory(ctx context.Context, m *category.Category) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *categoryRepo) GetCategory(ctx context.Context, id string) (*category.Category, error) {
	var out category.Category
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *categoryRepo) ListCategories(ctx context.Context, keyword string, page, pageSize int) ([]category.Category, int64, error) {
	var list []category.Category
	var total int64
	q := r.db.WithContext(ctx).Model(&category.Category{}).
		Where("is_deleted = 0")
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
	err := q.Order("name asc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *categoryRepo) UpdateCategory(ctx context.Context, id string, name string, code *string, pinyin *string, updateCode bool, updatePinyin bool) error {
	updates := map[string]any{
		"name": name,
	}
	if updateCode {
		if code != nil {
			updates["code"] = *code
		} else {
			updates["code"] = nil
		}
	}
	if updatePinyin {
		if pinyin != nil {
			updates["pinyin"] = *pinyin
		} else {
			updates["pinyin"] = nil
		}
	}
	return r.db.WithContext(ctx).Model(&category.Category{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(updates).Error
}

func (r *categoryRepo) DeleteCategory(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&category.Category{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}
