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

func (r *inquiryItemRepo) ListInquiryItems(ctx context.Context, params ListParams) ([]domain.InquiryItem, int64, error) {
	var list []domain.InquiryItem
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.InquiryItem{}).
		Where("is_deleted = 0 AND inquiry_id = ?", params.InquiryID)

	if params.CategoryID != nil {
		q = q.Where("category_id = ?", *params.CategoryID)
	}
	if params.GoodsID != nil {
		q = q.Where("goods_id = ?", *params.GoodsID)
	}

	q.Count(&total)
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 1000 {
		params.PageSize = 20
	}

	err := q.
		Order("sort ASC").
		Order("goods_name_snap ASC").
		Limit(params.PageSize).Offset((params.Page - 1) * params.PageSize).
		Find(&list).Error
	return list, total, err
}

func (r *inquiryItemRepo) UpdateInquiryItem(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	if params.GoodsNameSnap != nil {
		updates["goods_name_snap"] = *params.GoodsNameSnap
	}
	if params.CategoryNameSnap != nil {
		updates["category_name_snap"] = *params.CategoryNameSnap
	}
	if params.SpecNameSnap != nil {
		updates["spec_name_snap"] = *params.SpecNameSnap
	}
	if params.UnitNameSnap != nil {
		updates["unit_name_snap"] = *params.UnitNameSnap
	}
	if params.GuidePrice != nil {
		updates["guide_price"] = *params.GuidePrice
	}
	if params.LastMonthAvgPrice != nil {
		updates["last_month_avg_price"] = *params.LastMonthAvgPrice
	}
	if params.CurrentAvgPrice != nil {
		updates["current_avg_price"] = *params.CurrentAvgPrice
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

func (r *inquiryItemRepo) BatchCreateInquiryItems(ctx context.Context, items []domain.InquiryItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(items, 100).Error
}