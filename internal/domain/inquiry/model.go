package inquiry

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PriceInquiry represents a price inquiry header record.
// It maps to table `base_price_inquiry`.
type PriceInquiry struct {
	ID           string    `gorm:"primaryKey;type:char(36)"`
	InquiryTitle string    `gorm:"column:inquiry_title;size:64;not null;comment:询价单标题"`
	InquiryDate  time.Time `gorm:"column:inquiry_date;type:date;not null;comment:询价单日期（业务日）"`

	Market1 *string `gorm:"column:market_1;size:128;comment:市场1"`
	Market2 *string `gorm:"column:market_2;size:128;comment:市场2"`
	Market3 *string `gorm:"column:market_3;size:128;comment:市场3"`

	OrgID string `gorm:"column:org_id;type:char(36);not null;comment:中队ID"`

	IsDeleted int `gorm:"column:is_deleted;not null;default:0;comment:软删：0=有效 1=删除"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (p *PriceInquiry) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if p.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}
	return nil
}

func (PriceInquiry) TableName() string { return "base_price_inquiry" }

// GoodsAvgDetail represents the average price detail for a goods item.
// It maps to table `base_goods_avg_detail`.
type GoodsAvgDetail struct {
	ID           string   `gorm:"primaryKey;type:char(36)"`
	GoodsID      string   `gorm:"column:goods_id;type:char(36);not null;comment:商品Id"`
	GuidePrice   *float64 `gorm:"column:guide_price;type:decimal(10,2);comment:指导价"`
	Market1Price *float64 `gorm:"column:market1_price;type:decimal(10,2);comment:市场1价格"`
	Market2Price *float64 `gorm:"column:market2_price;type:decimal(10,2);comment:市场2价格"`
	Market3Price *float64 `gorm:"column:market3_price;type:decimal(10,2);comment:市场3价格"`
	AvgPrice     *float64 `gorm:"column:avg_price;type:decimal(10,2);comment:商品均价"`
	InquiryID    string   `gorm:"column:inquiry_id;type:char(36);not null;comment:询价记录Id"`
	OrgID        *string  `gorm:"column:org_id;type:char(36);comment:中队Id"`
	IsDeleted    int      `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效,1=已删除"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (g *GoodsAvgDetail) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.NewString()
	}
	return nil
}

func (GoodsAvgDetail) TableName() string { return "base_goods_avg_detail" }

// GoodsPrice represents the price for a goods item from a supplier.
// It maps to table `base_goods_price`.
type GoodsPrice struct {
	ID          string    `gorm:"primaryKey;type:char(36)"`
	GoodsID     string    `gorm:"column:goods_id;type:char(36);not null;comment:商品ID"`
	SupplierID  string    `gorm:"column:supplier_id;type:char(36);not null;comment:供应商ID"`
	InquiryID   string    `gorm:"column:inquiry_id;type:char(36);not null;comment:询价记录ID"`
	UnitPrice   float64   `gorm:"column:unit_price;type:decimal(10,2);not null;comment:商品单价"`
	FloatRatio  float64   `gorm:"column:float_ratio;type:decimal(6,4);not null;default:1.0000;comment:浮动比例快照"`
	OrgID       *string   `gorm:"column:org_id;type:char(36);comment:中队ID"`
	IsDeleted   int       `gorm:"column:is_deleted;not null;default:0;comment:软删：0=有效 1=删除"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (g *GoodsPrice) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.NewString()
	}
	return nil
}

func (GoodsPrice) TableName() string { return "base_goods_price" }
