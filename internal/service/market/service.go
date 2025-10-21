package market

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/market"
	repo "hdzk.cn/foodapp/internal/repository/market"
)

// ========== BaseMarket Service ==========

type MarketService struct {
	r repo.MarketRepository
}

func NewMarketService(r repo.MarketRepository) *MarketService {
	return &MarketService{r: r}
}

type MarketCreateParams struct {
	Name  string
	OrgID string
	Code  *string
	Sort  *int
}

type MarketUpdateParams struct {
	ID   string
	Name *string
	Code *string
	Sort *int
}

func (s *MarketService) CreateMarket(ctx context.Context, params MarketCreateParams) (*domain.BaseMarket, error) {
	name, err := normalizeRequiredValue(params.Name, "name")
	if err != nil {
		return nil, err
	}
	orgID, err := normalizeRequiredValue(params.OrgID, "org_id")
	if err != nil {
		return nil, err
	}

	normalizedCode, _ := normalizeOptional(params.Code)

	m := &domain.BaseMarket{
		ID:    uuid.NewString(),
		Name:  name,
		OrgID: orgID,
		Code:  normalizedCode,
	}
	if params.Sort != nil {
		m.Sort = *params.Sort
	}
	return m, s.r.CreateMarket(ctx, m)
}

func (s *MarketService) GetMarket(ctx context.Context, id string) (*domain.BaseMarket, error) {
	return s.r.GetMarket(ctx, strings.TrimSpace(id))
}

func (s *MarketService) ListMarkets(ctx context.Context, keyword string, orgID string, page, pageSize int) ([]domain.BaseMarket, int64, error) {
	trimmedOrg := strings.TrimSpace(orgID)
	if trimmedOrg == "" {
		return nil, 0, fmt.Errorf("org_id 不能为空")
	}
	kw := strings.TrimSpace(keyword)
	return s.r.ListMarkets(ctx, kw, trimmedOrg, page, pageSize)
}

func (s *MarketService) UpdateMarket(ctx context.Context, params MarketUpdateParams) error {
	normalizedName, err := normalizeOptionalRequired(params.Name, "name")
	if err != nil {
		return err
	}
	normalizedCode, err := normalizeOptionalRequired(params.Code, "code")
	if err != nil {
		return err
	}

	repoParams := repo.MarketUpdateParams{
		ID:   strings.TrimSpace(params.ID),
		Name: normalizedName,
		Code: normalizedCode,
		Sort: params.Sort,
	}
	return s.r.UpdateMarket(ctx, repoParams)
}

func (s *MarketService) SoftDeleteMarket(ctx context.Context, id string) error {
	return s.r.SoftDeleteMarket(ctx, strings.TrimSpace(id))
}

func (s *MarketService) HardDeleteMarket(ctx context.Context, id string) error {
	return s.r.HardDeleteMarket(ctx, strings.TrimSpace(id))
}

// ========== BasePriceInquiry Service ==========

type InquiryService struct {
	r repo.InquiryRepository
}

func NewInquiryService(r repo.InquiryRepository) *InquiryService {
	return &InquiryService{r: r}
}

type InquiryCreateParams struct {
	OrgID        string
	InquiryTitle string
	InquiryDate  string // YYYY-MM-DD format
}

type InquiryUpdateParams struct {
	ID           string
	InquiryTitle *string
	InquiryDate  *string // YYYY-MM-DD format
}

func (s *InquiryService) CreateInquiry(ctx context.Context, params InquiryCreateParams) (*domain.BasePriceInquiry, error) {
	orgID, err := normalizeRequiredValue(params.OrgID, "org_id")
	if err != nil {
		return nil, err
	}
	title, err := normalizeRequiredValue(params.InquiryTitle, "inquiry_title")
	if err != nil {
		return nil, err
	}
	dateStr, err := normalizeRequiredValue(params.InquiryDate, "inquiry_date")
	if err != nil {
		return nil, err
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("日期格式错误，应为 YYYY-MM-DD: %w", err)
	}

	m := &domain.BasePriceInquiry{
		ID:           uuid.NewString(),
		OrgID:        orgID,
		InquiryTitle: title,
		InquiryDate:  date,
	}
	return m, s.r.CreateInquiry(ctx, m)
}

