package organ

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/organ"
)

type gormRepo struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &gormRepo{db: db}
}

func (r *gormRepo) Create(ctx context.Context, organ *domain.Organ) error {
	return r.db.WithContext(ctx).Create(organ).Error
}

func (r *gormRepo) Update(ctx context.Context, organ *domain.Organ) error {
	return r.db.WithContext(ctx).
		Model(&domain.Organ{}).
		Where("id = ? AND is_deleted = 0", organ.ID).
		Updates(map[string]any{
			"name":       organ.Name,
			"code":       organ.Code,
			"leader":     organ.Leader,
			"phone":      organ.Phone,
			"sort":       organ.Sort,
			"status":     organ.Status,
			"remark":     organ.Remark,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *gormRepo) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Organ{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(map[string]any{
			"is_deleted": 1,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *gormRepo) GetByID(ctx context.Context, id string) (*domain.Organ, error) {
	var organ domain.Organ
	if err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&organ).Error; err != nil {
		return nil, err
	}
	return &organ, nil
}

func (r *gormRepo) List(ctx context.Context, query domain.ListQuery) ([]domain.Organ, int64, error) {
	db := r.db.WithContext(ctx).Model(&domain.Organ{}).Where("is_deleted = 0")
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	db = db.Order("sort ASC").Order("created_at DESC")

	var items []domain.Organ
	if err := db.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
