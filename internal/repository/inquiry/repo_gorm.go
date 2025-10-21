package inquiry

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/inquiry"
)

// ========== BaseMarket Repository ==========

type marketRepo struct{ db *gorm.DB }

func (r *marketRepo) CreateMarket(ctx context.Context, m *domain.BaseMarket) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *marketRepo) GetMarket(ctx context.Context, id string) (*domain.BaseMarket, error) {
	var out domain.BaseMarket
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *marketRepo) ListMarkets(ctx context.Context, keyword string, orgID string, page, pageSize int) ([]domain.BaseMarket, int64, error) {
	var list []domain.BaseMarket
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.BaseMarket{}).
		Where("is_deleted = 0 AND org_id = ?", orgID)

	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(name LIKE ? OR code LIKE ?)", pattern, pattern)
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
		Order("name ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *marketRepo) UpdateMarket(ctx context.Context, params MarketUpdateParams) error {
	updates := map[string]any{}
	if params.Name != nil {
		updates["name"] = *params.Name
	}
	if params.Code != nil {
		updates["code"] = *params.Code
	}
	if params.Sort != nil {
		updates["sort"] = *params.Sort
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.BaseMarket{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *marketRepo) SoftDeleteMarket(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.BaseMarket{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *marketRepo) HardDeleteMarket(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.BaseMarket{}).Error
}

// ========== BasePriceInquiry Repository ==========

type inquiryRepo struct{ db *gorm.DB }

func (r *inquiryRepo) CreateInquiry(ctx context.Context, m *domain.BasePriceInquiry) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *inquiryRepo) GetInquiry(ctx context.Context, id string) (*domain.BasePriceInquiry, error) {
	var out domain.BasePriceInquiry
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *inquiryRepo) ListInquiries(ctx context.Context, keyword string, orgID string, year, month, tenDay *int, page, pageSize int) ([]domain.BasePriceInquiry, int64, error) {
	var list []domain.BasePriceInquiry
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.BasePriceInquiry{}).
		Where("is_deleted = 0 AND org_id = ?", orgID)

	if year != nil && *year > 0 {
		q = q.Where("inquiry_year = ?", *year)
	}
	if month != nil && *month > 0 {
		q = q.Where("inquiry_month = ?", *month)
	}
	if tenDay != nil && *tenDay > 0 {
		q = q.Where("inquiry_ten_day = ?", *tenDay)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("inquiry_title LIKE ?", pattern)
	}

	q.Count(&total)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20
	}

	err := q.
		Order("inquiry_date DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *inquiryRepo) UpdateInquiry(ctx context.Context, params InquiryUpdateParams) error {
	updates := map[string]any{}
	if params.InquiryTitle != nil {
		updates["inquiry_title"] = *params.InquiryTitle
	}
	if params.InquiryDate != nil {
		// Parse date string to time.Time
		date, err := time.Parse("2006-01-02", *params.InquiryDate)
		if err != nil {
			return errors.New("日期格式错误，应为 YYYY-MM-DD")
		}
		updates["inquiry_date"] = date
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.BasePriceInquiry{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *inquiryRepo) SoftDeleteInquiry(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.BasePriceInquiry{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *inquiryRepo) HardDeleteInquiry(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.BasePriceInquiry{}).Error
}

// ========== PriceInquiryItem Repository ==========

type inquiryItemRepo struct{ db *gorm.DB }

func (r *inquiryItemRepo) CreateInquiryItem(ctx context.Context, m *domain.PriceInquiryItem) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *inquiryItemRepo) GetInquiryItem(ctx context.Context, id string) (*domain.PriceInquiryItem, error) {
	var out domain.PriceInquiryItem
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *inquiryItemRepo) ListInquiryItems(ctx context.Context, inquiryID string, categoryID *string, page, pageSize int) ([]domain.PriceInquiryItem, int64, error) {
	var list []domain.PriceInquiryItem
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.PriceInquiryItem{}).
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

func (r *inquiryItemRepo) UpdateInquiryItem(ctx context.Context, params InquiryItemUpdateParams) error {
	updates := map[string]any{}
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
	if params.UpdateSpecNameSnap {
		if params.SpecNameSnap != nil {
			updates["spec_name_snap"] = *params.SpecNameSnap
		} else {
			updates["spec_name_snap"] = nil
		}
	}
	if params.UpdateUnitNameSnap {
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
	return r.db.WithContext(ctx).Model(&domain.PriceInquiryItem{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *inquiryItemRepo) SoftDeleteInquiryItem(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.PriceInquiryItem{}).
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
		Delete(&domain.PriceInquiryItem{}).Error
}

// ========== PriceMarketInquiry Repository ==========

type marketInquiryRepo struct{ db *gorm.DB }

func (r *marketInquiryRepo) CreateMarketInquiry(ctx context.Context, m *domain.PriceMarketInquiry) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *marketInquiryRepo) GetMarketInquiry(ctx context.Context, id string) (*domain.PriceMarketInquiry, error) {
	var out domain.PriceMarketInquiry
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *marketInquiryRepo) ListMarketInquiries(ctx context.Context, inquiryID, itemID *string, page, pageSize int) ([]domain.PriceMarketInquiry, int64, error) {
	var list []domain.PriceMarketInquiry
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.PriceMarketInquiry{}).
		Where("is_deleted = 0")

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
		Order("market_name_snap ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *marketInquiryRepo) UpdateMarketInquiry(ctx context.Context, params MarketInquiryUpdateParams) error {
	updates := map[string]any{}
	if params.UpdateMarketID {
		if params.MarketID != nil {
			updates["market_id"] = *params.MarketID
		} else {
			updates["market_id"] = nil
		}
	}
	if params.MarketNameSnap != nil {
		updates["market_name_snap"] = *params.MarketNameSnap
	}
	if params.UpdatePrice {
		if params.Price != nil {
			updates["price"] = *params.Price
		} else {
			updates["price"] = nil
		}
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.PriceMarketInquiry{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *marketInquiryRepo) SoftDeleteMarketInquiry(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.PriceMarketInquiry{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *marketInquiryRepo) HardDeleteMarketInquiry(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.PriceMarketInquiry{}).Error
}

// ========== PriceSupplierSettlement Repository ==========

type supplierSettlementRepo struct{ db *gorm.DB }

func (r *supplierSettlementRepo) CreateSupplierSettlement(ctx context.Context, m *domain.PriceSupplierSettlement) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *supplierSettlementRepo) GetSupplierSettlement(ctx context.Context, id string) (*domain.PriceSupplierSettlement, error) {
	var out domain.PriceSupplierSettlement
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *supplierSettlementRepo) ListSupplierSettlements(ctx context.Context, inquiryID, itemID *string, page, pageSize int) ([]domain.PriceSupplierSettlement, int64, error) {
	var list []domain.PriceSupplierSettlement
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.PriceSupplierSettlement{}).
		Where("is_deleted = 0")

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
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *supplierSettlementRepo) UpdateSupplierSettlement(ctx context.Context, params SupplierSettlementUpdateParams) error {
	updates := map[string]any{}
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
	if params.UpdateSettlement {
		if params.SettlementPrice != nil {
			updates["settlement_price"] = *params.SettlementPrice
		} else {
			updates["settlement_price"] = nil
		}
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.PriceSupplierSettlement{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *supplierSettlementRepo) SoftDeleteSupplierSettlement(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.PriceSupplierSettlement{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *supplierSettlementRepo) HardDeleteSupplierSettlement(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.PriceSupplierSettlement{}).Error
}
