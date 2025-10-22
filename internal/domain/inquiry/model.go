package inquiry

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	goodsDomain "hdzk.cn/foodapp/internal/domain/goods"
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
	ID           string    `gorm:"primaryKey;type:char(36)"`
	OrgID        string    `gorm:"column:org_id;type:char(36);not null;comment:中队ID"`
	InquiryTitle string    `gorm:"column:inquiry_title;size:64;not null;comment:询价单标题"`
	InquiryDate  time.Time `gorm:"column:inquiry_date;type:date;not null;comment:业务日期"`

	// 改成 <-:create，允许创建时写入
	InquiryYear   *int16 `gorm:"column:inquiry_year;<-:create;comment:询价年份（应用计算）"`
	InquiryMonth  *int8  `gorm:"column:inquiry_month;<-:create;comment:询价月份（应用计算）"`
	InquiryTenDay *int8  `gorm:"column:inquiry_ten_day;<-:create;comment:旬：1=上旬 2=中旬 3=下旬（应用计算）"`

	IsDeleted int       `gorm:"column:is_deleted;not null;default:0;comment:软删：0=有效 1=删除"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (BasePriceInquiry) TableName() string { return "base_price_inquiry" }

func (m *BasePriceInquiry) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.OrgID == "" {
		return errors.New("OrgID(org_id) 不能为空")
	}
	// 如果 InquiryDate 为空，则用今天
	if m.InquiryDate.IsZero() {
		m.InquiryDate = time.Now()
	}

	// 依据 InquiryDate 计算 年/月/旬
	y, mm, d := m.InquiryDate.Date()
	yy := int16(y)
	mm8 := int8(mm)

	var td int8
	switch {
	case d <= 10:
		td = 1 // 上旬
	case d <= 20:
		td = 2 // 中旬
	default:
		td = 3 // 下旬
	}

	m.InquiryYear = &yy
	m.InquiryMonth = &mm8
	m.InquiryTenDay = &td

	return nil
}

// PriceInquiryItem 询价商品明细
type PriceInquiryItem struct {
	ID                string    `gorm:"primaryKey;type:char(36)"`
	InquiryID         string    `gorm:"column:inquiry_id;type:char(36);not null;comment:base_price_inquiry.id"`
	GoodsID           string    `gorm:"column:goods_id;type:char(36);not null;comment:base_goods.id"`
	CategoryID        string    `gorm:"column:category_id;type:char(36);not null;comment:base_category.id"`
	SpecID            *string   `gorm:"column:spec_id;type:char(36);comment:base_spec.id（快照）"`
	UnitID            *string   `gorm:"column:unit_id;type:char(36);comment:base_unit.id（快照）"`
	GoodsNameSnap     string    `gorm:"column:goods_name_snap;size:128;not null;comment:商品名称快照"`
	CategoryNameSnap  string    `gorm:"column:category_name_snap;size:64;not null;comment:品类名称快照"`
	SpecNameSnap      *string   `gorm:"column:spec_name_snap;size:32;comment:规格名称快照"`
	UnitNameSnap      *string   `gorm:"column:unit_name_snap;size:32;comment:单位名称快照"`
	GuidePrice        *float64  `gorm:"column:guide_price;type:decimal(12,2);comment:发改委指导价"`
	LastMonthAvgPrice *float64  `gorm:"column:last_month_avg_price;type:decimal(12,2);comment:上月均价"`
	CurrentAvgPrice   *float64  `gorm:"column:current_avg_price;type:decimal(12,2);comment:本期均价"`
	Sort              int       `gorm:"not null;default:0;comment:排序码"`
	IsDeleted         int       `gorm:"column:is_deleted;not null;default:0;comment:软删：0=有效 1=删除"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`

	// 临时字段，用于联合查询时存储商品拼音
	Pinyin *string `gorm:"column:pinyin;size:128;comment:商品拼音（联合查询用）"`

	// 关联对象，用于前端显示
	Goods           *goodsDomain.Goods   `gorm:"-" json:"goods,omitempty"`
	MarketInquiries []PriceMarketInquiry `gorm:"-" json:"market_inquiries,omitempty"`
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
	ID             string    `gorm:"primaryKey;type:char(36)"`
	GoodsID        string    `gorm:"column:goods_id;type:char(36);not null;comment:base_goods.id"`
	InquiryID      string    `gorm:"column:inquiry_id;type:char(36);not null;comment:base_price_inquiry.id"`
	ItemID         string    `gorm:"column:item_id;type:char(36);not null;comment:price_inquiry_item.id"`
	MarketID       *string   `gorm:"column:market_id;type:char(36);comment:base_market.id"`
	MarketNameSnap string    `gorm:"column:market_name_snap;size:64;not null;comment:市场名称快照"`
	InquiryDate    time.Time `gorm:"column:inquiry_date;type:date;not null;comment:询价日期"`
	Price          *float64  `gorm:"type:decimal(12,2);comment:该市场的单价"`
	IsDeleted      int       `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效 1=已删除"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
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
	ID               string    `gorm:"primaryKey;type:char(36)"`
	InquiryID        string    `gorm:"column:inquiry_id;type:char(36);not null;comment:base_price_inquiry.id"`
	ItemID           string    `gorm:"column:item_id;type:char(36);not null;comment:price_inquiry_item.id"`
	SupplierID       *string   `gorm:"column:supplier_id;type:char(36);comment:supplier.id"`
	SupplierNameSnap string    `gorm:"column:supplier_name_snap;size:128;not null;comment:供应商名称快照"`
	FloatRatioSnap   float64   `gorm:"column:float_ratio_snap;type:decimal(6,4);not null;comment:浮动比例快照"`
	SettlementPrice  *float64  `gorm:"column:settlement_price;type:decimal(12,2);comment:本期结算价"`
	IsDeleted        int       `gorm:"column:is_deleted;not null;default:0;comment:软删标记：0=有效 1=已删除"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
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
