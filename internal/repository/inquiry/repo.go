package inquiry

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/inquiry"
)

// ========== BaseMarket ==========

type MarketUpdateParams struct {
	ID   string
	Name *string
	Code *string
	Sort *int
}

type MarketRepository interface {
	CreateMarket(ctx context.Context, m *domain.BaseMarket) error
	GetMarket(ctx context.Context, id string) (*domain.BaseMarket, error)
	ListMarkets(ctx context.Context, keyword string, orgID string, page, pageSize int) ([]domain.BaseMarket, int64, error)
	UpdateMarket(ctx context.Context, params MarketUpdateParams) error
	SoftDeleteMarket(ctx context.Context, id string) error
	HardDeleteMarket(ctx context.Context, id string) error
}

// ========== BasePriceInquiry ==========

type InquiryUpdateParams struct {
	ID           string
	InquiryTitle *string
	InquiryDate  *string // YYYY-MM-DD format
}

type InquiryRepository interface {
	CreateInquiry(ctx context.Context, m *domain.BasePriceInquiry) error
	GetInquiry(ctx context.Context, id string) (*domain.BasePriceInquiry, error)
	ListInquiries(ctx context.Context, keyword string, orgID string, year, month, tenDay *int, page, pageSize int) ([]domain.BasePriceInquiry, int64, error)
	UpdateInquiry(ctx context.Context, params InquiryUpdateParams) error
	SoftDeleteInquiry(ctx context.Context, id string) error
	HardDeleteInquiry(ctx context.Context, id string) error
	DeleteInquiryWithCascade(ctx context.Context, id string) error
}

// ========== PriceInquiryItem ==========

type InquiryItemUpdateParams struct {
	ID                 string
	GoodsID            *string
	CategoryID         *string
	SpecID             *string
	UnitID             *string
	GoodsNameSnap      *string
	CategoryNameSnap   *string
	SpecNameSnap       *string
	UnitNameSnap       *string
	GuidePrice         *float64
	LastMonthAvgPrice  *float64
	CurrentAvgPrice    *float64
	Sort               *int
	UpdateSpecID       bool
	UpdateUnitID       bool
	UpdateSpecNameSnap bool
	UpdateUnitNameSnap bool
	UpdateGuidePrice   bool
	UpdateLastMonth    bool
	UpdateCurrentAvg   bool
}

type InquiryItemRepository interface {
	CreateInquiryItem(ctx context.Context, m *domain.PriceInquiryItem) error
	GetInquiryItem(ctx context.Context, id string) (*domain.PriceInquiryItem, error)
	ListInquiryItems(ctx context.Context, inquiryID string, categoryID *string, page, pageSize int) ([]domain.PriceInquiryItem, int64, error)
	UpdateInquiryItem(ctx context.Context, params InquiryItemUpdateParams) error
	SoftDeleteInquiryItem(ctx context.Context, id string) error
	HardDeleteInquiryItem(ctx context.Context, id string) error
}

// ========== PriceMarketInquiry ==========

type MarketInquiryUpdateParams struct {
	ID             string
	MarketID       *string
	MarketNameSnap *string
	Price          *float64
	UpdateMarketID bool
	UpdatePrice    bool
}

type MarketInquiryRepository interface {
	CreateMarketInquiry(ctx context.Context, m *domain.PriceMarketInquiry) error
	GetMarketInquiry(ctx context.Context, id string) (*domain.PriceMarketInquiry, error)
	ListMarketInquiries(ctx context.Context, inquiryID, itemID *string, page, pageSize int) ([]domain.PriceMarketInquiry, int64, error)
	UpdateMarketInquiry(ctx context.Context, params MarketInquiryUpdateParams) error
	SoftDeleteMarketInquiry(ctx context.Context, id string) error
	HardDeleteMarketInquiry(ctx context.Context, id string) error
}

// ========== PriceSupplierSettlement ==========

type SupplierSettlementUpdateParams struct {
	ID               string
	SupplierID       *string
	SupplierNameSnap *string
	FloatRatioSnap   *float64
	SettlementPrice  *float64
	UpdateSupplierID bool
	UpdateSettlement bool
}

type SupplierSettlementRepository interface {
	CreateSupplierSettlement(ctx context.Context, m *domain.PriceSupplierSettlement) error
	GetSupplierSettlement(ctx context.Context, id string) (*domain.PriceSupplierSettlement, error)
	ListSupplierSettlements(ctx context.Context, inquiryID, itemID *string, page, pageSize int) ([]domain.PriceSupplierSettlement, int64, error)
	UpdateSupplierSettlement(ctx context.Context, params SupplierSettlementUpdateParams) error
	SoftDeleteSupplierSettlement(ctx context.Context, id string) error
	HardDeleteSupplierSettlement(ctx context.Context, id string) error
}

// ========== Constructors ==========

func NewMarketRepository(db *gorm.DB) MarketRepository {
	return &marketRepo{db: db}
}

func NewInquiryRepository(db *gorm.DB) InquiryRepository {
	return &inquiryRepo{db: db}
}

func NewInquiryItemRepository(db *gorm.DB) InquiryItemRepository {
	return &inquiryItemRepo{db: db}
}

func NewMarketInquiryRepository(db *gorm.DB) MarketInquiryRepository {
	return &marketInquiryRepo{db: db}
}

func NewSupplierSettlementRepository(db *gorm.DB) SupplierSettlementRepository {
	return &supplierSettlementRepo{db: db}
}