func (s *InquiryService) GetInquiry(ctx context.Context, id string) (*domain.BasePriceInquiry, error) {
	return s.r.GetInquiry(ctx, strings.TrimSpace(id))
}

func (s *InquiryService) ListInquiries(ctx context.Context, keyword string, orgID string, year, month, tenDay *int, page, pageSize int) ([]domain.BasePriceInquiry, int64, error) {
	trimmedOrg := strings.TrimSpace(orgID)
	if trimmedOrg == "" {
		return nil, 0, fmt.Errorf("org_id 不能为空")
	}
	kw := strings.TrimSpace(keyword)
	return s.r.ListInquiries(ctx, kw, trimmedOrg, year, month, tenDay, page, pageSize)
}

func (s *InquiryService) UpdateInquiry(ctx context.Context, params InquiryUpdateParams) error {
	normalizedTitle, err := normalizeOptionalRequired(params.InquiryTitle, "inquiry_title")
	if err != nil {
		return err
	}
	
	var normalizedDate *string
	if params.InquiryDate != nil {
		dateStr := strings.TrimSpace(*params.InquiryDate)
		if dateStr == "" {
			return fmt.Errorf("inquiry_date 不能为空")
		}
		// Validate date format
		_, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		normalizedDate = &dateStr
	}

	repoParams := repo.InquiryUpdateParams{
		ID:           strings.TrimSpace(params.ID),
		InquiryTitle: normalizedTitle,
		InquiryDate:  normalizedDate,
	}
	return s.r.UpdateInquiry(ctx, repoParams)
}

func (s *InquiryService) SoftDeleteInquiry(ctx context.Context, id string) error {
	return s.r.SoftDeleteInquiry(ctx, strings.TrimSpace(id))
}

func (s *InquiryService) HardDeleteInquiry(ctx context.Context, id string) error {
	return s.r.HardDeleteInquiry(ctx, strings.TrimSpace(id))
}

// ========== PriceInquiryItem Service ==========

type InquiryItemService struct {
	r repo.InquiryItemRepository
}

func NewInquiryItemService(r repo.InquiryItemRepository) *InquiryItemService {
	return &InquiryItemService{r: r}
}

type InquiryItemCreateParams struct {
	InquiryID          string
	GoodsID            string
	CategoryID         string
	SpecID             *string
	UnitID             *string
	GoodsNameSnap      string
	CategoryNameSnap   string
	SpecNameSnap       *string
	UnitNameSnap       *string
	GuidePrice         *float64
	LastMonthAvgPrice  *float64
	CurrentAvgPrice    *float64
	Sort               *int
}

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
}

func (s *InquiryItemService) CreateInquiryItem(ctx context.Context, params InquiryItemCreateParams) (*domain.PriceInquiryItem, error) {
	inquiryID, err := normalizeRequiredValue(params.InquiryID, "inquiry_id")
	if err != nil {
		return nil, err
	}
	goodsID, err := normalizeRequiredValue(params.GoodsID, "goods_id")
	if err != nil {
		return nil, err
	}
	categoryID, err := normalizeRequiredValue(params.CategoryID, "category_id")
	if err != nil {
		return nil, err
	}
	goodsNameSnap, err := normalizeRequiredValue(params.GoodsNameSnap, "goods_name_snap")
	if err != nil {
		return nil, err
	}
	categoryNameSnap, err := normalizeRequiredValue(params.CategoryNameSnap, "category_name_snap")
	if err != nil {
		return nil, err
	}

	normalizedSpecID, _ := normalizeOptional(params.SpecID)
	normalizedUnitID, _ := normalizeOptional(params.UnitID)
	normalizedSpecNameSnap, _ := normalizeOptional(params.SpecNameSnap)
	normalizedUnitNameSnap, _ := normalizeOptional(params.UnitNameSnap)

	m := &domain.PriceInquiryItem{
		ID:                 uuid.NewString(),
		InquiryID:          inquiryID,
		GoodsID:            goodsID,
		CategoryID:         categoryID,
		SpecID:             normalizedSpecID,
		UnitID:             normalizedUnitID,
		GoodsNameSnap:      goodsNameSnap,
		CategoryNameSnap:   categoryNameSnap,
		SpecNameSnap:       normalizedSpecNameSnap,
		UnitNameSnap:       normalizedUnitNameSnap,
		GuidePrice:         params.GuidePrice,
		LastMonthAvgPrice:  params.LastMonthAvgPrice,
		CurrentAvgPrice:    params.CurrentAvgPrice,
	}
	if params.Sort != nil {
		m.Sort = *params.Sort
	}
	return m, s.r.CreateInquiryItem(ctx, m)
}

