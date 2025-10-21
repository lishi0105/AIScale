package supplier_settlement

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SupplierSettlement struct {
	ID                 string    `gorm:"primaryKey;type:char(36)"`
	InquiryID          string    `gorm:"column:inquiry_id;type:char(36);not null;comment:询价单ID"`
	ItemID             string    `gorm:"column:item_id;type:char(36);not null;comment:询价明细ID"`
	SupplierID         *string   `gorm:"column:supplier_id;type:char(36);comment:供应商ID"`
	SupplierNameSnap   string    `gorm:"column:supplier_name_snap;size:128;not null;comment:供应商名称快照"`
	FloatRatioSnap     float64   `gorm:"column:float_ratio_snap;type:decimal(6,4);not null;comment:浮动比例快照"`
	SettlementPrice    float64   `gorm:"column:settlement_price;type:decimal(12,2);not null;comment:本期结算价"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

func (s *SupplierSettlement) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.InquiryID == "" {
		return errors.New("InquiryID(inquiry_id) 不能为空")
	}
	if s.ItemID == "" {
		return errors.New("ItemID(item_id) 不能为空")
	}
	if s.SupplierNameSnap == "" {
		return errors.New("SupplierNameSnap(supplier_name_snap) 不能为空")
	}
	return nil
}

func (SupplierSettlement) TableName() string { return "price_supplier_settlement" }
