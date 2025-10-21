package inquiry

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	categoryDomain "hdzk.cn/foodapp/internal/domain/category"
	dictDomain "hdzk.cn/foodapp/internal/domain/dict"
	goodsDomain "hdzk.cn/foodapp/internal/domain/goods"
	inquiryDomain "hdzk.cn/foodapp/internal/domain/inquiry"
	supplierDomain "hdzk.cn/foodapp/internal/domain/supplier"
	"hdzk.cn/foodapp/pkg/logger"
)

type InquiryImportService struct {
	db *gorm.DB
}

func NewInquiryImportService(db *gorm.DB) *InquiryImportService {
	return &InquiryImportService{db: db}
}

// ValidationError Market校验错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// MarketData 解析后的Excel数据
type ExcelData struct {
	Title       string
	InquiryDate time.Time
	Sheets      []SheetData
	Markets     []string       // 询价市场列表
	Suppliers   []SupplierInfo // 供应商列表
}

// SheetData 每个sheet的数据
type SheetData struct {
	SheetName string     // 品类名称
	Items     []ItemData // 商品明细
}

// ItemData 商品明细数据
type ItemData struct {
	GoodsName         string              // 品名
	SpecName          string              // 规格标准
	UnitName          string              // 单位
	LastMonthAvgPrice *float64            // 上月均价
	CurrentAvgPrice   *float64            // 本期均价
	MarketPrices      map[string]*float64 // 市场报价 map[市场名称]价格
}

// SupplierInfo 供应商信息
type SupplierInfo struct {
	Name       string  // 供应商名称
	FloatRatio float64 // 浮动比例（如0.88表示下浮12%）
}

