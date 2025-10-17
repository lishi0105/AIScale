package category

import (
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
	IsDeleted int       `gorm:"not null;default:0;comment:软删标记：0=有效,1=已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	code, err := utils.NextDictionaryCode(tx, "base_category", "04")
	if err != nil {
		return err
	}
	c.Code = &code
	
	// Auto-generate pinyin if name contains Chinese and pinyin is not provided
	if c.Pinyin == nil || *c.Pinyin == "" {
		if utils.ContainsChinese(c.Name) {
			py := utils.GeneratePinyin(c.Name)
			if py != "" {
				c.Pinyin = &py
			}
		}
	}
	
	return nil
}

func (Category) TableName() string { return "base_category" }
