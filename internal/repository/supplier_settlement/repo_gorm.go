package supplier_settlement

import (
	"context"
	"errors"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/supplier_settlement"
)

type supplierSettlementRepo struct{ db *gorm.DB }

func (r *supplierSettlementRepo) CreateSupplierSettlement(ctx context.Context, m *domain.SupplierSettlement) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *supplierSettlementRepo) GetSupplierSettlement(ctx context.Context, id string) (*domain.SupplierSettlement, error) {
	var out domain.SupplierSettlement
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *supplierSettlementRepo) ListSupplierSettlements(ctx context.Context, inquiryID *string, itemID *string, page, pageSize int) ([]domain.SupplierSettlement, int64, error) {
	var list []domain.SupplierSettlement
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.SupplierSettlement{})

	if inquiryID != nil && *inquiryID != "" {
		q = q.Where("inquiry_id = ?", *inquiryID)
	}
	if itemID != nil && *itemID != "" {
		q = q.Where("item_id = ?", *itemID)
	}

	q.Count(&total)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20
	}

	err := q.
		Order("supplier_name_snap ASC").
		Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *supplierSettlementRepo) UpdateSupplierSettlement(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	
	if params.InquiryID != nil {
		updates["inquiry_id"] = *params.InquiryID
	}
	if params.ItemID != nil {
		updates["item_id"] = *params.ItemID
	}
	if params.UpdateSupplierID {
		if params.SupplierID != nil {
			updates["supplier_id"] = *params.SupplierID
		} else {
			updates["supplier_id"] = nil
		}
	}
	if params.SupplierNameSnap != nil {
		updates["supplier_name_snap"] = *params.SupplierNameSnap
	}
	if params.FloatRatioSnap != nil {
		updates["float_ratio_snap"] = *params.FloatRatioSnap
	}
	if params.SettlementPrice != nil {
		updates["settlement_price"] = *params.SettlementPrice
	}

	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.SupplierSettlement{}).
		Where("id = ?", params.ID).
		Updates(updates).Error
}

func (r *supplierSettlementRepo) DeleteSupplierSettlement(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&domain.SupplierSettlement{}).Error
}