func (s *InquiryItemService) GetInquiryItem(ctx context.Context, id string) (*domain.PriceInquiryItem, error) {
	return s.r.GetInquiryItem(ctx, strings.TrimSpace(id))
}

func (s *InquiryItemService) ListInquiryItems(ctx context.Context, inquiryID string, categoryID *string, page, pageSize int) ([]domain.PriceInquiryItem, int64, error) {
	trimmedInquiry := strings.TrimSpace(inquiryID)
	if trimmedInquiry == "" {
		return nil, 0, fmt.Errorf("inquiry_id 不能为空")
	}

	var categoryPtr *string
	if categoryID != nil {
		normalized, err := normalizeOptionalWithOriginal(categoryID)
		if err != nil {
			return nil, 0, err
		}
		categoryPtr = normalized
	}

	return s.r.ListInquiryItems(ctx, trimmedInquiry, categoryPtr, page, pageSize)
}

func (s *InquiryItemService) UpdateInquiryItem(ctx context.Context, params InquiryItemUpdateParams) error {
	normalizedGoodsID, err := normalizeOptionalRequired(params.GoodsID, "goods_id")
	if err != nil {
		return err
	}
	normalizedCategoryID, err := normalizeOptionalRequired(params.CategoryID, "category_id")
	if err != nil {
		return err
	}
	normalizedGoodsNameSnap, err := normalizeOptionalRequired(params.GoodsNameSnap, "goods_name_snap")
	if err != nil {
		return err
	}
	normalizedCategoryNameSnap, err := normalizeOptionalRequired(params.CategoryNameSnap, "category_name_snap")
	if err != nil {
		return err
	}

	normalizedSpecID, updateSpecID := normalizeOptional(params.SpecID)
	normalizedUnitID, updateUnitID := normalizeOptional(params.UnitID)
	normalizedSpecNameSnap, updateSpecNameSnap := normalizeOptional(params.SpecNameSnap)
	normalizedUnitNameSnap, updateUnitNameSnap := normalizeOptional(params.UnitNameSnap)

	repoParams := repo.InquiryItemUpdateParams{
		ID:                 strings.TrimSpace(params.ID),
		GoodsID:            normalizedGoodsID,
		CategoryID:         normalizedCategoryID,
		SpecID:             normalizedSpecID,
		UnitID:             normalizedUnitID,
		GoodsNameSnap:      normalizedGoodsNameSnap,
		CategoryNameSnap:   normalizedCategoryNameSnap,
		SpecNameSnap:       normalizedSpecNameSnap,
		UnitNameSnap:       normalizedUnitNameSnap,
		GuidePrice:         params.GuidePrice,
		LastMonthAvgPrice:  params.LastMonthAvgPrice,
		CurrentAvgPrice:    params.CurrentAvgPrice,
		Sort:               params.Sort,
		UpdateSpecID:       updateSpecID,
		UpdateUnitID:       updateUnitID,
		UpdateSpecNameSnap: updateSpecNameSnap,
		UpdateUnitNameSnap: updateUnitNameSnap,
		UpdateGuidePrice:   params.GuidePrice != nil,
		UpdateLastMonth:    params.LastMonthAvgPrice != nil,
		UpdateCurrentAvg:   params.CurrentAvgPrice != nil,
	}
	return s.r.UpdateInquiryItem(ctx, repoParams)
}

func (s *InquiryItemService) SoftDeleteInquiryItem(ctx context.Context, id string) error {
	return s.r.SoftDeleteInquiryItem(ctx, strings.TrimSpace(id))
}

func (s *InquiryItemService) HardDeleteInquiryItem(ctx context.Context, id string) error {
	return s.r.HardDeleteInquiryItem(ctx, strings.TrimSpace(id))
}

// ========== PriceMarketInquiry Service ==========

type MarketInquiryService struct {
	r repo.MarketInquiryRepository
}

