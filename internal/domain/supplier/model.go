package supplier

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	utils "hdzk.cn/foodapp/pkg/utils"
)

type Supplier struct {
	ID          string     `gorm:"primaryKey;type:char(36)"`
	Name        string     `gorm:"size:128;not null;uniqueIndex:uq_supplier_name;comment:供货商名称"`
	Code        *string    `gorm:"size:64;uniqueIndex:uq_supplier_code;comment:供货商编码（可选，建议唯一）"`
	Sort        int        `gorm:"not null;default:0;index;comment:排序值"`
	Pinyin      *string    `gorm:"size:64;comment:拼音（可选，用于搜索）"`
	Status      int        `gorm:"type:tinyint;not null;default:1;comment:状态：1=正常,2=禁用"`
	Description string     `gorm:"type:text;not null;comment:供应商描述"`
	FloatRatio  float64    `gorm:"type:decimal(6,4);not null;default:1.0000;comment:浮动比例：结算价=合同价*float_ratio"`
	OrgID       *string    `gorm:"column:org_id;type:char(36);comment:中队ID"`
	StartTime   *time.Time `gorm:"comment:开始时间"`
	EndTime     *time.Time `gorm:"comment:结束时间"`
	IsDeleted   int        `gorm:"not null;default:0;comment:软删标记：0=有效,1=已删除"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}

	// 如果提供了 org_id，则按照与 category 相同的规则生成 code 和 sort
	if s.OrgID != nil && *s.OrgID != "" {
		// 1) 查 organ 的 code/sort（FOR UPDATE）
		orgCode, orgSort, err := utils.GetOrgCodeAndSortByID(tx.Statement.Context, tx, *s.OrgID, true)
		if err != nil {
			return fmt.Errorf("查询 org code/sort 失败: %w", err)
		}
		if orgCode == "" {
			return errors.New("org.code 为空，无法派生 supplier.code")
		}
		if orgSort < 0 {
			return fmt.Errorf("org.sort 非法: %d", orgSort)
		}
		base := orgSort * 1000

		// 2) sort = org.sort*1000 + 最小缺口
		if s.Sort <= 0 {
			suf, err := nextSortSuffix(tx, *s.OrgID, base, true)
			if err != nil {
				return err
			}
			s.Sort = base + suf
		}

		// 3) code = org.code + 三位后缀（各自独立找缺口）
		if s.Code == nil || (s.Code != nil && *s.Code == "") {
			suf, err := nextCodeSuffixByPrefix(tx, *s.OrgID, orgCode, true)
			if err != nil {
				return err
			}
			auto := fmt.Sprintf("%s%03d", orgCode, suf)
			s.Code = &auto
		}
	}

	// 4) 自动拼音
	if s.Pinyin == nil || (s.Pinyin != nil && *s.Pinyin == "") {
		if utils.ContainsChinese(s.Name) {
			p := utils.GeneratePinyin(s.Name)
			s.Pinyin = &p
		}
	}

	// 5) 验证 float_ratio > 0
	if s.FloatRatio <= 0 {
		return errors.New("float_ratio 必须大于 0")
	}

	// 6) 验证时间范围：start_time IS NULL OR end_time IS NULL OR start_time <= end_time
	if s.StartTime != nil && s.EndTime != nil {
		if s.StartTime.After(*s.EndTime) {
			return errors.New("start_time 必须小于等于 end_time")
		}
	}

	return nil
}

func (Supplier) TableName() string { return "supplier" }

// ---------- helpers ----------

// sort 段的最小缺口：仅统计本 organ (base, base+999] 的 sort
func nextSortSuffix(tx *gorm.DB, orgID string, base int, forUpdate bool) (int, error) {
	type rec struct{ Sort int }
	var rows []rec

	q := tx.Table("supplier").
		Select("sort").
		Where(`
			org_id = ?
			AND is_deleted = 0
			AND sort > ? AND sort <= ?`,
			orgID, base, base+999).
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

// code 段的最小缺口：仅统计本 organ 形如 <orgCode><三位数字>
func nextCodeSuffixByPrefix(tx *gorm.DB, orgID, orgCode string, forUpdate bool) (int, error) {
	type rec struct{ Suffix int }
	var rows []rec

	prefixLen := len(orgCode)

	q := tx.Table("supplier").
		Select("CAST(SUBSTRING(code, ? + 1) AS UNSIGNED) AS suffix", prefixLen).
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
