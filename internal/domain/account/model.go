package account

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Account struct {
	ID           string     `gorm:"primaryKey;type:char(36)"`
	Username     string     `gorm:"size:64;uniqueIndex:uk_account_username;not null;comment:用户名"`
	PasswordHash string     `gorm:"size:255;not null;comment:密码Hash" json:"-"`
	OrgID        string     `gorm:"size:36;not null;comment:所属机构ID"`
	Description  *string    `gorm:"type:text;comment:描述"`
	Role         int        `gorm:"not null;default:0;comment:角色 1管理员 0用户"`
	IsDeleted    int        `gorm:"not null;default:0;index;comment:是否已删除 0未删除 1已删除" json:"-"`
	LastLoginAt  *time.Time `gorm:"comment:最后登录时间"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime"`
}

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}

func (Account) TableName() string { return "account" }

type ListQuery struct {
	UsernameLike string // 模糊匹配
	Deleted      *int
	Role         *int
	Limit        int
	Offset       int
}
