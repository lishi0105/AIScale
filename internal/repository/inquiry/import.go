package inquiry

import (
    "context"
    "fmt"
    "strings"

    "gorm.io/gorm"
    cat "hdzk.cn/foodapp/internal/domain/category"
    dict "hdzk.cn/foodapp/internal/domain/dict"
    goodsdom "hdzk.cn/foodapp/internal/domain/goods"
    inq "hdzk.cn/foodapp/internal/domain/inquiry"
    supdom "hdzk.cn/foodapp/internal/domain/supplier"
)

// ImportRepository defines methods needed during Excel import.
type ImportRepository interface {
    BeginTx(ctx context.Context) *gorm.DB
    // find-or-create helpers within org
    GetOrCreateCategory(ctx context.Context, tx *gorm.DB, orgID, name string) (*cat.Category, error)
    GetOrCreateSpec(ctx context.Context, tx *gorm.DB, name string) (*dict.Spec, error)
    GetOrCreateUnit(ctx context.Context, tx *gorm.DB, name string) (*dict.Unit, error)
    GetOrCreateGoods(ctx context.Context, tx *gorm.DB, orgID, name, categoryID, specID, unitID string) (*goodsdom.Goods, error)
    GetOrCreateSupplierWithRatio(ctx context.Context, tx *gorm.DB, orgID, name string, ratio float64) (*supdom.Supplier, error)

    CreateInquiry(ctx context.Context, tx *gorm.DB, inq *inq.PriceInquiry) error
    CreateGoodsAvgDetail(ctx context.Context, tx *gorm.DB, m *inq.GoodsAvgDetail) error
    UpsertGoodsPrice(ctx context.Context, tx *gorm.DB, m *inq.GoodsPrice) error
}

// repoImport is a gorm-based implementation that reuses existing tables and unique constraints.
type repoImport struct { db *gorm.DB }

func NewImportRepository(db *gorm.DB) ImportRepository { return &repoImport{db: db} }

// expose DB for helper usage in service
func (r *repoImport) DB() *gorm.DB { return r.db }

func (r *repoImport) BeginTx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx).Begin() }

func (r *repoImport) GetOrCreateCategory(ctx context.Context, tx *gorm.DB, orgID, name string) (*cat.Category, error) {
    var out cat.Category
    q := tx.WithContext(ctx).Where("org_id = ? AND name = ? AND is_deleted = 0", orgID, strings.TrimSpace(name)).First(&out)
    if q.Error == nil { return &out, nil }
    if q.Error != nil && q.Error != gorm.ErrRecordNotFound { return nil, q.Error }
    m := &cat.Category{Name: strings.TrimSpace(name), OrgID: orgID}
    if err := tx.WithContext(ctx).Create(m).Error; err != nil { return nil, err }
    return m, nil
}

func (r *repoImport) GetOrCreateSpec(ctx context.Context, tx *gorm.DB, name string) (*dict.Spec, error) {
    var out dict.Spec
    q := tx.WithContext(ctx).Where("name = ? AND is_deleted = 0", strings.TrimSpace(name)).First(&out)
    if q.Error == nil { return &out, nil }
    if q.Error != nil && q.Error != gorm.ErrRecordNotFound { return nil, q.Error }
    m := &dict.Spec{Name: strings.TrimSpace(name)}
    if err := tx.WithContext(ctx).Create(m).Error; err != nil { return nil, err }
    return m, nil
}

func (r *repoImport) GetOrCreateUnit(ctx context.Context, tx *gorm.DB, name string) (*dict.Unit, error) {
    var out dict.Unit
    q := tx.WithContext(ctx).Where("name = ? AND is_deleted = 0", strings.TrimSpace(name)).First(&out)
    if q.Error == nil { return &out, nil }
    if q.Error != nil && q.Error != gorm.ErrRecordNotFound { return nil, q.Error }
    m := &dict.Unit{Name: strings.TrimSpace(name)}
    if err := tx.WithContext(ctx).Create(m).Error; err != nil { return nil, err }
    return m, nil
}

func (r *repoImport) GetOrCreateGoods(ctx context.Context, tx *gorm.DB, orgID, name, categoryID, specID, unitID string) (*goodsdom.Goods, error) {
    var out goodsdom.Goods
    q := tx.WithContext(ctx).Where("org_id=? AND name=? AND spec_id=? AND unit_id=? AND is_deleted=0", orgID, strings.TrimSpace(name), specID, unitID).First(&out)
    if q.Error == nil { return &out, nil }
    if q.Error != nil && q.Error != gorm.ErrRecordNotFound { return nil, q.Error }
    m := &goodsdom.Goods{Name: strings.TrimSpace(name), OrgID: orgID, CategoryID: categoryID, SpecID: specID, UnitID: unitID}
    if err := tx.WithContext(ctx).Create(m).Error; err != nil { return nil, err }
    return m, nil
}

func (r *repoImport) GetOrCreateSupplierWithRatio(ctx context.Context, tx *gorm.DB, orgID, name string, ratio float64) (*supdom.Supplier, error) {
    var out supdom.Supplier
    q := tx.WithContext(ctx).Where("org_id=? AND name=? AND is_deleted=0", orgID, strings.TrimSpace(name)).First(&out)
    if q.Error != nil && q.Error != gorm.ErrRecordNotFound { return nil, q.Error }
    if q.Error == nil {
        // update ratio if different
        if out.FloatRatio != ratio {
            if err := tx.WithContext(ctx).Model(&supdom.Supplier{}).Where("id=?", out.ID).Update("float_ratio", ratio).Error; err != nil { return nil, err }
            out.FloatRatio = ratio
        }
        return &out, nil
    }
    m := &supdom.Supplier{Name: strings.TrimSpace(name), OrgID: orgID, Description: " ", FloatRatio: ratio}
    if err := tx.WithContext(ctx).Create(m).Error; err != nil { return nil, err }
    return m, nil
}

func (r *repoImport) CreateInquiry(ctx context.Context, tx *gorm.DB, inqHdr *inq.PriceInquiry) error {
    return tx.WithContext(ctx).Create(inqHdr).Error
}

func (r *repoImport) CreateGoodsAvgDetail(ctx context.Context, tx *gorm.DB, m *inq.GoodsAvgDetail) error {
    return tx.WithContext(ctx).Create(m).Error
}

func (r *repoImport) UpsertGoodsPrice(ctx context.Context, tx *gorm.DB, m *inq.GoodsPrice) error {
    // unique (inquiry_id, supplier_id, goods_id)
    var exists inq.GoodsPrice
    err := tx.WithContext(ctx).Where("inquiry_id=? AND supplier_id=? AND goods_id=? AND is_deleted=0", m.InquiryID, m.SupplierID, m.GoodsID).First(&exists).Error
    if err == nil {
        return tx.WithContext(ctx).Model(&inq.GoodsPrice{}).Where("id=?", exists.ID).Updates(map[string]any{
            "unit_price": m.UnitPrice,
            "float_ratio": m.FloatRatio,
            "org_id": m.OrgID,
        }).Error
    }
    if err != nil && err != gorm.ErrRecordNotFound { return err }
    return tx.WithContext(ctx).Create(m).Error
}

func (r *repoImport) tx(ctx context.Context) (*gorm.DB, error) {
    if r.db == nil { return nil, fmt.Errorf("nil db") }
    return r.db.WithContext(ctx).Begin(), nil
}
