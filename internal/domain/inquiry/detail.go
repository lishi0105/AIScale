package inquiry

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GoodsAvgDetail struct {
	ID           string   `gorm:"primaryKey;type:char(36)"`
	GoodsID      string   `gorm:"column:goods_id;type:char(36);not null"`
	GuidePrice   *float64 `gorm:"column:guide_price;type:decimal(10,2)"`
	Market1Price *float64 `gorm:"column:market1_price;type:decimal(10,2)"`
	Market2Price *float64 `gorm:"column:market2_price;type:decimal(10,2)"`
	Market3Price *float64 `gorm:"column:market3_price;type:decimal(10,2)"`
	InquiryID    string   `gorm:"column:inquiry_id;type:char(36);not null"`
	OrgID        string   `gorm:"column:org_id;type:char(36);not null"`
	IsDeleted    int      `gorm:"column:is_deleted;not null;default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (g *GoodsAvgDetail) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.NewString()
	}
	if g.GoodsID == "" {
		return errors.New("goods_id 不能为空")
	}
	if g.InquiryID == "" {
		return errors.New("inquiry_id 不能为空")
	}
	if g.OrgID == "" {
		return errors.New("org_id 不能为空")
	}
	return nil
}

func (GoodsAvgDetail) TableName() string { return "base_goods_avg_detail" }

type GoodsPrice struct {
	ID         string  `gorm:"primaryKey;type:char(36)"`
	GoodsID    string  `gorm:"column:goods_id;type:char(36);not null"`
	SupplierID string  `gorm:"column:supplier_id;type:char(36);not null"`
	InquiryID  string  `gorm:"column:inquiry_id;type:char(36);not null"`
	UnitPrice  float64 `gorm:"column:unit_price;type:decimal(10,2);not null"`
	FloatRatio float64 `gorm:"column:float_ratio;type:decimal(6,4);not null;default:1.0000"`
	OrgID      string  `gorm:"column:org_id;type:char(36);not null"`
	IsDeleted  int     `gorm:"column:is_deleted;not null;default:0"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (g *GoodsPrice) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.NewString()
	}
	if g.GoodsID == "" {
		return errors.New("goods_id 不能为空")
	}
	if g.SupplierID == "" {
		return errors.New("supplier_id 不能为空")
	}
	if g.InquiryID == "" {
		return errors.New("inquiry_id 不能为空")
	}
	if g.OrgID == "" {
		return errors.New("org_id 不能为空")
	}
	if g.FloatRatio <= 0 {
		return errors.New("float_ratio 必须大于0")
	}
	return nil
}

func (GoodsPrice) TableName() string { return "base_goods_price" }
