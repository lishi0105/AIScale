package inquiry

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	categoryDomain "hdzk.cn/foodapp/internal/domain/category"
	dictDomain "hdzk.cn/foodapp/internal/domain/dict"
	goodsDomain "hdzk.cn/foodapp/internal/domain/goods"
	domain "hdzk.cn/foodapp/internal/domain/inquiry"
	supplierDomain "hdzk.cn/foodapp/internal/domain/supplier"
	categoryRepo "hdzk.cn/foodapp/internal/repository/category"
	dictRepo "hdzk.cn/foodapp/internal/repository/dict"
	goodsRepo "hdzk.cn/foodapp/internal/repository/goods"
	supplierRepo "hdzk.cn/foodapp/internal/repository/supplier"
)

type ExcelImportParams struct {
	FilePath string
	OrgID    string
}

type SupplierInfo struct {
	Name       string
	FloatRatio float64
}

// ImportFromExcel imports price inquiry records from an Excel file
func (s *Service) ImportFromExcel(ctx context.Context, params ExcelImportParams) error {
	// Open Excel file
	f, err := excelize.OpenFile(params.FilePath)
	if err != nil {
		return fmt.Errorf("打开Excel文件失败: %w", err)
	}
	defer f.Close()

	// Get all sheet names
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return errors.New("Excel文件中没有工作表")
	}

	// Get the database from repository
	db := s.r.GetDB()

	// Process all sheets
	for _, sheetName := range sheets {
		if err := s.processSheet(ctx, f, sheetName, params.OrgID, db); err != nil {
			return fmt.Errorf("处理工作表 %s 失败: %w", sheetName, err)
		}
	}

	return nil
}

func (s *Service) processSheet(ctx context.Context, f *excelize.File, sheetName string, orgID string, db *gorm.DB) error {
	// Get all rows from the sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取工作表行失败: %w", err)
	}

	if len(rows) < 3 {
		return errors.New("Excel表格行数不足，至少需要3行（标题行+表头行+数据行）")
	}

	// Step 1: Validate and extract title
	title := strings.TrimSpace(rows[0][0])
	if title == "" {
		return errors.New("Excel必须包含标题行")
	}

	// Step 2: Find category name (should be in the tab name or extract from title)
	categoryName := sheetName

	// Step 3: Validate table structure
	headerRowIdx := -1
	for i, row := range rows {
		if len(row) > 0 && containsRequiredColumns(row) {
			headerRowIdx = i
			break
		}
	}

	if headerRowIdx == -1 {
		return errors.New("Excel表格必须包含品名/规格标准/单位/本期均价四项")
	}

	headerRow := rows[headerRowIdx]

	// Step 4: Extract inquiry markets and suppliers from header
	markets, suppliers, err := extractMarketsAndSuppliers(headerRow)
	if err != nil {
		return err
	}

	if len(markets) < 3 {
		return fmt.Errorf("Excel表格必须包含至少3个询价项，当前只有%d个", len(markets))
	}

	if len(suppliers) < 1 {
		return errors.New("Excel表格必须包含至少1个供应商")
	}

	// Start transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// Create repositories with transaction
		categoryRepository := categoryRepo.NewRepository(tx)
		dictRepository := dictRepo.NewRepository(tx)
		goodsRepository := goodsRepo.NewRepository(tx)
		supplierRepository := supplierRepo.NewRepository(tx)

		// Step 5: Find or create category
		category, err := s.findOrCreateCategory(ctx, categoryRepository, categoryName, orgID)
		if err != nil {
			return fmt.Errorf("查找或创建品类失败: %w", err)
		}

		// Step 6: Create price inquiry
		inquiry := &domain.PriceInquiry{
			OrgID:        orgID,
			InquiryTitle: title,
			InquiryDate:  parseInquiryDate(title),
		}
		if len(markets) > 0 {
			inquiry.Market1 = &markets[0]
		}
		if len(markets) > 1 {
			inquiry.Market2 = &markets[1]
		}
		if len(markets) > 2 {
			inquiry.Market3 = &markets[2]
		}
		if err := tx.Create(inquiry).Error; err != nil {
			return fmt.Errorf("创建询价记录失败: %w", err)
		}

		// Step 7: Process each data row
		for i := headerRowIdx + 1; i < len(rows); i++ {
			row := rows[i]
			if len(row) < 4 {
				continue // Skip empty rows
			}

			goodsName := strings.TrimSpace(row[0])
			if goodsName == "" {
				continue
			}

			// Process this goods item
			if err := s.processGoodsRow(ctx, tx, row, headerRow, category.ID, orgID, inquiry.ID,
				dictRepository, goodsRepository, supplierRepository, suppliers); err != nil {
				return fmt.Errorf("处理商品 %s 失败: %w", goodsName, err)
			}
		}

		return nil
	})
}

