package category

import (
	"context"
	"errors"

	"gorm.io/gorm"
	category "hdzk.cn/foodapp/internal/domain/category"
)

type categoryRepo struct{ db *gorm.DB }

func (r *categoryRepo) Create(ctx context.Context, m *category.Category) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *categoryRepo) Get(ctx context.Context, id string) (*category.Category, error) {
	var out category.Category
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *categoryRepo) List(ctx context.Context, keyword string, org_id string, page, pageSize int) ([]category.Category, int64, error) {
	var list []category.Category
	var total int64
	q := r.db.WithContext(ctx).Model(&category.Category{}).
		Where("is_deleted = 0 AND org_id = ?", org_id)
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
		Order("name asc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *categoryRepo) Update(ctx context.Context, id string, name string, code *string, pinyin *string, sort *int, updateCode bool, updatePinyin bool, updateSort bool) error {
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
	if updateSort {
		if sort != nil {
			updates["sort"] = *sort
		}
	}
	return r.db.WithContext(ctx).Model(&category.Category{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(updates).Error
}

func (r *categoryRepo) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&category.Category{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *categoryRepo) HardDelete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&category.Category{}).Error
}

func (r *categoryRepo) FindByName(ctx context.Context, name string, orgID string) (*category.Category, error) {
	var out category.Category
	err := r.db.WithContext(ctx).
		Where("name = ? AND org_id = ? AND is_deleted = 0", name, orgID).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}
