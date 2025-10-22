package inquiry

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	domain "hdzk.cn/foodapp/internal/domain/inquiry"
	repo "hdzk.cn/foodapp/internal/repository/inquiry"
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
	r     repo.InquiryRepository
	itemR repo.InquiryItemRepository
}

func NewInquiryService(r repo.InquiryRepository, itemR repo.InquiryItemRepository) *InquiryService {
	return &InquiryService{r: r, itemR: itemR}
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

// ImportExcel 从Excel文件导入询价单和商品明细
func (s *InquiryService) ImportExcel(ctx context.Context, reader io.Reader, orgID string) (*domain.BasePriceInquiry, int, error) {
	// 读取Excel文件
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("打开Excel文件失败: %w", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, 0, fmt.Errorf("Excel文件中没有工作表")
	}
	sheetName := sheets[0]

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, 0, fmt.Errorf("读取Excel数据失败: %w", err)
	}

	if len(rows) < 2 {
		return nil, 0, fmt.Errorf("Excel文件数据不足，至少需要标题行和一行数据")
	}

	// 第一行是标题信息，例如: "2025年9月上旬都匀市场参考价" (最近一次更新时间: 2025-08-31) (单位: 元)
	titleRow := rows[0]
	if len(titleRow) == 0 {
		return nil, 0, fmt.Errorf("Excel文件缺少标题行")
	}

	// 解析标题，提取询价单标题和日期
	inquiryTitle, inquiryDate, err := parseExcelTitle(titleRow[0])
	if err != nil {
		return nil, 0, fmt.Errorf("解析标题失败: %w", err)
	}

	// 创建询价单
	inquiry := &domain.BasePriceInquiry{
		ID:           uuid.NewString(),
		OrgID:        orgID,
		InquiryTitle: inquiryTitle,
		InquiryDate:  inquiryDate,
	}

	if err := s.r.CreateInquiry(ctx, inquiry); err != nil {
		return nil, 0, fmt.Errorf("创建询价单失败: %w", err)
	}

	// 第二行是表头，跳过
	// 从第三行开始是数据
	itemCount := 0
	for i := 2; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 10 {
			continue // 跳过数据不完整的行
		}

		// 解析商品数据
		// 假设列顺序为：序号、商品图、品名、拼音首字母、编码、规格标准、单位、指导价、上期均价、本期均价
		goodsName := strings.TrimSpace(row[2])
		if goodsName == "" {
			continue
		}

		categoryName := "未分类" // 默认分类，可以根据需要调整
		specName := ""
		unitName := ""
		if len(row) > 5 {
			specName = strings.TrimSpace(row[5])
		}
		if len(row) > 6 {
			unitName = strings.TrimSpace(row[6])
		}

		var guidePrice, lastMonthAvgPrice, currentAvgPrice *float64
		if len(row) > 7 && row[7] != "" {
			if price, err := strconv.ParseFloat(strings.TrimSpace(row[7]), 64); err == nil {
				guidePrice = &price
			}
		}
		if len(row) > 8 && row[8] != "" {
			if price, err := strconv.ParseFloat(strings.TrimSpace(row[8]), 64); err == nil {
				lastMonthAvgPrice = &price
			}
		}
		if len(row) > 9 && row[9] != "" {
			if price, err := strconv.ParseFloat(strings.TrimSpace(row[9]), 64); err == nil {
				currentAvgPrice = &price
			}
		}

		// 创建询价商品明细（使用快照，不关联具体的商品ID）
		item := &domain.PriceInquiryItem{
			ID:                uuid.NewString(),
			InquiryID:         inquiry.ID,
			GoodsID:           uuid.NewString(), // 临时ID，表示未关联
			CategoryID:        uuid.NewString(), // 临时ID，表示未关联
			GoodsNameSnap:     goodsName,
			CategoryNameSnap:  categoryName,
			GuidePrice:        guidePrice,
			LastMonthAvgPrice: lastMonthAvgPrice,
			CurrentAvgPrice:   currentAvgPrice,
			Sort:              i - 2,
		}

		if specName != "" {
			item.SpecNameSnap = &specName
		}
		if unitName != "" {
			item.UnitNameSnap = &unitName
		}

		// 创建询价商品明细
		if err := s.itemR.CreateInquiryItem(ctx, item); err != nil {
			// 忽略创建失败的项，继续处理下一项
			continue
		}
		itemCount++
	}

	return inquiry, itemCount, nil
}

