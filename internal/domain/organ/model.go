package organ

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"hdzk.cn/foodapp/pkg/utils"
)

type Organ struct {
	ID          string    `gorm:"primaryKey;type:char(36)"`
	Name        string    `gorm:"size:64;not null;uniqueIndex:uk_organ_name;comment:组织名称"`
	Code        *string   `gorm:"size:32;uniqueIndex:uk_organ_code;comment:组织编码"`
	Parent      string    `gorm:"size:36;comment:上级组织ID"`
	Description string    `gorm:"type:text;comment:组织描述"`
	Sort        int       `gorm:"not null;default:0;index;comment:排序码"`
	IsDeleted   int       `gorm:"not null;default:0;index;comment:是否已删除"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (m *Organ) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	// 仅当 Code 为空或空字符串时自动生成
	if m.Code == nil || (m.Code != nil && *m.Code == "") {
		code, err := utils.NextDictionaryCode(tx, "organization", "01")
		if err != nil {
			return err
		}
		m.Code = &code
	}
	return nil
}

func (Organ) TableName() string { return "organization" }

type ListQuery struct {
	NameLike string // 模糊匹配
	Deleted  *int
	Role     *int
	Limit    int
	Offset   int
}
