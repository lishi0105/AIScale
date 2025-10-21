package market_inquiry

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MarketInquiry struct {
	ID             string   `gorm:"primaryKey;type:char(36)"`
	InquiryID      string   `gorm:"column:inquiry_id;type:char(36);not null;comment:base_price_inquiry.id"`
	ItemID         string   `gorm:"column:item_id;type:char(36);not null;comment:price_inquiry_item.id"`
	MarketID       *string  `gorm:"column:market_id;type:char(36);comment:base_market.id（可为空，仅保存名称）"`
	MarketNameSnap string   `gorm:"column:market_name_snap;size:64;not null;comment:市场名称快照"`
	Price          *float64 `gorm:"column:price;type:decimal(12,2);comment:该市场的单价"`
	IsDeleted      int      `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效,1=删除"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func (m *MarketInquiry) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.InquiryID == "" {
		return errors.New("InquiryID(inquiry_id) 不能为空")
	}
	if m.ItemID == "" {
		return errors.New("ItemID(item_id) 不能为空")
	}
	if m.MarketNameSnap == "" {
		return errors.New("MarketNameSnap(market_name_snap) 不能为空")
	}
	return nil
}

func (MarketInquiry) TableName() string { return "price_market_inquiry" }