// parseExcelTitle 解析Excel标题，提取询价单标题和日期
// 例如: "2025年9月上旬都匀市场参考价 (最近一次更新时间: 2025-08-31) (单位: 元)"
func parseExcelTitle(titleStr string) (string, time.Time, error) {
	titleStr = strings.TrimSpace(titleStr)
	if titleStr == "" {
		return "", time.Time{}, fmt.Errorf("标题为空")
	}

	// 提取日期（在括号中）
	dateStr := ""
	if strings.Contains(titleStr, "(最近一次更新时间:") {
		start := strings.Index(titleStr, "(最近一次更新时间:") + len("(最近一次更新时间:")
		end := strings.Index(titleStr[start:], ")")
		if end > 0 {
			dateStr = strings.TrimSpace(titleStr[start : start+end])
		}
	}

	// 如果没有找到日期，使用当前日期
	var inquiryDate time.Time
	var err error
	if dateStr != "" {
		inquiryDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			inquiryDate = time.Now()
		}
	} else {
		inquiryDate = time.Now()
	}

	// 提取标题（去除括号部分）
	inquiryTitle := titleStr
	if idx := strings.Index(titleStr, "("); idx > 0 {
		inquiryTitle = strings.TrimSpace(titleStr[:idx])
	}

	return inquiryTitle, inquiryDate, nil
}

// ========== PriceInquiryItem Service ==========

type InquiryItemService struct {
	r repo.InquiryItemRepository
}

func NewInquiryItemService(r repo.InquiryItemRepository) *InquiryItemService {
	return &InquiryItemService{r: r}
}

type InquiryItemCreateParams struct {
	InquiryID         string
	GoodsID           string
	CategoryID        string
	SpecID            *string
	UnitID            *string
	GoodsNameSnap     string
	CategoryNameSnap  string
	SpecNameSnap      *string
	UnitNameSnap      *string
	GuidePrice        *float64
	LastMonthAvgPrice *float64
	CurrentAvgPrice   *float64
	Sort              *int
}

type InquiryItemUpdateParams struct {
	ID                string
	GoodsID           *string
	CategoryID        *string
	SpecID            *string
	UnitID            *string
	GoodsNameSnap     *string
	CategoryNameSnap  *string
	SpecNameSnap      *string
	UnitNameSnap      *string
	GuidePrice        *float64
	LastMonthAvgPrice *float64
	CurrentAvgPrice   *float64
	Sort              *int
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
		ID:                uuid.NewString(),
		InquiryID:         inquiryID,
		GoodsID:           goodsID,
		CategoryID:        categoryID,
		SpecID:            normalizedSpecID,
		UnitID:            normalizedUnitID,
		GoodsNameSnap:     goodsNameSnap,
		CategoryNameSnap:  categoryNameSnap,
		SpecNameSnap:      normalizedSpecNameSnap,
		UnitNameSnap:      normalizedUnitNameSnap,
		GuidePrice:        params.GuidePrice,
		LastMonthAvgPrice: params.LastMonthAvgPrice,
		CurrentAvgPrice:   params.CurrentAvgPrice,
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
	GoodsID        string
	MarketID       *string
	MarketNameSnap string
	InquiryDate    time.Time
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
	goodsID, err := normalizeRequiredValue(params.GoodsID, "goods_id")
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
		GoodsID:        goodsID,
		InquiryID:      inquiryID,
		ItemID:         itemID,
		MarketID:       normalizedMarketID,
		MarketNameSnap: marketNameSnap,
		InquiryDate:    params.InquiryDate,
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
