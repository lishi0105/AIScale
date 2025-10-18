package supplier

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"hdzk.cn/foodapp/internal/domain/supplier"
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

func (r *supplierRepo) ListSuppliers(ctx context.Context, keyword string, orgID *string, status *int, contactName, contactPhone, contactEmail, contactAddress *string, page, pageSize int) ([]supplier.Supplier, int64, error) {
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

	if contactName != nil {
		q = q.Where("contact_name = ?", *contactName)
	}
	if contactPhone != nil {
		q = q.Where("contact_phone = ?", *contactPhone)
	}
	if contactEmail != nil {
		q = q.Where("contact_email = ?", *contactEmail)
	}
	if contactAddress != nil {
		q = q.Where("contact_address = ?", *contactAddress)
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
	if params.UpdateContactName {
		if params.ContactName != nil {
			updates["contact_name"] = *params.ContactName
		} else {
			updates["contact_name"] = nil
		}
	}
	if params.UpdateContactPhone {
		if params.ContactPhone != nil {
			updates["contact_phone"] = *params.ContactPhone
		} else {
			updates["contact_phone"] = nil
		}
	}
	if params.UpdateContactEmail {
		if params.ContactEmail != nil {
			updates["contact_email"] = *params.ContactEmail
		} else {
			updates["contact_email"] = nil
		}
	}
	if params.UpdateContactAddress {
		if params.ContactAddress != nil {
			updates["contact_address"] = *params.ContactAddress
		} else {
			updates["contact_address"] = nil
		}
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
