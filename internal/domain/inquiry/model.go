package inquiry

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Inquiry struct {
	ID               string     `gorm:"primaryKey;type:char(36)"`
	InquiryTitle     string     `gorm:"column:inquiry_title;size:64;not null;comment:询价单标题"`
	InquiryDate      time.Time  `gorm:"column:inquiry_date;type:date;not null;comment:询价单日期（业务日）"`
	Market1          *string    `gorm:"column:market_1;size:128;comment:市场1"`
	Market2          *string    `gorm:"column:market_2;size:128;comment:市场2"`
	Market3          *string    `gorm:"column:market_3;size:128;comment:市场3"`
	OrgID            string     `gorm:"column:org_id;type:char(36);not null;comment:中队ID"`
	IsDeleted        int        `gorm:"column:is_deleted;not null;default:0;comment:软删：0=有效 1=删除"`
	InquiryStartDate time.Time  `gorm:"column:inquiry_start_date;not null;comment:开始时间"`
	InquiryEndDate   time.Time  `gorm:"column:inquiry_end_date;not null;comment:结束时间"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime"`
	// ActiveTitle 是计算列，不需要在结构体中定义
}

func (i *Inquiry) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.NewString()
	}
	if i.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}
	// 验证结束时间必须晚于开始时间
	if !i.InquiryEndDate.After(i.InquiryStartDate) {
		return errors.New("结束时间必须晚于开始时间")
	}
	return nil
}

func (Inquiry) TableName() string { return "base_price_inquiry" }