func (s *Service) processGoodsRow(
	ctx context.Context,
	tx *gorm.DB,
	row []string,
	headerRow []string,
	categoryID string,
	orgID string,
	inquiryID string,
	dictRepository dictRepo.DictRepository,
	goodsRepository goodsRepo.GoodsRepository,
	supplierRepository supplierRepo.SupplierRepository,
	suppliers []SupplierInfo,
) error {
	// Extract goods info
	goodsName := strings.TrimSpace(row[0])
	specName := ""
	if len(row) > 1 {
		specName = strings.TrimSpace(row[1])
	}
	unitName := ""
	if len(row) > 2 {
		unitName = strings.TrimSpace(row[2])
	}

	// Find or create spec
	spec, err := s.findOrCreateSpec(ctx, dictRepository, specName)
	if err != nil {
		return fmt.Errorf("查找或创建规格失败: %w", err)
	}

	// Find or create unit
	unit, err := s.findOrCreateUnit(ctx, dictRepository, unitName)
	if err != nil {
		return fmt.Errorf("查找或创建单位失败: %w", err)
	}

	// Find or create goods
	goods, err := s.findOrCreateGoods(ctx, goodsRepository, goodsName, spec.ID, unit.ID, categoryID, orgID)
	if err != nil {
		return fmt.Errorf("查找或创建商品失败: %w", err)
	}

	// Extract market prices - need to find market columns from header
	marketPrices := make([]*float64, 3)
	marketColIndices := findMarketColumns(headerRow)
	for i := 0; i < len(marketColIndices) && i < 3; i++ {
		colIdx := marketColIndices[i]
		if colIdx < len(row) {
			if price := parsePrice(row[colIdx]); price != nil {
				marketPrices[i] = price
			}
		}
	}

	// Extract average price
	var avgPrice *float64
	avgPriceColIdx := findColumnIndex(headerRow, "本期均价")
	if avgPriceColIdx >= 0 && avgPriceColIdx < len(row) {
		avgPrice = parsePrice(row[avgPriceColIdx])
	}

	// Create goods avg detail
	avgDetail := &domain.GoodsAvgDetail{
		GoodsID:      goods.ID,
		InquiryID:    inquiryID,
		Market1Price: marketPrices[0],
		Market2Price: marketPrices[1],
		Market3Price: marketPrices[2],
		AvgPrice:     avgPrice,
		OrgID:        &orgID,
	}
	if err := tx.Create(avgDetail).Error; err != nil {
		return fmt.Errorf("创建商品均价明细失败: %w", err)
	}

	// Process each supplier
	for _, supplierInfo := range suppliers {
		// Find or update supplier
		supplier, err := s.findOrUpdateSupplier(ctx, supplierRepository, supplierInfo.Name, supplierInfo.FloatRatio, orgID)
		if err != nil {
			return fmt.Errorf("查找或更新供应商失败: %w", err)
		}

		// Find supplier price column
		supplierColIdx := findSupplierPriceColumn(headerRow, supplierInfo.Name)
		if supplierColIdx < 0 || supplierColIdx >= len(row) {
			continue
		}

		unitPrice := parsePrice(row[supplierColIdx])
		if unitPrice == nil {
			continue
		}

		// Create goods price
		goodsPrice := &domain.GoodsPrice{
			GoodsID:    goods.ID,
			SupplierID: supplier.ID,
			InquiryID:  inquiryID,
			UnitPrice:  *unitPrice,
			FloatRatio: supplierInfo.FloatRatio,
			OrgID:      &orgID,
		}
		if err := tx.Create(goodsPrice).Error; err != nil {
			return fmt.Errorf("创建商品单价失败: %w", err)
		}
	}

	return nil
}

