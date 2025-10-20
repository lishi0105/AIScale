package goods

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	utils "hdzk.cn/foodapp/pkg/utils"
)

type Goods struct {
	ID                 string    `gorm:"primaryKey;type:char(36)"`
	Name               string    `gorm:"size:128;not null;comment:商品名称"`
	Code               *string   `gorm:"size:64;not null;uniqueIndex:uq_goods_code;comment:SKU/条码"`
	Sort               int       `gorm:"not null;default:0;index;comment:排序码"`
	Pinyin             *string   `gorm:"size:128;comment:商品拼音（检索用）"`
	SpecID             string    `gorm:"column:spec_id;type:char(36);not null;comment:规格ID（base_spec.id）"`
	ImageURL           *string   `gorm:"column:image_url;size:512;comment:商品图片URL"`
	AcceptanceStandard *string   `gorm:"column:acceptance_standard;size:512;comment:验收标准"`
	CategoryID         string    `gorm:"column:category_id;type:char(36);not null;comment:品类ID（base_category.id）"`
	OrgID              string    `gorm:"column:org_id;type:char(36);not null;comment:组织ID"`
	IsDeleted          int       `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效,1=删除"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

func (g *Goods) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.NewString()
	}
	if g.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}

	// 对 sort 做基于机构的自动补位，保持和品类/供应商一致
	orgCode, orgSort, err := utils.GetOrgCodeAndSortByID(tx.Statement.Context, tx, g.OrgID, true)
	if err != nil {
		return fmt.Errorf("查询 org sort 失败: %w", err)
	}
	if orgSort < 0 {
		return fmt.Errorf("org.sort 非法: %d", orgSort)
	}
	base := orgSort * 1000

	if g.Sort <= 0 {
		suf, err := g.nextGoodsSortSuffix(tx, g.OrgID, base, true)
		if err != nil {
			return err
		}
		g.Sort = base + suf
	}
	if g.Code == nil || (g.Code != nil && *g.Code == "") {
		suf, err := g.nextGoodsCodeSuffix(tx, g.OrgID, orgCode, true)
		if err != nil {
			return err
		}
		auto := fmt.Sprintf("%s%03d", orgCode, suf)
		g.Code = &auto
	}
	if g.Pinyin == nil || (g.Pinyin != nil && *g.Pinyin == "") {
		if utils.ContainsChinese(g.Name) {
			p := utils.GeneratePinyin(g.Name)
			g.Pinyin = &p
		}
	}
	return nil
}

func (Goods) TableName() string { return "base_goods" }
func codeFromSort(sort int) string {
	return fmt.Sprintf("%02d", sort)
}

func (g *Goods) nextGoodsSortSuffix(tx *gorm.DB, orgID string, base int, forUpdate bool) (int, error) {
	type rec struct{ Sort int }
	var rows []rec
	q := tx.Table(g.TableName()).
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

func (g *Goods) nextGoodsCodeSuffix(tx *gorm.DB, orgID, orgCode string, forUpdate bool) (int, error) {
	type rec struct{ Suffix int }
	var rows []rec

	prefixLen := len(orgCode)

	q := tx.Table(g.TableName()).
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
