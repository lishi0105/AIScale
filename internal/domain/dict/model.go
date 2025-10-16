package dict

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	utils "hdzk.cn/foodapp/pkg/utils"
)

type Unit struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:32;not null;uniqueIndex:uk_unit_name;comment:单位"`
	Code      *string   `gorm:"size:32;uniqueIndex:uk_unit_code;comment:单位编码"`
	Sort      int       `gorm:"not null;default:0;index;comment:排序码"`
	IsDeleted int       `gorm:"not null;default:0;index;comment:是否已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (u *Unit) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	code, err := utils.NextDictionaryCode(tx, "base_unit", "01")
	if err != nil {
		return err
	}
	u.Code = &code
	return nil
}

func (Unit) TableName() string { return "base_unit" }

type Spec struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:32;not null;uniqueIndex:uk_spec_name;comment:规格名称"`
	Code      *string   `gorm:"size:32;uniqueIndex:uk_spec_code;comment:规格编码"`
	Sort      int       `gorm:"not null;default:0;index"`
	IsDeleted int       `gorm:"not null;default:0;index;comment:是否已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (s *Spec) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	code, err := utils.NextDictionaryCode(tx, "base_spec", "02")
	if err != nil {
		return err
	}
	s.Code = &code
	return nil
}

func (Spec) TableName() string { return "base_spec" }

type MealTime struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:32;not null;uniqueIndex:uk_menu_meal_name;comment:餐次"`
	Code      *string   `gorm:"size:32;uniqueIndex:uk_meal_code;comment:餐次编码"`
	Sort      int       `gorm:"not null;default:0;index;comment:排序码"`
	IsDeleted int       `gorm:"not null;default:0;index;comment:是否已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (m *MealTime) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	code, err := utils.NextDictionaryCode(tx, "menu_meal", "03")
	if err != nil {
		return err
	}
	m.Code = &code
	return nil
}

func (MealTime) TableName() string { return "menu_meal" }
