package price_inquiry

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PriceInquiry struct {
	ID            string    `gorm:"primaryKey;type:char(36)"`
	OrgID         string    `gorm:"column:org_id;type:char(36);not null;comment:中队ID"`
	InquiryTitle  string    `gorm:"column:inquiry_title;size:64;not null;comment:询价单标题"`
	InquiryDate   time.Time `gorm:"column:inquiry_date;type:date;not null;comment:业务日期"`
	InquiryYear   int16     `gorm:"column:inquiry_year;comment:年份（生成列）"`
	InquiryMonth  int8      `gorm:"column:inquiry_month;comment:月份（生成列）"`
	InquiryTenDay int8      `gorm:"column:inquiry_ten_day;comment:旬（生成列）"`
	IsDeleted     int       `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效,1=删除"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (p *PriceInquiry) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if p.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}
	if p.InquiryTitle == "" {
		return errors.New("InquiryTitle(inquiry_title) 不能为空")
	}
	if p.InquiryDate.IsZero() {
		return errors.New("InquiryDate(inquiry_date) 不能为空")
	}
	return nil
}

func (PriceInquiry) TableName() string { return "base_price_inquiry" }