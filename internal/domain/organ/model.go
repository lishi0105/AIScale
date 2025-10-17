package organ

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"hdzk.cn/foodapp/pkg/utils"
)

type Organ struct {
	ID          string    `gorm:"primaryKey;type:char(36)"`
	Name        string    `gorm:"size:128;not null;uniqueIndex:uk_org_parent_name;comment:组织名称"` // 注意 size 与 DB 一致
	Code        *string   `gorm:"size:64;not null;uniqueIndex:uk_org_code;comment:组织编码"`         // 改为非空 string
	Pinyin      *string   `gorm:"size:256;comment:拼音"`                                           // 新增拼音字段
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
	if m.Sort <= 0 {
		next, err := utils.NextColoumSort(tx, m.TableName())
		if err != nil {
			return err
		}
		m.Sort = next
	}

	// 2) 生成 code（两位字符串，来源于 sort）
	if m.Code == nil || (m.Code != nil && *m.Code == "") {
		code := codeFromSort(m.Sort)
		m.Code = &code
	}
	// 如果name为汉字且没有传入pinyin，自动生成pinyin
	if m.Pinyin == nil || (m.Pinyin != nil && *m.Pinyin == "") {
		if utils.ContainsChinese(m.Name) {
			pinyin := utils.GeneratePinyin(m.Name)
			m.Pinyin = &pinyin
		}
	}
	return nil
}

func codeFromSort(sort int) string {
	return fmt.Sprintf("%02d", sort)
}

func (Organ) TableName() string { return "base_org" }

type ListQuery struct {
	NameLike string // 模糊匹配
	Deleted  *int
	Role     *int
	Limit    int
	Offset   int
}
