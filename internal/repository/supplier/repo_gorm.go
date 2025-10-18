package supplier

import (
	"context"
	"errors"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/supplier"
)

type supplierRepo struct{ db *gorm.DB }

func (r *supplierRepo) CreateSupplier(ctx context.Context, m *domain.Supplier) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *supplierRepo) GetSupplier(ctx context.Context, id string) (*domain.Supplier, error) {
	var out domain.Supplier
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *supplierRepo) ListSuppliers(ctx context.Context, params ListParams) ([]domain.Supplier, int64, error) {
	var list []domain.Supplier
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Supplier{}).
		Where("is_deleted = 0")
	if params.OrgID != "" {
		q = q.Where("org_id = ?", params.OrgID)
	}
	if params.Keyword != "" {
		pattern := "%" + params.Keyword + "%"
		q = q.Where("(name LIKE ? OR code LIKE ? OR pinyin LIKE ?)", pattern, pattern, pattern)
	}
	if params.Status != nil {
		q = q.Where("status = ?", *params.Status)
	}

	q.Count(&total)
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20
	}
	err := q.
		Order("sort ASC").
		Order("name ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *supplierRepo) UpdateSupplier(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	if params.Name != nil {
		updates["name"] = *params.Name
	}
	if params.UpdateCode {
		if params.Code != nil {
			updates["code"] = *params.Code
		} else {
			updates["code"] = nil
		}
	}
	if params.UpdatePinyin {
		if params.Pinyin != nil {
			updates["pinyin"] = *params.Pinyin
		} else {
			updates["pinyin"] = nil
		}
	}
	if params.UpdateSort {
		if params.Sort != nil {
			updates["sort"] = *params.Sort
		}
	}
	if params.Status != nil {
		updates["status"] = *params.Status
	}
	if params.Description != nil {
		updates["description"] = *params.Description
	}
	if params.FloatRatio != nil {
		updates["float_ratio"] = *params.FloatRatio
	}
	if params.UpdateStartTime {
		if params.StartTime != nil {
			updates["start_time"] = *params.StartTime
		} else {
			updates["start_time"] = nil
		}
	}
	if params.UpdateEndTime {
		if params.EndTime != nil {
			updates["end_time"] = *params.EndTime
		} else {
			updates["end_time"] = nil
		}
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.Supplier{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *supplierRepo) SoftDeleteSupplier(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Supplier{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *supplierRepo) HardDeleteSupplier(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.Supplier{}).Error
}
