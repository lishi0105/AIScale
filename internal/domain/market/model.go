package market

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	utils "hdzk.cn/foodapp/pkg/utils"
)

// BaseMarket 基础市场主数据
type BaseMarket struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	Name      string    `gorm:"size:64;not null;comment:市场名称"`
	OrgID     string    `gorm:"column:org_id;type:char(36);not null;comment:中队ID"`
	Code      *string   `gorm:"size:64;uniqueIndex:uq_market_code;comment:市场编码"`
	Sort      int       `gorm:"not null;default:0;comment:排序码"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效 1=已删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (BaseMarket) TableName() string { return "base_market" }

func (m *BaseMarket) BeforeCreate(tx *gorm.DB) error {
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

	// 3) code = org.code + 三位后缀
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

// BasePriceInquiry 询价单（表头）
type BasePriceInquiry struct {
	ID            string    `gorm:"primaryKey;type:char(36)"`
	OrgID         string    `gorm:"column:org_id;type:char(36);not null;comment:中队ID"`
	InquiryTitle  string    `gorm:"column:inquiry_title;size:64;not null;comment:询价单标题"`
	InquiryDate   time.Time `gorm:"column:inquiry_date;type:date;not null;comment:业务日期"`
	InquiryYear   *int16    `gorm:"column:inquiry_year;->;comment:询价年份（生成列）"`
	InquiryMonth  *int8     `gorm:"column:inquiry_month;->;comment:询价月份（生成列）"`
	InquiryTenDay *int8     `gorm:"column:inquiry_ten_day;->;comment:旬：1=上旬 2=中旬 3=下旬（生成列）"`
	IsDeleted     int       `gorm:"column:is_deleted;not null;default:0;comment:软删：0=有效 1=删除"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (BasePriceInquiry) TableName() string { return "base_price_inquiry" }

func (m *BasePriceInquiry) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}
	return nil
}

// PriceInquiryItem 询价商品明细
type PriceInquiryItem struct {
	ID                 string     `gorm:"primaryKey;type:char(36)"`
	InquiryID          string     `gorm:"column:inquiry_id;type:char(36);not null;comment:base_price_inquiry.id"`
	GoodsID            string     `gorm:"column:goods_id;type:char(36);not null;comment:base_goods.id"`
	CategoryID         string     `gorm:"column:category_id;type:char(36);not null;comment:base_category.id"`
	SpecID             *string    `gorm:"column:spec_id;type:char(36);comment:base_spec.id（快照）"`
	UnitID             *string    `gorm:"column:unit_id;type:char(36);comment:base_unit.id（快照）"`
	GoodsNameSnap      string     `gorm:"column:goods_name_snap;size:128;not null;comment:商品名称快照"`
	CategoryNameSnap   string     `gorm:"column:category_name_snap;size:64;not null;comment:品类名称快照"`
	SpecNameSnap       *string    `gorm:"column:spec_name_snap;size:32;comment:规格名称快照"`
	UnitNameSnap       *string    `gorm:"column:unit_name_snap;size:32;comment:单位名称快照"`
	GuidePrice         *float64   `gorm:"column:guide_price;type:decimal(12,2);comment:发改委指导价"`
	LastMonthAvgPrice  *float64   `gorm:"column:last_month_avg_price;type:decimal(12,2);comment:上月均价"`
	CurrentAvgPrice    *float64   `gorm:"column:current_avg_price;type:decimal(12,2);comment:本期均价"`
	Sort               int        `gorm:"not null;default:0;comment:排序码"`
	IsDeleted          int        `gorm:"column:is_deleted;not null;default:0;comment:软删：0=有效 1=删除"`
	CreatedAt          time.Time  `gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime"`
}

func (PriceInquiryItem) TableName() string { return "price_inquiry_item" }

func (m *PriceInquiryItem) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.InquiryID == "" {
		return errors.New("InquiryID(inquiry_id) 不能为空")
	}
	if m.GoodsID == "" {
		return errors.New("GoodsID(goods_id) 不能为空")
	}
	if m.CategoryID == "" {
		return errors.New("CategoryID(category_id) 不能为空")
	}
	return nil
}

// PriceMarketInquiry 市场报价
type PriceMarketInquiry struct {
	ID             string     `gorm:"primaryKey;type:char(36)"`
	InquiryID      string     `gorm:"column:inquiry_id;type:char(36);not null;comment:base_price_inquiry.id"`
	ItemID         string     `gorm:"column:item_id;type:char(36);not null;comment:price_inquiry_item.id"`
	MarketID       *string    `gorm:"column:market_id;type:char(36);comment:base_market.id"`
	MarketNameSnap string     `gorm:"column:market_name_snap;size:64;not null;comment:市场名称快照"`
	Price          *float64   `gorm:"type:decimal(12,2);comment:该市场的单价"`
	IsDeleted      int        `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效 1=已删除"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime"`
}

func (PriceMarketInquiry) TableName() string { return "price_market_inquiry" }

func (m *PriceMarketInquiry) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.InquiryID == "" {
		return errors.New("InquiryID(inquiry_id) 不能为空")
	}
	if m.ItemID == "" {
		return errors.New("ItemID(item_id) 不能为空")
	}
	return nil
}

// PriceSupplierSettlement 供应商结算
type PriceSupplierSettlement struct {
	ID                 string     `gorm:"primaryKey;type:char(36)"`
	InquiryID          string     `gorm:"column:inquiry_id;type:char(36);not null;comment:base_price_inquiry.id"`
	ItemID             string     `gorm:"column:item_id;type:char(36);not null;comment:price_inquiry_item.id"`
	SupplierID         *string    `gorm:"column:supplier_id;type:char(36);comment:supplier.id"`
	SupplierNameSnap   string     `gorm:"column:supplier_name_snap;size:128;not null;comment:供应商名称快照"`
	FloatRatioSnap     float64    `gorm:"column:float_ratio_snap;type:decimal(6,4);not null;comment:浮动比例快照"`
	SettlementPrice    *float64   `gorm:"column:settlement_price;type:decimal(12,2);comment:本期结算价"`
	IsDeleted          int        `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效 1=已删除"`
	CreatedAt          time.Time  `gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime"`
}

func (PriceSupplierSettlement) TableName() string { return "price_supplier_settlement" }

func (m *PriceSupplierSettlement) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.InquiryID == "" {
		return errors.New("InquiryID(inquiry_id) 不能为空")
	}
	if m.ItemID == "" {
		return errors.New("ItemID(item_id) 不能为空")
	}
	return nil
}