func NewMarketInquiryService(r repo.MarketInquiryRepository) *MarketInquiryService {
	return &MarketInquiryService{r: r}
}

type MarketInquiryCreateParams struct {
	InquiryID      string
	ItemID         string
	MarketID       *string
	MarketNameSnap string
	Price          *float64
}

type MarketInquiryUpdateParams struct {
	ID             string
	MarketID       *string
	MarketNameSnap *string
	Price          *float64
}

func (s *MarketInquiryService) CreateMarketInquiry(ctx context.Context, params MarketInquiryCreateParams) (*domain.PriceMarketInquiry, error) {
	inquiryID, err := normalizeRequiredValue(params.InquiryID, "inquiry_id")
	if err != nil {
		return nil, err
	}
	itemID, err := normalizeRequiredValue(params.ItemID, "item_id")
	if err != nil {
		return nil, err
	}
	marketNameSnap, err := normalizeRequiredValue(params.MarketNameSnap, "market_name_snap")
	if err != nil {
		return nil, err
	}

	normalizedMarketID, _ := normalizeOptional(params.MarketID)

	m := &domain.PriceMarketInquiry{
		ID:             uuid.NewString(),
		InquiryID:      inquiryID,
		ItemID:         itemID,
		MarketID:       normalizedMarketID,
		MarketNameSnap: marketNameSnap,
		Price:          params.Price,
	}
	return m, s.r.CreateMarketInquiry(ctx, m)
}

func (s *MarketInquiryService) GetMarketInquiry(ctx context.Context, id string) (*domain.PriceMarketInquiry, error) {
	return s.r.GetMarketInquiry(ctx, strings.TrimSpace(id))
}

func (s *MarketInquiryService) ListMarketInquiries(ctx context.Context, inquiryID, itemID *string, page, pageSize int) ([]domain.PriceMarketInquiry, int64, error) {
	var inquiryPtr *string
	if inquiryID != nil {
		normalized, err := normalizeOptionalWithOriginal(inquiryID)
		if err != nil {
			return nil, 0, err
		}
		inquiryPtr = normalized
	}

	var itemPtr *string
	if itemID != nil {
		normalized, err := normalizeOptionalWithOriginal(itemID)
		if err != nil {
			return nil, 0, err
		}
		itemPtr = normalized
	}

	return s.r.ListMarketInquiries(ctx, inquiryPtr, itemPtr, page, pageSize)
}

func (s *MarketInquiryService) UpdateMarketInquiry(ctx context.Context, params MarketInquiryUpdateParams) error {
	normalizedMarketNameSnap, err := normalizeOptionalRequired(params.MarketNameSnap, "market_name_snap")
	if err != nil {
		return err
	}

	normalizedMarketID, updateMarketID := normalizeOptional(params.MarketID)

	repoParams := repo.MarketInquiryUpdateParams{
		ID:             strings.TrimSpace(params.ID),
		MarketID:       normalizedMarketID,
		MarketNameSnap: normalizedMarketNameSnap,
		Price:          params.Price,
		UpdateMarketID: updateMarketID,
		UpdatePrice:    params.Price != nil,
	}
	return s.r.UpdateMarketInquiry(ctx, repoParams)
}

func (s *MarketInquiryService) SoftDeleteMarketInquiry(ctx context.Context, id string) error {
	return s.r.SoftDeleteMarketInquiry(ctx, strings.TrimSpace(id))
}

func (s *MarketInquiryService) HardDeleteMarketInquiry(ctx context.Context, id string) error {
	return s.r.HardDeleteMarketInquiry(ctx, strings.TrimSpace(id))
}

// ========== PriceSupplierSettlement Service ==========

type SupplierSettlementService struct {
	r repo.SupplierSettlementRepository
}

func NewSupplierSettlementService(r repo.SupplierSettlementRepository) *SupplierSettlementService {
	return &SupplierSettlementService{r: r}
}

type SupplierSettlementCreateParams struct {
	InquiryID        string
	ItemID           string
	SupplierID       *string
	SupplierNameSnap string
	FloatRatioSnap   float64
	SettlementPrice  *float64
}

type SupplierSettlementUpdateParams struct {
	ID               string
	SupplierID       *string
	SupplierNameSnap *string
	FloatRatioSnap   *float64
	SettlementPrice  *float64
}

