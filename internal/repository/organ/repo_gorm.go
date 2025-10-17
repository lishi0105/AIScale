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

func (r *gormRepo) List(ctx context.Context, q domain.ListQuery) ([]*domain.Organ, int64, error) {
	tx := r.db.WithContext(ctx).Model(&domain.Organ{}).
		Where("sort <> ?", -1)
	if q.NameLike != "" {
		tx = tx.Where("name LIKE ?", "%"+q.NameLike+"%")
	}
	if q.Deleted != nil {
		tx = tx.Where("is_deleted = ?", *q.Deleted)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit, offset := q.Limit, q.Offset
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var rows []*domain.Organ
	if err := tx.
		Order("sort ASC, created_at DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
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