func (s *Service) findOrCreateCategory(ctx context.Context, repo categoryRepo.CategoryRepository, name string, orgID string) (*categoryDomain.Category, error) {
	category, err := repo.FindByName(ctx, name, orgID)
	if err == nil {
		return category, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new category
	category = &categoryDomain.Category{
		Name:  name,
		OrgID: orgID,
	}
	if err := repo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *Service) findOrCreateSpec(ctx context.Context, repo dictRepo.DictRepository, name string) (*dictDomain.Spec, error) {
	if name == "" {
		name = "未分类"
	}

	spec, err := repo.FindSpecByName(ctx, name)
	if err == nil {
		return spec, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new spec
	spec = &dictDomain.Spec{
		Name: name,
	}
	if err := repo.CreateSpec(ctx, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func (s *Service) findOrCreateUnit(ctx context.Context, repo dictRepo.DictRepository, name string) (*dictDomain.Unit, error) {
	if name == "" {
		name = "个"
	}

	unit, err := repo.FindUnitByName(ctx, name)
	if err == nil {
		return unit, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new unit
	unit = &dictDomain.Unit{
		Name: name,
	}
	if err := repo.CreateUnit(ctx, unit); err != nil {
		return nil, err
	}
	return unit, nil
}

func (s *Service) findOrCreateGoods(ctx context.Context, repo goodsRepo.GoodsRepository, name string, specID string, unitID string, categoryID string, orgID string) (*goodsDomain.Goods, error) {
	goods, err := repo.FindByNameAndSpec(ctx, name, specID, unitID, orgID)
	if err == nil {
		return goods, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new goods
	goods = &goodsDomain.Goods{
		Name:       name,
		SpecID:     specID,
		UnitID:     unitID,
		CategoryID: categoryID,
		OrgID:      orgID,
	}
	if err := repo.CreateGoods(ctx, goods); err != nil {
		return nil, err
	}
	return goods, nil
}

func (s *Service) findOrUpdateSupplier(ctx context.Context, repo supplierRepo.SupplierRepository, name string, floatRatio float64, orgID string) (*supplierDomain.Supplier, error) {
	supplier, err := repo.FindByName(ctx, name, orgID)
	if err == nil {
		// Update float ratio if different
		if supplier.FloatRatio != floatRatio {
			updateParams := supplierRepo.UpdateParams{
				ID:         supplier.ID,
				FloatRatio: &floatRatio,
			}
			if err := repo.UpdateSupplier(ctx, updateParams); err != nil {
				return nil, err
			}
			supplier.FloatRatio = floatRatio
		}
		return supplier, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new supplier
	supplier = &supplierDomain.Supplier{
		Name:        name,
		FloatRatio:  floatRatio,
		OrgID:       orgID,
		Description: "",
		Status:      1,
	}
	if err := repo.CreateSupplier(ctx, supplier); err != nil {
		return nil, err
	}
	return supplier, nil
}

// Helper functions

func containsRequiredColumns(row []string) bool {
	hasGoodsName := false
	hasSpec := false
	hasUnit := false
	hasAvgPrice := false

	for _, cell := range row {
		cell = strings.TrimSpace(cell)
		if strings.Contains(cell, "品名") {
			hasGoodsName = true
		}
		if strings.Contains(cell, "规格") || strings.Contains(cell, "标准") {
			hasSpec = true
		}
		if strings.Contains(cell, "单位") {
			hasUnit = true
		}
		if strings.Contains(cell, "本期均价") {
			hasAvgPrice = true
		}
	}

	return hasGoodsName && hasSpec && hasUnit && hasAvgPrice
}

func extractMarketsAndSuppliers(headerRow []string) ([]string, []SupplierInfo, error) {
	markets := []string{}
	suppliers := []SupplierInfo{}

	// Patterns for identifying columns
	// Match columns like "富万家超市", "育英巷菜市场", "大润发"
	marketPattern := regexp.MustCompile(`(超市|市场|商店|发)`)
	// Match columns like "胡坤本期结算价（下浮12%）", "贵海本期结算价（下浮14%）"
	supplierPattern := regexp.MustCompile(`^(.*?)本期结算价.*?[（(].*?下浮.*?(\d+(?:\.\d+)?)[%％][)）]`)

	for _, cell := range headerRow {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			continue
		}

		// Check if it's a supplier column first (more specific pattern)
		matches := supplierPattern.FindStringSubmatch(cell)
		if len(matches) >= 3 {
			supplierName := strings.TrimSpace(matches[1])
			floatRatioStr := matches[2]
			floatRatioPercent, err := strconv.ParseFloat(floatRatioStr, 64)
			if err != nil {
				continue
			}
			// Convert percentage to ratio (e.g., 12% -> 0.88)
			floatRatio := 1.0 - (floatRatioPercent / 100.0)
			suppliers = append(suppliers, SupplierInfo{
				Name:       supplierName,
				FloatRatio: floatRatio,
			})
			continue
		}

		// Check if it's a market column
		if marketPattern.MatchString(cell) && !strings.Contains(cell, "本期") {
			markets = append(markets, cell)
			continue
		}
	}

	return markets, suppliers, nil
}

func findColumnIndex(headerRow []string, pattern string) int {
	for i, cell := range headerRow {
		if strings.Contains(strings.TrimSpace(cell), pattern) {
			return i
		}
	}
	return -1
}

func getMarketColumnPattern(index int) string {
	// Match market columns based on their content
	// The actual matching is done by looking for market names in the header
	// This function is used to identify which column contains which market price
	return ""
}

func findSupplierPriceColumn(headerRow []string, supplierName string) int {
	for i, cell := range headerRow {
		if strings.Contains(cell, supplierName) && strings.Contains(cell, "结算价") {
			return i
		}
	}
	return -1
}

func findMarketColumns(headerRow []string) []int {
	indices := []int{}
	marketPattern := regexp.MustCompile(`(超市|市场|商店|发)`)
	
	for i, cell := range headerRow {
		cell = strings.TrimSpace(cell)
		if marketPattern.MatchString(cell) && !strings.Contains(cell, "本期") && !strings.Contains(cell, "结算") {
			indices = append(indices, i)
		}
	}
	
	return indices
}

func parsePrice(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Remove any non-numeric characters except decimal point
	s = regexp.MustCompile(`[^\d.]`).ReplaceAllString(s, "")
	if s == "" {
		return nil
	}

	price, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}

	return &price
}

func parseInquiryDate(title string) time.Time {
	// Try to extract date from title
	// Example: "2025年9月上旬都匀市主要蔬菜类市场参考价"
	// For now, return current date
	// TODO: Implement proper date parsing
	return time.Now()
}