func (s *SupplierSettlementService) CreateSupplierSettlement(ctx context.Context, params SupplierSettlementCreateParams) (*domain.PriceSupplierSettlement, error) {
	inquiryID, err := normalizeRequiredValue(params.InquiryID, "inquiry_id")
	if err != nil {
		return nil, err
	}
	itemID, err := normalizeRequiredValue(params.ItemID, "item_id")
	if err != nil {
		return nil, err
	}
	supplierNameSnap, err := normalizeRequiredValue(params.SupplierNameSnap, "supplier_name_snap")
	if err != nil {
		return nil, err
	}

	normalizedSupplierID, _ := normalizeOptional(params.SupplierID)

	m := &domain.PriceSupplierSettlement{
		ID:               uuid.NewString(),
		InquiryID:        inquiryID,
		ItemID:           itemID,
		SupplierID:       normalizedSupplierID,
		SupplierNameSnap: supplierNameSnap,
		FloatRatioSnap:   params.FloatRatioSnap,
		SettlementPrice:  params.SettlementPrice,
	}
	return m, s.r.CreateSupplierSettlement(ctx, m)
}

func (s *SupplierSettlementService) GetSupplierSettlement(ctx context.Context, id string) (*domain.PriceSupplierSettlement, error) {
	return s.r.GetSupplierSettlement(ctx, strings.TrimSpace(id))
}

func (s *SupplierSettlementService) ListSupplierSettlements(ctx context.Context, inquiryID, itemID *string, page, pageSize int) ([]domain.PriceSupplierSettlement, int64, error) {
	var inquiryPtr *string
	if inquiryID != nil {
		normalized, err := normalizeOptionalWithOriginal(inquiryID)
		if err != nil {
			return nil, 0, err
		}
		inquiryPtr = normalized
	}

	var itemPtr *string
	if itemID != nil {
		normalized, err := normalizeOptionalWithOriginal(itemID)
		if err != nil {
			return nil, 0, err
		}
		itemPtr = normalized
	}

	return s.r.ListSupplierSettlements(ctx, inquiryPtr, itemPtr, page, pageSize)
}

func (s *SupplierSettlementService) UpdateSupplierSettlement(ctx context.Context, params SupplierSettlementUpdateParams) error {
	normalizedSupplierNameSnap, err := normalizeOptionalRequired(params.SupplierNameSnap, "supplier_name_snap")
	if err != nil {
		return err
	}

	normalizedSupplierID, updateSupplierID := normalizeOptional(params.SupplierID)

	repoParams := repo.SupplierSettlementUpdateParams{
		ID:               strings.TrimSpace(params.ID),
		SupplierID:       normalizedSupplierID,
		SupplierNameSnap: normalizedSupplierNameSnap,
		FloatRatioSnap:   params.FloatRatioSnap,
		SettlementPrice:  params.SettlementPrice,
		UpdateSupplierID: updateSupplierID,
		UpdateSettlement: params.SettlementPrice != nil,
	}
	return s.r.UpdateSupplierSettlement(ctx, repoParams)
}

func (s *SupplierSettlementService) SoftDeleteSupplierSettlement(ctx context.Context, id string) error {
	return s.r.SoftDeleteSupplierSettlement(ctx, strings.TrimSpace(id))
}

func (s *SupplierSettlementService) HardDeleteSupplierSettlement(ctx context.Context, id string) error {
	return s.r.HardDeleteSupplierSettlement(ctx, strings.TrimSpace(id))
}

// ========== Helper Functions ==========

func normalizeOptional(str *string) (*string, bool) {
	if str == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(*str)
	if trimmed == "" {
		return nil, true
	}
	normalized := trimmed
	return &normalized, true
}

func normalizeRequiredValue(val string, field string) (string, error) {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return "", fmt.Errorf("%s 不能为空", field)
	}
	return trimmed, nil
}

func normalizeOptionalRequired(str *string, field string) (*string, error) {
	if str == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*str)
	if trimmed == "" {
		return nil, fmt.Errorf("%s 不能为空", field)
	}
	normalized := trimmed
	return &normalized, nil
}

func normalizeOptionalWithOriginal(str *string) (*string, error) {
	if str == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*str)
	if trimmed == "" {
		return nil, nil
	}
	normalized := trimmed
	return &normalized, nil
}
