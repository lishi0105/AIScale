package inquiry_item

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InquiryItem struct {
	ID                 string     `gorm:"primaryKey;type:char(36)"`
	InquiryID          string     `gorm:"column:inquiry_id;type:char(36);not null;comment:询价单ID"`
	GoodsID            string     `gorm:"column:goods_id;type:char(36);not null;comment:商品ID"`
	CategoryID         string     `gorm:"column:category_id;type:char(36);not null;comment:品类ID"`
	SpecID             *string    `gorm:"column:spec_id;type:char(36);comment:规格ID"`
	UnitID             *string    `gorm:"column:unit_id;type:char(36);comment:单位ID"`
	GoodsNameSnap      string     `gorm:"column:goods_name_snap;size:128;not null;comment:商品名称快照"`
	CategoryNameSnap   string     `gorm:"column:category_name_snap;size:64;not null;comment:品类名称快照"`
	SpecNameSnap       *string    `gorm:"column:spec_name_snap;size:32;comment:规格名称快照"`
	UnitNameSnap       *string    `gorm:"column:unit_name_snap;size:32;comment:单位名称快照"`
	GuidePrice         *float64   `gorm:"column:guide_price;type:decimal(12,2);comment:发改委指导价"`
	LastMonthAvgPrice  *float64   `gorm:"column:last_month_avg_price;type:decimal(12,2);comment:上月均价"`
	CurrentAvgPrice    *float64   `gorm:"column:current_avg_price;type:decimal(12,2);comment:本期均价"`
	Sort               int        `gorm:"not null;default:0;comment:排序码"`
	IsDeleted          int        `gorm:"column:is_deleted;not null;default:0;comment:软删：0=有效 1=删除"`
	CreatedAt          time.Time  `gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime"`
}

func (i *InquiryItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.NewString()
	}
	if i.InquiryID == "" {
		return errors.New("InquiryID(inquiry_id) 不能为空")
	}
	if i.GoodsID == "" {
		return errors.New("GoodsID(goods_id) 不能为空")
	}
	if i.CategoryID == "" {
		return errors.New("CategoryID(category_id) 不能为空")
	}
	return nil
}

func (InquiryItem) TableName() string { return "price_inquiry_item" }