// ValidateExcelStructure 校验Excel文件结构
func (s *InquiryImportService) ValidateExcelStructure(filePath string) (*ExcelData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %w", err)
	}
	defer f.Close()
	logger.L().Warn("开始校验excel")
	// 1. 检查标题（A1单元格）
	firstSheet := f.GetSheetName(0)
	if firstSheet == "" {
		return nil, &ValidationError{Field: "sheet", Message: "Excel文件必须至少包含一个sheet"}
	}

	title, err := f.GetCellValue(firstSheet, "A1")
	if err != nil || strings.TrimSpace(title) == "" {
		return nil, &ValidationError{Field: "title", Message: "Excel必须包含标题（A1单元格）"}
	}

	// 从标题中提取日期信息（如：2025年9月上旬都匀市主要蔬菜类市场参考价）
	inquiryDate, err := extractDateFromTitle(title)
	if err != nil {
		return nil, &ValidationError{Field: "title", Message: fmt.Sprintf("无法从标题提取日期: %v", err)}
	}

	// 2. 获取所有sheet（排除第一行标题行所在sheet后的其他sheet作为品类）
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, &ValidationError{Field: "sheets", Message: "Excel文件必须至少包含一个sheet"}
	}

	excelData := &ExcelData{
		Title:       title,
		InquiryDate: inquiryDate,
		Sheets:      []SheetData{},
		Markets:     []string{},
		Suppliers:   []SupplierInfo{},
	}

	// 用于去重
	marketsMap := make(map[string]bool)
	suppliersMap := make(map[string]float64)

	// 3. 遍历每个sheet进行校验
	for _, sheetName := range sheetList {
		// 读取表头（第2行或第3行，取决于表格格式）
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("读取sheet %s 失败: %w", sheetName, err)
		}

		if len(rows) < 3 {
			return nil, &ValidationError{Field: sheetName, Message: "sheet数据行数不足"}
		}

		// 查找表头行（包含"品名"/"规格标准"/"单位"/"本期均价"的行）
		headerRowIdx := -1
		var headerRow []string
		for idx, row := range rows {
			rowStr := strings.Join(row, "")
			if strings.Contains(rowStr, "品名") && strings.Contains(rowStr, "规格标准") &&
				strings.Contains(rowStr, "单位") && strings.Contains(rowStr, "本期均价") {
				headerRowIdx = idx
				headerRow = row
				break
			}
		}

		if headerRowIdx == -1 {
			return nil, &ValidationError{
				Field:   sheetName,
				Message: "未找到必需的表头：品名/规格标准/单位/本期均价",
			}
		}

		// 解析表头，获取列索引
		colMap, markets, suppliers, err := parseHeader(headerRow)
		if err != nil {
			return nil, &ValidationError{Field: sheetName, Message: err.Error()}
		}

		// 4. 检查询价项（市场）
		if len(markets) == 0 {
			return nil, &ValidationError{
				Field:   sheetName,
				Message: "必须包含至少一个询价项（市场）",
			}
		}

		// 5. 检查供应商
		if len(suppliers) == 0 {
			return nil, &ValidationError{
				Field:   sheetName,
				Message: "必须包含至少一个供应商及其浮动比例",
			}
		}

		// 收集市场和供应商信息
		for _, market := range markets {
			if !marketsMap[market] {
				excelData.Markets = append(excelData.Markets, market)
				marketsMap[market] = true
			}
		}

		for _, supplier := range suppliers {
			if _, exists := suppliersMap[supplier.Name]; !exists {
				excelData.Suppliers = append(excelData.Suppliers, supplier)
				suppliersMap[supplier.Name] = supplier.FloatRatio
			}
		}

		// 解析数据行
		sheetData := SheetData{
			SheetName: sheetName,
			Items:     []ItemData{},
		}

		for i := headerRowIdx + 1; i < len(rows); i++ {
			row := rows[i]
			if len(row) == 0 {
				continue
			}

			// 检查品名是否为空
			goodsName := getCellValue(row, colMap["品名"])
			if strings.TrimSpace(goodsName) == "" {
				continue // 跳过空行
			}

			item := ItemData{
				GoodsName:    goodsName,
				SpecName:     getCellValue(row, colMap["规格标准"]),
				UnitName:     getCellValue(row, colMap["单位"]),
				MarketPrices: make(map[string]*float64),
			}

			// 解析上月均价
			if colIdx, ok := colMap["上月均价"]; ok {
				if price := parseFloat(getCellValue(row, colIdx)); price != nil {
					item.LastMonthAvgPrice = price
				}
			}

			// 解析本期均价
			if colIdx, ok := colMap["本期均价"]; ok {
				if price := parseFloat(getCellValue(row, colIdx)); price != nil {
					item.CurrentAvgPrice = price
				}
			}

			// 解析市场报价
			for _, market := range markets {
				if colIdx, ok := colMap[market]; ok {
					price := parseFloat(getCellValue(row, colIdx))
					item.MarketPrices[market] = price
				}
			}

			sheetData.Items = append(sheetData.Items, item)
		}

		if len(sheetData.Items) == 0 {
			return nil, &ValidationError{
				Field:   sheetName,
				Message: "sheet中没有有效的商品数据",
			}
		}

		logger.L().Info("excel:", zap.String("sheet_name:", sheetData.SheetName))

		excelData.Sheets = append(excelData.Sheets, sheetData)
	}

	return excelData, nil
}

