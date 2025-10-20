package inquiry

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BasePriceInquiry struct {
	ID               string     `gorm:"primaryKey;type:char(36);comment:UUID"`
	InquiryTitle     string     `gorm:"column:inquiry_title;size:64;not null;comment:询价单标题"`
	InquiryDate      time.Time  `gorm:"column:inquiry_date;type:date;not null;comment:询价单日期（业务日）"`
	Market1          *string    `gorm:"column:market_1;size:128;comment:市场1"`
	Market2          *string    `gorm:"column:market_2;size:128;comment:市场2"`
	Market3          *string    `gorm:"column:market_3;size:128;comment:市场3"`
	OrgID            string     `gorm:"column:org_id;type:char(36);not null;comment:中队ID"`
	IsDeleted        int        `gorm:"column:is_deleted;type:tinyint(1);not null;default:0;comment:软删：0=有效 1=删除"`
	InquiryStartDate time.Time  `gorm:"column:inquiry_start_date;type:datetime;not null;comment:开始时间"`
	InquiryEndDate   time.Time  `gorm:"column:inquiry_end_date;type:datetime;not null;comment:结束时间"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
	ActiveTitle      *string    `gorm:"column:active_title;type:varchar(64);comment:仅对未删除行生效的唯一：利用 NULL 不参与唯一的特性"`
}

func (b *BasePriceInquiry) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	
	// 验证必填字段
	if b.InquiryTitle == "" {
		return errors.New("inquiry_title 不能为空")
	}
	if b.OrgID == "" {
		return errors.New("org_id 不能为空")
	}
	
	// 验证时间逻辑
	if b.InquiryEndDate.Before(b.InquiryStartDate) {
		return errors.New("结束时间必须晚于开始时间")
	}
	
	// 设置业务日期（如果未设置）
	if b.InquiryDate.IsZero() {
		b.InquiryDate = b.InquiryStartDate
	}
	
	// 设置软删除标记
	if b.IsDeleted == 0 {
		b.ActiveTitle = &b.InquiryTitle
	} else {
		b.ActiveTitle = nil
	}
	
	return nil
}

func (b *BasePriceInquiry) BeforeUpdate(tx *gorm.DB) error {
	// 更新时重新计算 ActiveTitle
	if b.IsDeleted == 0 {
		b.ActiveTitle = &b.InquiryTitle
	} else {
		b.ActiveTitle = nil
	}
	
	// 验证时间逻辑
	if b.InquiryEndDate.Before(b.InquiryStartDate) {
		return errors.New("结束时间必须晚于开始时间")
	}
	
	return nil
}

func (BasePriceInquiry) TableName() string {
	return "base_price_inquiry"
}

// 查询参数结构
type InquiryQueryParams struct {
	Keyword    string
	OrgID      string
	StartDate  *time.Time
	EndDate    *time.Time
	Market1    *string
	Market2    *string
	Market3    *string
	Page       int
	PageSize   int
}

// 创建参数结构
type CreateInquiryParams struct {
	InquiryTitle     string
	InquiryDate      time.Time
	Market1          *string
	Market2          *string
	Market3          *string
	OrgID            string
	InquiryStartDate time.Time
	InquiryEndDate   time.Time
}

// 更新参数结构
type UpdateInquiryParams struct {
	ID               string
	InquiryTitle     *string
	InquiryDate      *time.Time
	Market1          *string
	Market2          *string
	Market3          *string
	InquiryStartDate *time.Time
	InquiryEndDate   *time.Time
}