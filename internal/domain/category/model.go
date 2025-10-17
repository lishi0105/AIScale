package category

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	utils "hdzk.cn/foodapp/pkg/utils"
)

type Category struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:64;not null;uniqueIndex:uq_category_name;comment:品类名称（唯一）"`
	Code      *string   `gorm:"size:64;uniqueIndex:uq_category_code;comment:品类编码（可选，建议唯一）"`
	Pinyin    *string   `gorm:"size:64;comment:拼音（可选，用于搜索）"`
	Sort      int       `gorm:"not null;default:0;index;comment:排序值"`
	OrgID     string    `gorm:"type:char(36);not null;comment:所属机构ID"`
	IsDeleted int       `gorm:"not null;default:0;comment:软删标记：0=有效,1=已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if c.Sort <= 0 {
		next, err := utils.NextColoumSort(tx, c.TableName())
		if err != nil {
			return err
		}
		c.Sort = next
	}

	// 2) 生成 code（两位字符串，来源于 sort）
	if c.Code == nil || (c.Code != nil && *c.Code == "") {
		code := codeFromSort(c.Sort)
		c.Code = &code
	}
	// 如果name为汉字且没有传入pinyin，自动生成pinyin
	if c.Pinyin == nil || (c.Pinyin != nil && *c.Pinyin == "") {
		if utils.ContainsChinese(c.Name) {
			pinyin := utils.GeneratePinyin(c.Name)
			c.Pinyin = &pinyin
		}
	}
	return nil
}

func codeFromSort(sort int) string {
	return fmt.Sprintf("%02d", sort)
}

func (Category) TableName() string { return "base_category" }
