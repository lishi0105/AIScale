package category

import (
	"errors"
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
	OrgID     string    `gorm:"column:org_id;type:char(36);not null;comment:所属机构ID"` // 注意 tag
	IsDeleted int       `gorm:"not null;default:0;comment:软删标记：0=有效,1=已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if c.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}

	// 1) 查 organ 的 code/sort（FOR UPDATE）
	orgCode, orgSort, err := utils.GetOrgCodeAndSortByID(tx.Statement.Context, tx, c.OrgID, true)
	if err != nil {
		return fmt.Errorf("查询 org code/sort 失败: %w", err)
	}
	if orgCode == "" {
		return errors.New("org.code 为空，无法派生 category.code")
	}
	if orgSort < 0 {
		return fmt.Errorf("org.sort 非法: %d", orgSort)
	}
	base := orgSort * 1000

	// 2) sort = org.sort*1000 + 最小缺口
	if c.Sort <= 0 {
		suf, err := utils.NextSortSuffix(tx, c.TableName(), c.OrgID, base, true)
		if err != nil {
			return err
		}
		c.Sort = base + suf
	}

	// 3) code = org.code + 三位后缀（各自独立找缺口）
	if c.Code == nil || (c.Code != nil && *c.Code == "") {
		suf, err := utils.NextCodeSuffixByPrefix(tx, c.TableName(), c.OrgID, orgCode, true)
		if err != nil {
			return err
		}
		auto := fmt.Sprintf("%s%03d", orgCode, suf)
		c.Code = &auto
	}

	// 4) 自动拼音
	if c.Pinyin == nil || (c.Pinyin != nil && *c.Pinyin == "") {
		if utils.ContainsChinese(c.Name) {
			p := utils.GeneratePinyin(c.Name)
			c.Pinyin = &p
		}
	}
	return nil
}

func (Category) TableName() string { return "base_category" }

// ---------- helpers ----------

// sort 段的最小缺口：仅统计本 organ (base, base+999] 的 sort
