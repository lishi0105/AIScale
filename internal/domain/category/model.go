package category

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	utils "hdzk.cn/foodapp/pkg/utils"
)

type Category struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:64;not null;uniqueIndex:uq_category_name;comment:品类名称（唯一）"`
	Code      *string   `gorm:"size:64;uniqueIndex:uq_category_code;comment:品类编码（可选，建议唯一）"`
	Pinyin    *string   `gorm:"size:64;comment:拼音（可选，用于搜索）"`
	Sort      int       `gorm:"not null;default:0;index;comment:排序值"`
	TeamID    string    `gorm:"column:team_id;type:char(36);not null;comment:所属机构ID"` // 注意 tag
	IsDeleted int       `gorm:"not null;default:0;comment:软删标记：0=有效,1=已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if c.TeamID == "" {
		return errors.New("TeamID(team_id) 不能为空")
	}

	// 1) 查 team 的 code/sort（FOR UPDATE）
	orgCode, orgSort, err := utils.GetOrgCodeAndSortByID(tx.Statement.Context, tx, c.TeamID, true)
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
		suf, err := nextSortSuffix(tx, c.TeamID, base, true)
		if err != nil {
			return err
		}
		c.Sort = base + suf
	}

	// 3) code = org.code + 三位后缀（各自独立找缺口）
	if c.Code == nil || (c.Code != nil && *c.Code == "") {
		suf, err := nextCodeSuffixByPrefix(tx, c.TeamID, orgCode, true)
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

// sort 段的最小缺口：仅统计本 team、未删除、落在 (base, base+999] 的 sort
func nextSortSuffix(tx *gorm.DB, teamID string, base int, forUpdate bool) (int, error) {
	type rec struct{ Sort int }
	var rows []rec

	// 不在 SELECT 里用占位，直接选出 sort 再在 Go 里计算 suffix，避免参数计数错乱
	q := tx.Table("base_category").
		Select("sort").
		Where(`
			team_id = ?
			AND is_deleted = 0
			AND sort > ? AND sort <= ?`,
			teamID, base, base+999).
		Order("sort ASC")
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Scan(&rows).Error; err != nil {
		return 0, fmt.Errorf("扫描 sort 失败: %w", err)
	}

	next := 1
	for _, r := range rows {
		suffix := r.Sort - base
		if suffix < next {
			continue
		}
		if suffix == next {
			next++
			continue
		}
		break
	}
	if next > 999 {
		return 0, fmt.Errorf("该 org 的 sort 段已满（1..999）")
	}
	return next, nil
}

// code 段的最小缺口：仅统计本 team、未删除、code 形如 <orgCode><三位数字>
func nextCodeSuffixByPrefix(tx *gorm.DB, teamID, orgCode string, forUpdate bool) (int, error) {
	type rec struct{ Suffix int }
	var rows []rec

	// 用 SUBSTRING(code, ? + 1) + REGEXP/LIKE；以 len(orgCode) 作为位置参数，避免 CHAR_LENGTH(?) 触发计数差异
	prefixLen := len(orgCode)

	q := tx.Table("base_category").
		Select("CAST(SUBSTRING(code, ? + 1) AS UNSIGNED) AS suffix", prefixLen).
		Where(`
			team_id = ?
			AND is_deleted = 0
			AND code LIKE CONCAT(?, '___')
			AND code REGEXP CONCAT('^', ?, '[0-9]{3}$')`,
			teamID, orgCode, orgCode).
		Order("suffix ASC")
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Scan(&rows).Error; err != nil {
		return 0, fmt.Errorf("扫描 code 后缀失败: %w", err)
	}

	next := 1
	for _, r := range rows {
		if r.Suffix < next {
			continue
		}
		if r.Suffix == next {
			next++
			continue
		}
		break
	}
	if next > 999 {
		return 0, fmt.Errorf("该 org 的 code 段已满（001..999）")
	}
	return next, nil
}
