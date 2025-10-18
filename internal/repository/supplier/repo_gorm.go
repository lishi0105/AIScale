package supplier

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	supplier "hdzk.cn/foodapp/internal/domain/supplier"
)

type supplierRepo struct{ db *gorm.DB }

func (r *supplierRepo) CreateSupplier(ctx context.Context, m *supplier.Supplier) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *supplierRepo) GetSupplier(ctx context.Context, id string) (*supplier.Supplier, error) {
	var out supplier.Supplier
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *supplierRepo) ListSuppliers(ctx context.Context, keyword string, orgID *string, status *int, page, pageSize int) ([]supplier.Supplier, int64, error) {
	var list []supplier.Supplier
	var total int64
	q := r.db.WithContext(ctx).Model(&supplier.Supplier{}).
		Where("is_deleted = 0")
	
	// 过滤 org_id
	if orgID != nil && *orgID != "" {
		q = q.Where("org_id = ?", *orgID)
	}
	
	// 过滤 status
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	
	// 关键词搜索
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

func (r *supplierRepo) UpdateSupplier(ctx context.Context, id string, name string, code *string, pinyin *string, sort *int, status *int, description *string, floatRatio *float64, orgID *string, startTime *time.Time, endTime *time.Time, updateCode bool, updatePinyin bool, updateSort bool, updateStatus bool, updateDescription bool, updateFloatRatio bool, updateOrgID bool, updateStartTime bool, updateEndTime bool) error {
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
	if updateStatus {
		if status != nil {
			updates["status"] = *status
		}
	}
	if updateDescription {
		if description != nil {
			updates["description"] = *description
		}
	}
	if updateFloatRatio {
		if floatRatio != nil {
			// 验证 float_ratio > 0
			if *floatRatio <= 0 {
				return errors.New("float_ratio 必须大于 0")
			}
			updates["float_ratio"] = *floatRatio
		}
	}
	if updateOrgID {
		if orgID != nil {
			updates["org_id"] = *orgID
		} else {
			updates["org_id"] = nil
		}
	}
	if updateStartTime {
		if startTime != nil {
			updates["start_time"] = *startTime
		} else {
			updates["start_time"] = nil
		}
	}
	if updateEndTime {
		if endTime != nil {
			updates["end_time"] = *endTime
		} else {
			updates["end_time"] = nil
		}
	}
	
	// 验证时间范围
	if updateStartTime && updateEndTime && startTime != nil && endTime != nil {
		if startTime.After(*endTime) {
			return errors.New("start_time 必须小于等于 end_time")
		}
	}
	
	return r.db.WithContext(ctx).Model(&supplier.Supplier{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(updates).Error
}

func (r *supplierRepo) SoftDeleteSupplier(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&supplier.Supplier{}).
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
		Delete(&supplier.Supplier{}).Error
}
