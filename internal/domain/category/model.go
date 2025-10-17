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
	TeamID    string    `gorm:"type:char(36);not null;comment:所属机构ID"` // team_id
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

	// 1) 查 org 的 code/sort（加锁；GORM 单条写入默认在事务里）
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

	// 2) 生成 sort（org.sort*1000 + 最小缺口）
	if c.Sort <= 0 {
		base := orgSort * 1000
		suf, err := nextSortSuffix(tx, c.TeamID, base, true)
		if err != nil {
			return err
		}
		c.Sort = base + suf
	}

	// 3) 生成 code（org.code + 三位后缀的最小缺口）
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

// ---- helpers ----

// sort 段的最小缺口：仅统计本 org、未删除、落在 (base, base+999] 的 sort
func nextSortSuffix(tx *gorm.DB, orgID string, base int, forUpdate bool) (int, error) {
	type rec struct{ Suffix int }
	var rows []rec

	q := tx.Table("base_category").
		Select("(sort - ?) AS suffix", base).
		Where(`
			org_id = ?
			AND is_deleted = 0
			AND sort > ? AND sort <= ?`,
			orgID, base, base, base+999).
		Order("suffix ASC")
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Scan(&rows).Error; err != nil {
		return 0, fmt.Errorf("扫描 sort 后缀失败: %w", err)
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
		return 0, fmt.Errorf("该 org 的 sort 段已满（1..999）")
	}
	return next, nil
}

// code 段的最小缺口：仅统计本 org、未删除、code 形如 <orgCode><三位数字>
// 用 SUBSTRING + CAST 抽出后三位数字并排序找缺口
func nextCodeSuffixByPrefix(tx *gorm.DB, orgID, orgCode string, forUpdate bool) (int, error) {
	type rec struct{ Suffix int }
	var rows []rec

	q := tx.Table("base_category").
		// 从 code 的 (len(orgCode)+1) 位置提取后三位并转为数字
		Select("CAST(SUBSTRING(code, CHAR_LENGTH(?) + 1) AS UNSIGNED) AS suffix", orgCode).
		Where(`
			org_id = ?
			AND is_deleted = 0
			AND code LIKE CONCAT(?, '___')
			AND code REGEXP CONCAT('^', ?, '[0-9]{3}$')`,
			orgID, orgCode, orgCode).
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
