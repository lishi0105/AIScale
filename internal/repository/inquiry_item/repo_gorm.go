package inquiry_item

import (
	"context"
	"errors"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/inquiry_item"
)

type inquiryItemRepo struct{ db *gorm.DB }

func (r *inquiryItemRepo) CreateInquiryItem(ctx context.Context, m *domain.InquiryItem) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *inquiryItemRepo) GetInquiryItem(ctx context.Context, id string) (*domain.InquiryItem, error) {
	var out domain.InquiryItem
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *inquiryItemRepo) ListInquiryItems(ctx context.Context, inquiryID string, categoryID *string, page, pageSize int) ([]domain.InquiryItem, int64, error) {
	var list []domain.InquiryItem
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.InquiryItem{}).
		Where("is_deleted = 0 AND inquiry_id = ?", inquiryID)

	if categoryID != nil && *categoryID != "" {
		q = q.Where("category_id = ?", *categoryID)
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
		Order("goods_name_snap ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *inquiryItemRepo) UpdateInquiryItem(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	
	if params.InquiryID != nil {
		updates["inquiry_id"] = *params.InquiryID
	}
	if params.GoodsID != nil {
		updates["goods_id"] = *params.GoodsID
	}
	if params.CategoryID != nil {
		updates["category_id"] = *params.CategoryID
	}
	if params.UpdateSpecID {
		if params.SpecID != nil {
			updates["spec_id"] = *params.SpecID
		} else {
			updates["spec_id"] = nil
		}
	}
	if params.UpdateUnitID {
		if params.UnitID != nil {
			updates["unit_id"] = *params.UnitID
		} else {
			updates["unit_id"] = nil
		}
	}
	if params.GoodsNameSnap != nil {
		updates["goods_name_snap"] = *params.GoodsNameSnap
	}
	if params.CategoryNameSnap != nil {
		updates["category_name_snap"] = *params.CategoryNameSnap
	}
	if params.UpdateSpecName {
		if params.SpecNameSnap != nil {
			updates["spec_name_snap"] = *params.SpecNameSnap
		} else {
			updates["spec_name_snap"] = nil
		}
	}
	if params.UpdateUnitName {
		if params.UnitNameSnap != nil {
			updates["unit_name_snap"] = *params.UnitNameSnap
		} else {
			updates["unit_name_snap"] = nil
		}
	}
	if params.UpdateGuidePrice {
		if params.GuidePrice != nil {
			updates["guide_price"] = *params.GuidePrice
		} else {
			updates["guide_price"] = nil
		}
	}
	if params.UpdateLastMonth {
		if params.LastMonthAvgPrice != nil {
			updates["last_month_avg_price"] = *params.LastMonthAvgPrice
		} else {
			updates["last_month_avg_price"] = nil
		}
	}
	if params.UpdateCurrentAvg {
		if params.CurrentAvgPrice != nil {
			updates["current_avg_price"] = *params.CurrentAvgPrice
		} else {
			updates["current_avg_price"] = nil
		}
	}
	if params.Sort != nil {
		updates["sort"] = *params.Sort
	}

	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.InquiryItem{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *inquiryItemRepo) SoftDeleteInquiryItem(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.InquiryItem{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *inquiryItemRepo) HardDeleteInquiryItem(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.InquiryItem{}).Error
}