// ImportExcelData 导入Excel数据到数据库
func (s *InquiryImportService) ImportExcelData(ctx context.Context, excelData *ExcelData, orgID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建或获取询价单
		inquiry := &inquiryDomain.BasePriceInquiry{
			OrgID:        orgID,
			InquiryTitle: excelData.Title,
			InquiryDate:  excelData.InquiryDate,
		}
		if err := tx.Create(inquiry).Error; err != nil {
			return fmt.Errorf("创建询价单失败: %w", err)
		}

		// 2. 处理市场
		marketIDMap := make(map[string]string)
		for _, marketName := range excelData.Markets {
			marketID, err := s.getOrCreateMarket(tx, marketName, orgID)
			if err != nil {
				return fmt.Errorf("处理市场 %s 失败: %w", marketName, err)
			}
			marketIDMap[marketName] = marketID
		}

		// 3. 处理供应商
		supplierIDMap := make(map[string]string)
		for _, supplier := range excelData.Suppliers {
			supplierID, err := s.getOrCreateSupplier(tx, supplier.Name, supplier.FloatRatio, orgID)
			if err != nil {
				return fmt.Errorf("处理供应商 %s 失败: %w", supplier.Name, err)
			}
			supplierIDMap[supplier.Name] = supplierID
		}

		// 4. 遍历每个sheet（品类）
		for _, sheet := range excelData.Sheets {
			// 处理品类
			categoryID, err := s.getOrCreateCategory(tx, sheet.SheetName, orgID)
			if err != nil {
				return fmt.Errorf("处理品类 %s 失败: %w", sheet.SheetName, err)
			}

			// 5. 遍历每个商品
			for idx, item := range sheet.Items {
				// 处理规格
				specID, err := s.getOrCreateSpec(tx, item.SpecName)
				if err != nil {
					return fmt.Errorf("处理规格 %s 失败: %w", item.SpecName, err)
				}

				// 处理单位
				unitID, err := s.getOrCreateUnit(tx, item.UnitName)
				if err != nil {
					return fmt.Errorf("处理单位 %s 失败: %w", item.UnitName, err)
				}

				// 处理商品
				goodsID, err := s.getOrCreateGoods(tx, item.GoodsName, categoryID, specID, unitID, orgID)
				if err != nil {
					return fmt.Errorf("处理商品 %s 失败: %w", item.GoodsName, err)
				}

				// 创建询价商品明细
				inquiryItem := &inquiryDomain.PriceInquiryItem{
					InquiryID:         inquiry.ID,
					GoodsID:           goodsID,
					CategoryID:        categoryID,
					SpecID:            &specID,
					UnitID:            &unitID,
					GoodsNameSnap:     item.GoodsName,
					CategoryNameSnap:  sheet.SheetName,
					SpecNameSnap:      &item.SpecName,
					UnitNameSnap:      &item.UnitName,
					LastMonthAvgPrice: item.LastMonthAvgPrice,
					CurrentAvgPrice:   item.CurrentAvgPrice,
					Sort:              idx,
				}
				if err := tx.Create(inquiryItem).Error; err != nil {
					return fmt.Errorf("创建询价商品明细失败: %w", err)
				}

				// 创建市场报价
				for marketName, price := range item.MarketPrices {
					marketID := marketIDMap[marketName]
					marketInquiry := &inquiryDomain.PriceMarketInquiry{
						InquiryID:      inquiry.ID,
						ItemID:         inquiryItem.ID,
						MarketID:       &marketID,
						MarketNameSnap: marketName,
						Price:          price,
					}
					if err := tx.Create(marketInquiry).Error; err != nil {
						return fmt.Errorf("创建市场报价失败: %w", err)
					}
				}

				// 创建供应商结算（结算价 = 本期均价 * 浮动比例）
				for _, supplier := range excelData.Suppliers {
					supplierID := supplierIDMap[supplier.Name]
					var settlementPrice *float64
					if item.CurrentAvgPrice != nil {
						price := *item.CurrentAvgPrice * supplier.FloatRatio
						settlementPrice = &price
					}

					settlement := &inquiryDomain.PriceSupplierSettlement{
						InquiryID:        inquiry.ID,
						ItemID:           inquiryItem.ID,
						SupplierID:       &supplierID,
						SupplierNameSnap: supplier.Name,
						FloatRatioSnap:   supplier.FloatRatio,
						SettlementPrice:  settlementPrice,
					}
					if err := tx.Create(settlement).Error; err != nil {
						return fmt.Errorf("创建供应商结算失败: %w", err)
					}
				}
			}
		}

		return nil
	})
}

// Helper functions

