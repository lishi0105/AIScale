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

func (r *supplierSettlementRepo) ListSupplierSettlements(ctx context.Context, params ListParams) ([]domain.SupplierSettlement, int64, error) {
	var list []domain.SupplierSettlement
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.SupplierSettlement{}).
		Where("inquiry_id = ?", params.InquiryID)

	if params.ItemID != nil {
		q = q.Where("item_id = ?", *params.ItemID)
	}
	if params.SupplierID != nil {
		q = q.Where("supplier_id = ?", *params.SupplierID)
	}

	q.Count(&total)
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 1000 {
		params.PageSize = 20
	}

	err := q.
		Order("supplier_name_snap ASC").
		Order("created_at ASC").
		Limit(params.PageSize).Offset((params.Page - 1) * params.PageSize).
		Find(&list).Error
	return list, total, err
}

func (r *supplierSettlementRepo) UpdateSupplierSettlement(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	if params.SupplierID != nil {
		updates["supplier_id"] = *params.SupplierID
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

func (r *supplierSettlementRepo) SoftDeleteSupplierSettlement(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.SupplierSettlement{}).
		Where("id = ?", id).
		Delete(&domain.SupplierSettlement{}).Error
}

func (r *supplierSettlementRepo) HardDeleteSupplierSettlement(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.SupplierSettlement{}).Error
}

func (r *supplierSettlementRepo) BatchCreateSupplierSettlements(ctx context.Context, settlements []domain.SupplierSettlement) error {
	if len(settlements) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(settlements, 100).Error
}

func (r *supplierSettlementRepo) GetByItemAndSupplier(ctx context.Context, itemID, supplierName string) (*domain.SupplierSettlement, error) {
	var out domain.SupplierSettlement
	err := r.db.WithContext(ctx).
		Where("item_id = ? AND supplier_name_snap = ?", itemID, supplierName).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}