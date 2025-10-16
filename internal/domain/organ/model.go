package organ

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organ struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:64;not null;uniqueIndex:uk_organ_name;comment:中队名称"`
	Code      *string   `gorm:"size:32;uniqueIndex:uk_organ_code;comment:中队编码"`
	Leader    string    `gorm:"size:64;comment:负责人"`
	Phone     string    `gorm:"size:32;comment:联系电话"`
	Sort      int       `gorm:"not null;default:0;index;comment:排序码"`
	Status    int       `gorm:"not null;default:1;index;comment:状态 1启用 0停用"`
	Remark    string    `gorm:"size:255;comment:备注"`
	IsDeleted int       `gorm:"not null;default:0;index;comment:是否已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (o *Organ) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.NewString()
	}
	o.Name = strings.TrimSpace(o.Name)
	o.Leader = strings.TrimSpace(o.Leader)
	o.Phone = strings.TrimSpace(o.Phone)
	o.Remark = strings.TrimSpace(o.Remark)
	return nil
}

func (Organ) TableName() string { return "organ" }

type ListQuery struct {
	Keyword string
	Status  *int
	Limit   int
	Offset  int
}