// extractDateFromTitle 从标题中提取日期
func extractDateFromTitle(title string) (time.Time, error) {
	// 匹配格式：2025年9月上旬 或 2025年9月中旬 或 2025年9月下旬
	re := regexp.MustCompile(`(\d{4})年(\d{1,2})月([上中下])旬`)
	matches := re.FindStringSubmatch(title)
	if len(matches) != 4 {
		return time.Time{}, fmt.Errorf("标题格式不正确，无法提取日期信息")
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])

	// 根据旬确定日期
	var day int
	switch matches[3] {
	case "上":
		day = 5 // 上旬取5号
	case "中":
		day = 15 // 中旬取15号
	case "下":
		day = 25 // 下旬取25号
	default:
		day = 15
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// parseHeader 解析表头，获取列索引、市场列表和供应商列表
func parseHeader(headerRow []string) (map[string]int, []string, []SupplierInfo, error) {
	colMap := make(map[string]int)
	markets := []string{}
	suppliers := []SupplierInfo{}

	requiredCols := []string{"品名", "规格标准", "单位", "本期均价"}

	for idx, cell := range headerRow {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			continue
		}

		// 检查必需列
		for _, required := range requiredCols {
			if strings.Contains(cell, required) {
				colMap[required] = idx
			}
		}

		// 检查上月均价
		if strings.Contains(cell, "上月均价") || strings.Contains(cell, "上期均价") {
			colMap["上月均价"] = idx
		}

		// 检查发改委指导价
		if strings.Contains(cell, "发改委") && strings.Contains(cell, "指导价") {
			colMap["发改委指导价"] = idx
		}

		// 识别询价市场（不包含"本期"、"上月"、"发改委"、"供应商"、"结算"等关键字）
		if !strings.Contains(cell, "序号") && !strings.Contains(cell, "品名") &&
			!strings.Contains(cell, "规格") && !strings.Contains(cell, "单位") &&
			!strings.Contains(cell, "发改委") && !strings.Contains(cell, "上月") &&
			!strings.Contains(cell, "本期") && !strings.Contains(cell, "结算") &&
			!strings.Contains(cell, "供应商") && !strings.Contains(cell, "浮动") &&
			!strings.Contains(cell, "下浮") && !strings.Contains(cell, "上浮") {
			// 可能是市场名称
			if !strings.Contains(cell, "均价") {
				markets = append(markets, cell)
				colMap[cell] = idx
			}
		}

		// 识别供应商及浮动比例（如：胡坤本期结算价（下浮12%））
		if strings.Contains(cell, "结算价") || strings.Contains(cell, "供应商") {
			supplier, ratio, err := parseSupplierInfo(cell)
			if err == nil && supplier != "" {
				suppliers = append(suppliers, SupplierInfo{
					Name:       supplier,
					FloatRatio: ratio,
				})
			}
		}
	}

	// 检查必需列
	for _, required := range requiredCols {
		if _, ok := colMap[required]; !ok {
			return nil, nil, nil, fmt.Errorf("缺少必需列: %s", required)
		}
	}

	return colMap, markets, suppliers, nil
}

// parseSupplierInfo 解析供应商信息
// 如："胡坤本期结算价（下浮12%）" -> ("胡坤", 0.88)
// 如："贵海本期结算价（下浮14%）" -> ("贵海", 0.86)
func parseSupplierInfo(cell string) (string, float64, error) {
	// 匹配格式：XXX本期结算价（下浮/上浮 XX%）
	re := regexp.MustCompile(`(.+?)(?:本期)?结算价.*?([下上])浮\s*(\d+(?:\.\d+)?)%`)
	matches := re.FindStringSubmatch(cell)
	if len(matches) != 4 {
		return "", 1.0, fmt.Errorf("无法解析供应商信息")
	}

	name := strings.TrimSpace(matches[1])
	direction := matches[2] // "下" 或 "上"
	percent, _ := strconv.ParseFloat(matches[3], 64)

	var ratio float64
	if direction == "下" {
		ratio = 1.0 - (percent / 100.0)
	} else {
		ratio = 1.0 + (percent / 100.0)
	}

	return name, ratio, nil
}

