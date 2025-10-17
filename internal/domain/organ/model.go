package organ

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"hdzk.cn/foodapp/pkg/utils"
)

type Organ struct {
	ID          string    `gorm:"primaryKey;type:char(36)"`
	Name        string    `gorm:"size:128;not null;uniqueIndex:uk_org_parent_name;comment:组织名称"` // 注意 size 与 DB 一致
	Code        *string   `gorm:"size:64;not null;uniqueIndex:uk_org_code;comment:组织编码"`         // 改为非空 string
	Pinyin      *string   `gorm:"size:256;comment:拼音"`                                              // 新增拼音字段
	ParentID    *string   `gorm:"column:parent_id;type:char(36);not null;index;comment:上级组织ID"`  // 重命名 + 映射
	Description string    `gorm:"type:text;not null;comment:组织描述"`
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
		code, err := utils.NextDictionaryCode(tx, "base_org", "01")
		if err != nil {
			return err
		}
		m.Code = &code
	}
	// 如果name为汉字且没有传入pinyin，自动生成pinyin
	if m.Pinyin == nil || (m.Pinyin != nil && *m.Pinyin == "") {
		if containsChinese(m.Name) {
			pinyin := convertToPinyin(m.Name)
			m.Pinyin = &pinyin
		}
	}
	return nil
}

func (Organ) TableName() string { return "base_org" }

type ListQuery struct {
	NameLike string // 模糊匹配
	Deleted  *int
	Role     *int
	Limit    int
	Offset   int
}
