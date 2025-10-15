package account

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Account struct {
	ID           string         `gorm:"primaryKey;type:char(36)"`
	Username     string         `gorm:"size:64;uniqueIndex:uk_username;not null;comment:登录名"`
	Email        string         `gorm:"size:128;index:idx_email;comment:邮箱"` // 可选就别 not null
	PasswordHash string         `gorm:"size:128;not null;comment:加密后的密码" json:"-"`
	Status       int            `gorm:"not null;default:0;comment:0 正常 1 禁用等"`
	Role         int            `gorm:"not null;default:0;comment:0 普通用户 1 超级用户"`
	LastLoginAt  *time.Time     `gorm:"comment:最近登录时间"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}

type ListQuery struct {
	UsernameLike string // 模糊匹配
	EmailLike    string
	Status       *int
	Role         *int
	Limit        int
	Offset       int
}
