package dict

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Unit struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:32;not null;uniqueIndex:uk_unit_name;comment:单位"`
	Sort      int       `gorm:"not null;default:0;index;comment:排序码"`
	IsDeleted int       `gorm:"not null;default:0;index;comment:是否已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (u *Unit) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	return nil
}

func (Unit) TableName() string { return "base_unit" }

type Spec struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:32;not null;uniqueIndex:uk_spec_name;comment:规格名称"`
	Sort      int       `gorm:"not null;default:0;index"`
	IsDeleted int       `gorm:"not null;default:0;index;comment:是否已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (s *Spec) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	return nil
}

func (Spec) TableName() string { return "base_spec" }

type MealTime struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:32;not null;uniqueIndex:uk_menu_meal_name;comment:餐次"`
	Sort      int       `gorm:"not null;default:0;index;comment:排序码"`
	IsDeleted int       `gorm:"not null;default:0;index;comment:是否已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (m *MealTime) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

func (MealTime) TableName() string { return "menu_meal" }
