package category

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	utils "hdzk.cn/foodapp/pkg/utils"
)

type Category struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:64;not null;uniqueIndex:uq_category_name;comment:品类名称"`
	Code      *string   `gorm:"size:64;uniqueIndex:uq_category_code;comment:品类编码"`
	Pinyin    *string   `gorm:"size:64;comment:拼音"`
	IsDeleted int       `gorm:"not null;default:0;index;comment:是否已删除"`
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
	return nil
}

func (Category) TableName() string { return "base_category" }