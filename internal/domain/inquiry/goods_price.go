package inquiry

import "time"

// GoodsPrice maps to table `base_goods_price`.
// Stores per-supplier unit price for a goods item under a price inquiry.
type GoodsPrice struct {
    ID         string    `gorm:"primaryKey;type:char(36)"`
    GoodsID    string    `gorm:"column:goods_id;type:char(36);not null"`
    SupplierID string    `gorm:"column:supplier_id;type:char(36);not null"`
    InquiryID  string    `gorm:"column:inquiry_id;type:char(36);not null"`
    UnitPrice  float64   `gorm:"column:unit_price;not null"`
    FloatRatio float64   `gorm:"column:float_ratio;not null;default:1.0000"`
    OrgID      *string   `gorm:"column:org_id;type:char(36)"`
    IsDeleted  int       `gorm:"column:is_deleted;not null;default:0"`
    CreatedAt  time.Time `gorm:"autoCreateTime"`
    UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

func (GoodsPrice) TableName() string { return "base_goods_price" }
