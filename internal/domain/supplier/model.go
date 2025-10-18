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
	ID             string     `gorm:"primaryKey;type:char(36)"`
	Name           string     `gorm:"size:128;not null;comment:供货商名称"`
	Code           *string    `gorm:"size:64;comment:供货商编码（可选）"`
	Sort           int        `gorm:"not null;default:0;index;comment:排序值"`
	ContactName    *string    `gorm:"type:varchar(64);comment:联系人姓名"`
	ContactPhone   *string    `gorm:"type:varchar(32);comment:联系电话"`
	ContactEmail   *string    `gorm:"type:varchar(128);comment:联系邮箱"`
	ContactAddress *string    `gorm:"type:varchar(255);comment:联系地址"`
	Pinyin         *string    `gorm:"size:64;comment:拼音（可选，用于搜索）"`
	Status         int        `gorm:"not null;default:1;comment:状态：1=正常,2=禁用"`
	Description    string     `gorm:"type:text;not null;comment:供应商描述"`
	FloatRatio     float64    `gorm:"type:decimal(6,4);not null;default:1.0000;comment:浮动比例"`
	OrgID          string     `gorm:"column:org_id;type:char(36);not null;comment:所属机构ID"`
	StartTime      *time.Time `gorm:"column:start_time"`
	EndTime        *time.Time `gorm:"column:end_time"`
	IsDeleted      int        `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效,1=已删除"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime"`
}

func (s *Supplier) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}
	if s.Status == 0 {
		s.Status = 1
	}

	orgCode, orgSort, err := utils.GetOrgCodeAndSortByID(tx.Statement.Context, tx, s.OrgID, true)
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

	if s.Sort <= 0 {
		suf, err := nextSupplierSortSuffix(tx, s.OrgID, base, true)
		if err != nil {
			return err
		}
		s.Sort = base + suf
	}

	if s.Code == nil || (s.Code != nil && *s.Code == "") {
		suf, err := nextSupplierCodeSuffix(tx, s.OrgID, orgCode, true)
		if err != nil {
			return err
		}
		auto := fmt.Sprintf("%s%03d", orgCode, suf)
		s.Code = &auto
	}

	if s.Pinyin == nil || (s.Pinyin != nil && *s.Pinyin == "") {
		if utils.ContainsChinese(s.Name) {
			p := utils.GeneratePinyin(s.Name)
			s.Pinyin = &p
		}
	}
	return nil
}

func (Supplier) TableName() string { return "supplier" }

func nextSupplierSortSuffix(tx *gorm.DB, orgID string, base int, forUpdate bool) (int, error) {
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

func nextSupplierCodeSuffix(tx *gorm.DB, orgID, orgCode string, forUpdate bool) (int, error) {
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
