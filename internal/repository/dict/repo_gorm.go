package dict

import (
	"context"

	"gorm.io/gorm"
	dict "hdzk.cn/foodapp/internal/domain/dict"
)

type dictRepo struct{ db *gorm.DB }

func New(db *gorm.DB) DictRepo { return &dictRepo{db: db} }

// ---------- Unit ----------
func (r *dictRepo) CreateUnit(ctx context.Context, m *dict.Unit) error {
	return r.db.WithContext(ctx).Create(m).Error
}
func (r *dictRepo) GetUnit(ctx context.Context, id string) (*dict.Unit, error) {
	var out dict.Unit
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}
func (r *dictRepo) ListUnits(ctx context.Context, keyword string, page, pageSize int) ([]dict.Unit, int64, error) {
	var list []dict.Unit
	var total int64
	q := r.db.WithContext(ctx).Model(&dict.Unit{}).
		Where("is_deleted = 0")
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
	err := q.Order("sort asc, name asc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *dictRepo) UpdateUnit(ctx context.Context, id string, name string, code *string, sort int, updateCode bool) error {
	updates := map[string]any{
		"name": name,
		"sort": sort,
	}
	if updateCode {
		if code != nil {
			updates["code"] = *code
		} else {
			updates["code"] = nil
		}
	}
	return r.db.WithContext(ctx).Model(&dict.Unit{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(updates).Error
}

func (r *dictRepo) DeleteUnit(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&dict.Unit{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

// ---------- Spec ----------
func (r *dictRepo) CreateSpec(ctx context.Context, m *dict.Spec) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *dictRepo) GetSpec(ctx context.Context, id string) (*dict.Spec, error) {
	var out dict.Spec
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}
func (r *dictRepo) ListSpecs(ctx context.Context, keyword string, page, pageSize int) ([]dict.Spec, int64, error) {
	var list []dict.Spec
	var total int64
	q := r.db.WithContext(ctx).Model(&dict.Spec{}).
		Where("is_deleted = 0")
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
	err := q.Order("sort asc, name asc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *dictRepo) UpdateSpec(ctx context.Context, id string, name string, code *string, sort int, updateCode bool) error {
	updates := map[string]any{"name": name, "sort": sort}
	if updateCode {
		if code != nil {
			updates["code"] = *code
		} else {
			updates["code"] = nil
		}
	}
	return r.db.WithContext(ctx).Model(&dict.Spec{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(updates).Error
}

func (r *dictRepo) DeleteSpec(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&dict.Spec{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

// ---------- MealTime ----------
func (r *dictRepo) CreateMealTime(ctx context.Context, m *dict.MealTime) error {
	return r.db.WithContext(ctx).Create(m).Error
}
func (r *dictRepo) GetMealTime(ctx context.Context, id string) (*dict.MealTime, error) {
	var out dict.MealTime
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}
func (r *dictRepo) ListMealTimes(ctx context.Context, keyword string, page, pageSize int) ([]dict.MealTime, int64, error) {
	var list []dict.MealTime
	var total int64
	q := r.db.WithContext(ctx).Model(&dict.MealTime{}).
		Where("is_deleted = 0")
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
	err := q.Order("sort asc, name asc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}
func (r *dictRepo) UpdateMealTime(ctx context.Context, id string, name string, code *string, sort int, updateCode bool) error {
	updates := map[string]any{"name": name, "sort": sort}
	if updateCode {
		if code != nil {
			updates["code"] = *code
		} else {
			updates["code"] = nil
		}
	}
	return r.db.WithContext(ctx).Model(&dict.MealTime{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(updates).Error
}
func (r *dictRepo) DeleteMealTime(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&dict.MealTime{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}
