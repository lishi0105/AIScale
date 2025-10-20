package goods

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	utils "hdzk.cn/foodapp/pkg/utils"
)

type Goods struct {
	ID          string    `gorm:"primaryKey;type:char(36)"`
	Name        string    `gorm:"size:128;not null;comment:商品名称"`
	Code        *string   `gorm:"size:64;not null;uniqueIndex:uq_goods_code;comment:SKU/条码"`
	Sort        int       `gorm:"not null;default:0;index;comment:排序码"`
	Pinyin      *string   `gorm:"size:128;comment:商品拼音（检索用）"`
	SpecID      string    `gorm:"column:spec_id;type:char(36);not null;comment:规格ID（base_spec.id）"`
	UnitID      string    `gorm:"column:unit_id;type:char(36);not null;comment:规格ID（base_unit.id）"`
	ImageURL    *string   `gorm:"column:image_url;size:512;comment:商品图片URL"`
	Description *string   `gorm:"column:description;size:512;comment:商品描述"`
	CategoryID  string    `gorm:"column:category_id;type:char(36);not null;comment:品类ID（base_category.id）"`
	OrgID       string    `gorm:"column:org_id;type:char(36);not null;comment:组织ID"`
	IsDeleted   int       `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效,1=删除"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (g *Goods) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.NewString()
	}
	if g.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}

	// 1) 查 organ 的 code/sort（FOR UPDATE）
	orgCode, orgSort, err := utils.GetOrgCodeAndSortByID(tx.Statement.Context, tx, g.OrgID, true)
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
	if g.Sort <= 0 {
		suf, err := utils.NextSortSuffix(tx, g.TableName(), g.OrgID, base, true)
		if err != nil {
			return err
		}
		g.Sort = base + suf
	}

	// 3) code = org.code + 三位后缀（各自独立找缺口）
	if g.Code == nil || (g.Code != nil && *g.Code == "") {
		suf, err := utils.NextCodeSuffixByPrefix(tx, g.TableName(), g.OrgID, orgCode, true)
		if err != nil {
			return err
		}
		auto := fmt.Sprintf("%s%03d", orgCode, suf)
		g.Code = &auto
	}

	// 4) 自动拼音
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
