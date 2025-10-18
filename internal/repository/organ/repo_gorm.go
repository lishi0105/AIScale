package organ

import (
	"context"
	"errors"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/organ"
	foodDB "hdzk.cn/foodapp/internal/storage/db"
)

type gormRepo struct {
	db *gorm.DB
}

func (r *gormRepo) Create(ctx context.Context, m *domain.Organ) error {
	if m.ParentID == nil {
		tmp := foodDB.DefaultOrgID
		m.ParentID = &tmp
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *gormRepo) GetByID(ctx context.Context, id string) (*domain.Organ, error) {
	var o domain.Organ
	if err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *gormRepo) List(ctx context.Context, keyword string, Deleted, Role *int, page, pageSize int) ([]domain.Organ, int64, error) {
	var (
		list  []domain.Organ
		total int64
	)

	q := r.db.WithContext(ctx).Model(&domain.Organ{})

	if Deleted != nil {
		q = q.Where("is_deleted = ?", *Deleted)
	}
	if Role != nil {
		q = q.Where("role = ?", *Role)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(name LIKE ? OR code LIKE ? OR pinyin LIKE ?)", pattern, pattern, pattern)
	}

	// ---------- 分页参数 ----------
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20
	}

	// ---------- 统计总数 ----------
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ---------- 查询结果 ----------
	err := q.
		Order("sort ASC").
		Order("name ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error

	return list, total, err
}

func (r *gormRepo) UpdateFields(ctx context.Context, id string, fields map[string]any) error {
	if len(fields) == 0 {
		return errors.New("没有要更新项目")
	}
	return r.db.WithContext(ctx).
		Model(&domain.Organ{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(fields).Error
}

func (r *gormRepo) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Organ{}).
		Where("id = ? AND is_deleted = 0", id).
		Update("is_deleted", 1).Error
}

func (r *gormRepo) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id =?", id).
		Delete(&domain.Organ{}).Error
}