// getCellValue 安全获取单元格值
func getCellValue(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// parseFloat 解析浮点数
func parseFloat(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &val
}

// getOrCreateCategory 获取或创建品类
func (s *InquiryImportService) getOrCreateCategory(tx *gorm.DB, name string, orgID string) (string, error) {
	var category categoryDomain.Category
	err := tx.Where("name = ? AND org_id = ? AND is_deleted = 0", name, orgID).First(&category).Error
	if err == nil {
		return category.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	// 创建新品类
	category = categoryDomain.Category{
		Name:  name,
		OrgID: orgID,
	}
	if err := tx.Create(&category).Error; err != nil {
		return "", err
	}

	return category.ID, nil
}

// getOrCreateSpec 获取或创建规格
func (s *InquiryImportService) getOrCreateSpec(tx *gorm.DB, name string) (string, error) {
	if name == "" {
		name = "默认"
	}

	var spec dictDomain.Spec
	err := tx.Where("name = ? AND is_deleted = 0", name).First(&spec).Error
	if err == nil {
		return spec.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	// 创建新规格
	spec = dictDomain.Spec{
		Name: name,
	}
	if err := tx.Create(&spec).Error; err != nil {
		return "", err
	}

	return spec.ID, nil
}

// getOrCreateUnit 获取或创建单位
func (s *InquiryImportService) getOrCreateUnit(tx *gorm.DB, name string) (string, error) {
	if name == "" {
		name = "个"
	}

	var unit dictDomain.Unit
	err := tx.Where("name = ? AND is_deleted = 0", name).First(&unit).Error
	if err == nil {
		return unit.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	// 创建新单位
	unit = dictDomain.Unit{
		Name: name,
	}
	if err := tx.Create(&unit).Error; err != nil {
		return "", err
	}

	return unit.ID, nil
}

// getOrCreateGoods 获取或创建商品
func (s *InquiryImportService) getOrCreateGoods(tx *gorm.DB, name string, categoryID, specID, unitID, orgID string) (string, error) {
	var goods goodsDomain.Goods
	err := tx.Where("name = ? AND category_id = ? AND spec_id = ? AND unit_id = ? AND org_id = ? AND is_deleted = 0",
		name, categoryID, specID, unitID, orgID).First(&goods).Error
	if err == nil {
		return goods.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	// 创建新商品
	goods = goodsDomain.Goods{
		Name:       name,
		CategoryID: categoryID,
		SpecID:     specID,
		UnitID:     unitID,
		OrgID:      orgID,
	}
	if err := tx.Create(&goods).Error; err != nil {
		return "", err
	}

	return goods.ID, nil
}

// getOrCreateMarket 获取或创建市场
func (s *InquiryImportService) getOrCreateMarket(tx *gorm.DB, name string, orgID string) (string, error) {
	var market inquiryDomain.BaseMarket
	err := tx.Where("name = ? AND org_id = ? AND is_deleted = 0", name, orgID).First(&market).Error
	if err == nil {
		return market.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	// 创建新市场
	market = inquiryDomain.BaseMarket{
		Name:  name,
		OrgID: orgID,
	}
	if err := tx.Create(&market).Error; err != nil {
		return "", err
	}

	return market.ID, nil
}

// getOrCreateSupplier 获取或创建供应商，如果存在但浮动比例不同则更新
func (s *InquiryImportService) getOrCreateSupplier(tx *gorm.DB, name string, floatRatio float64, orgID string) (string, error) {
	var supplier supplierDomain.Supplier
	err := tx.Where("name = ? AND org_id = ? AND is_deleted = 0", name, orgID).First(&supplier).Error

	if err == nil {
		// 存在但浮动比例不同，更新
		if supplier.FloatRatio != floatRatio {
			if err := tx.Model(&supplier).Update("float_ratio", floatRatio).Error; err != nil {
				return "", err
			}
		}
		return supplier.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return "", err
	}

	// 创建新供应商
	supplier = supplierDomain.Supplier{
		Name:        name,
		FloatRatio:  floatRatio,
		OrgID:       orgID,
		Description: "从Excel导入自动创建",
	}
	if err := tx.Create(&supplier).Error; err != nil {
		return "", err
	}

	return supplier.ID, nil
}
