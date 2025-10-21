package market

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	utils "hdzk.cn/foodapp/pkg/utils"
)

type Market struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:64;not null;comment:市场名称"`
	OrgID     string    `gorm:"column:org_id;type:char(36);not null;comment:中队ID"`
	Code      *string   `gorm:"size:64;uniqueIndex:uq_market_code;comment:市场编码"`
	Sort      int       `gorm:"not null;default:0;comment:排序码"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效 1=已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (m *Market) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}

	// 1) 查 organ 的 code/sort（FOR UPDATE）
	orgCode, orgSort, err := utils.GetOrgCodeAndSortByID(tx.Statement.Context, tx, m.OrgID, true)
	if err != nil {
		return fmt.Errorf("查询 org code/sort 失败: %w", err)
	}
	if orgCode == "" {
		return errors.New("org.code 为空，无法派生 market.code")
	}
	if orgSort < 0 {
		return fmt.Errorf("org.sort 非法: %d", orgSort)
	}
	base := orgSort * 1000

	// 2) sort = org.sort*1000 + 最小缺口
	if m.Sort <= 0 {
		suf, err := utils.NextSortSuffix(tx, m.TableName(), m.OrgID, base, true)
		if err != nil {
			return err
		}
		m.Sort = base + suf
	}

	// 3) code = org.code + 三位后缀（各自独立找缺口）
	if m.Code == nil || (m.Code != nil && *m.Code == "") {
		suf, err := utils.NextCodeSuffixByPrefix(tx, m.TableName(), m.OrgID, orgCode, true)
		if err != nil {
			return err
		}
		auto := fmt.Sprintf("%s%03d", orgCode, suf)
		m.Code = &auto
	}

	return nil
}

func (Market) TableName() string { return "base_market" }
