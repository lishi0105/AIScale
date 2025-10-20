package inquiry

import "time"

// GoodsAvgDetail maps to table `base_goods_avg_detail`.
// It stores three market prices for a goods item under a specific inquiry.
// The database defines `avg_price` as a generated column, so it's omitted here.
type GoodsAvgDetail struct {
    ID           string    `gorm:"primaryKey;type:char(36)"`
    GoodsID      string    `gorm:"column:goods_id;type:char(36);not null"`
    GuidePrice   *float64  `gorm:"column:guide_price"`
    Market1Price *float64  `gorm:"column:market1_price"`
    Market2Price *float64  `gorm:"column:market2_price"`
    Market3Price *float64  `gorm:"column:market3_price"`
    InquiryID    string    `gorm:"column:inquiry_id;type:char(36);not null"`
    OrgID        *string   `gorm:"column:org_id;type:char(36)"`
    IsDeleted    int       `gorm:"column:is_deleted;not null;default:0"`
    CreatedAt    time.Time `gorm:"autoCreateTime"`
    UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (GoodsAvgDetail) TableName() string { return "base_goods_avg_detail" }
