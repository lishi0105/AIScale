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
	"hdzk.cn/foodapp/pkg/utils"
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

func safeDeref(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// ValidateExcelStructure 校验Excel文件结构
func (s *InquiryImportService) ValidateExcelStructure(filePath string) (*ExcelData, error) {
	logger.L().Info("开始校验Excel文件结构", zap.String("file_path", filePath))

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		logger.L().Error("打开Excel文件失败", zap.String("file_path", filePath), zap.Error(err))
		return nil, fmt.Errorf("打开Excel文件失败: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logger.L().Warn("关闭Excel文件时出错", zap.String("file_path", filePath), zap.Error(closeErr))
		}
	}()

	// 1. 检查第一个sheet
	firstSheet := f.GetSheetName(0)
	logger.L().Debug("获取第一个sheet名称", zap.String("first_sheet", firstSheet))
	if firstSheet == "" {
		errMsg := "Excel文件必须至少包含一个sheet"
		logger.L().Error(errMsg, zap.String("file_path", filePath))
		return nil, &ValidationError{Field: "sheet", Message: errMsg}
	}

	// 2. 读取标题（A1）
	title, err := f.GetCellValue(firstSheet, "A1")
	if err != nil {
		logger.L().Error("读取A1单元格失败", zap.String("sheet", firstSheet), zap.Error(err))
		return nil, &ValidationError{Field: "title", Message: "读取标题（A1）失败"}
	}
	titleRaw := title
	title = strings.TrimSpace(title)
	logger.L().Info("成功读取Excel标题", zap.String("title", title), zap.String("title_raw", titleRaw))

	if title == "" {
		errMsg := "Excel必须包含标题（A1单元格）"
		logger.L().Error(errMsg, zap.String("sheet", firstSheet))
		return nil, &ValidationError{Field: "title", Message: errMsg}
	}

	// 3. 从标题提取日期
	inquiryDate, err := extractDateFromTitle(title)
	if err != nil {
		logger.L().Error("从标题提取日期失败", zap.String("title", title), zap.Error(err))
		return nil, &ValidationError{Field: "title", Message: fmt.Sprintf("无法从标题提取日期: %v", err)}
	}
	logger.L().Info("成功从标题提取日期", zap.String("title", title), zap.Time("inquiry_date", inquiryDate))

	// 4. 获取所有sheet
	sheetList := f.GetSheetList()
	logger.L().Info("获取到的sheet列表", zap.Strings("sheets", sheetList))
	if len(sheetList) == 0 {
		errMsg := "Excel文件必须至少包含一个sheet"
		logger.L().Error(errMsg, zap.String("file_path", filePath))
		return nil, &ValidationError{Field: "sheets", Message: errMsg}
	}

	excelData := &ExcelData{
		Title:       title,
		InquiryDate: inquiryDate,
		Sheets:      []SheetData{},
		Markets:     []string{},
		Suppliers:   []SupplierInfo{},
	}

	marketsMap := make(map[string]bool)
	suppliersMap := make(map[string]float64)

	// 5. 遍历每个sheet
	for _, sheetName := range sheetList {
		logger.L().Info("开始校验sheet", zap.String("sheet_name", sheetName))

		rows, err := f.GetRows(sheetName)
		if err != nil {
			logger.L().Error("读取sheet行数据失败", zap.String("sheet_name", sheetName), zap.Error(err))
			return nil, fmt.Errorf("读取sheet %s 失败: %w", sheetName, err)
		}

		logger.L().Debug("sheet行数统计", zap.String("sheet_name", sheetName), zap.Int("row_count", len(rows)))

		if len(rows) < 3 {
			errMsg := "sheet数据行数不足（至少需要3行：标题+表头+数据）"
			logger.L().Error(errMsg, zap.String("sheet_name", sheetName), zap.Int("actual_rows", len(rows)))
			return nil, &ValidationError{Field: sheetName, Message: errMsg}
		}

		// 查找表头行
		headerRowIdx := -1
		var headerRow []string
		for idx, row := range rows {
			// 注意：这里按你的要求使用逗号拼接
			rawRowStr := strings.Join(row, ",") // ← 已按你的要求改为 ","
			rowStr := utils.RemoveAllSpaces(rawRowStr)
			logger.L().Debug("检查行是否为表头",
				zap.String("sheet_name", sheetName),
				zap.Int("row_index", idx),
				zap.String("row_raw", rawRowStr),
				zap.String("row_no_space", rowStr))

			if strings.Contains(rowStr, "品名") &&
				strings.Contains(rowStr, "规格标准") &&
				strings.Contains(rowStr, "单位") &&
				strings.Contains(rowStr, "本期均价") {
				headerRowIdx = idx
				headerRow = row
				logger.L().Info("找到表头行",
					zap.String("sheet_name", sheetName),
					zap.Int("header_row_index", idx),
					zap.Strings("header_columns_raw", row))
				break
			}
		}

		if headerRowIdx == -1 {
			errMsg := "未找到必需的表头：品名/规格标准/单位/本期均价"
			logger.L().Error(errMsg, zap.String("sheet_name", sheetName))
			return nil, &ValidationError{Field: sheetName, Message: errMsg}
		}

		// 解析表头
		colMap, markets, suppliers, err := parseHeader(headerRow)
		if err != nil {
			logger.L().Error("解析表头失败",
				zap.String("sheet_name", sheetName),
				zap.Strings("header_row", headerRow),
				zap.Error(err))
			return nil, &ValidationError{Field: sheetName, Message: err.Error()}
		}

		logger.L().Info("表头解析结果",
			zap.String("sheet_name", sheetName),
			zap.Any("column_map", colMap),
			zap.Strings("markets", markets),
			zap.Any("suppliers", suppliers))

		if len(markets) == 0 {
			errMsg := "必须包含至少一个询价项（市场）"
			logger.L().Error(errMsg, zap.String("sheet_name", sheetName))
			return nil, &ValidationError{Field: sheetName, Message: errMsg}
		}

		if len(suppliers) == 0 {
			errMsg := "必须包含至少一个供应商及其浮动比例"
			logger.L().Error(errMsg, zap.String("sheet_name", sheetName))
			return nil, &ValidationError{Field: sheetName, Message: errMsg}
		}

		// 收集全局市场和供应商
		for _, market := range markets {
			if !marketsMap[market] {
				excelData.Markets = append(excelData.Markets, market)
				marketsMap[market] = true
				logger.L().Debug("新增市场", zap.String("market", market))
			}
		}

		for _, supplier := range suppliers {
			if _, exists := suppliersMap[supplier.Name]; !exists {
				excelData.Suppliers = append(excelData.Suppliers, supplier)
				suppliersMap[supplier.Name] = supplier.FloatRatio
				logger.L().Debug("新增供应商",
					zap.String("supplier_name", supplier.Name),
					zap.Float64("float_ratio", supplier.FloatRatio))
			} else if suppliersMap[supplier.Name] != supplier.FloatRatio {
				// 同名不同比例时记录一下（最终以后续入库逻辑为准）
				logger.L().Warn("检测到同名供应商浮动比例不一致（将以首次出现为准汇总，入库时以最新为准）",
					zap.String("supplier_name", supplier.Name),
					zap.Float64("existing_ratio", suppliersMap[supplier.Name]),
					zap.Float64("new_ratio", supplier.FloatRatio))
			}
		}

		// 解析数据行
		sheetData := SheetData{SheetName: sheetName, Items: []ItemData{}}

		logger.L().Info("开始解析数据行",
			zap.String("sheet_name", sheetName),
			zap.Int("data_row_start", headerRowIdx+1),
			zap.Int("total_rows", len(rows)))

		for i := headerRowIdx + 1; i < len(rows); i++ {
			row := rows[i]
			if len(row) == 0 {
				logger.L().Debug("跳过空行", zap.String("sheet_name", sheetName), zap.Int("row_index", i))
				continue
			}

			goodsName := getCellValue(row, colMap["品名"])
			goodsNameRaw := goodsName
			goodsName = strings.TrimSpace(goodsName)
			if goodsName == "" {
				logger.L().Debug("跳过空商品行", zap.Int("row_index", i), zap.Any("row_cells", row))
				continue
			}

			item := ItemData{
				GoodsName:    goodsName,
				SpecName:     getCellValue(row, colMap["规格标准"]),
				UnitName:     getCellValue(row, colMap["单位"]),
				MarketPrices: make(map[string]*float64),
			}
			logger.L().Debug("解析基本字段",
				zap.Int("row_index", i),
				zap.String("goods_raw", goodsNameRaw),
				zap.String("goods", item.GoodsName),
				zap.String("spec", item.SpecName),
				zap.String("unit", item.UnitName))

			// 上月均价
			if colIdx, ok := colMap["上月均价"]; ok {
				rawVal := getCellValue(row, colIdx)
				price := parseFloat(rawVal)
				item.LastMonthAvgPrice = price
				logger.L().Debug("解析上月均价",
					zap.String("goods", goodsName),
					zap.String("raw_value", rawVal),
					zap.Float64("parsed_price", safeDeref(price)))
			}

			// 本期均价
			if colIdx, ok := colMap["本期均价"]; ok {
				rawVal := getCellValue(row, colIdx)
				price := parseFloat(rawVal)
				item.CurrentAvgPrice = price
				logger.L().Debug("解析本期均价",
					zap.String("goods", goodsName),
					zap.String("raw_value", rawVal),
					zap.Float64("parsed_price", safeDeref(price)))
			}

			// 市场报价
			for _, market := range markets {
				if colIdx, ok := colMap[market]; ok {
					rawVal := getCellValue(row, colIdx)
					price := parseFloat(rawVal)
					item.MarketPrices[market] = price
					logger.L().Debug("解析市场报价",
						zap.String("goods", goodsName),
						zap.String("market", market),
						zap.String("raw_value", rawVal),
						zap.Float64("parsed_price", safeDeref(price)))
				}
			}

			sheetData.Items = append(sheetData.Items, item)
			logger.L().Debug("成功解析商品项",
				zap.String("sheet", sheetName),
				zap.Int("row_index", i),
				zap.String("goods", item.GoodsName),
				zap.String("spec", item.SpecName),
				zap.String("unit", item.UnitName))
		}

		if len(sheetData.Items) == 0 {
			errMsg := "sheet中没有有效的商品数据"
			logger.L().Error(errMsg, zap.String("sheet_name", sheetName))
			return nil, &ValidationError{Field: sheetName, Message: errMsg}
		}

		excelData.Sheets = append(excelData.Sheets, sheetData)
		logger.L().Info("完成sheet解析",
			zap.String("sheet_name", sheetName),
			zap.Int("item_count", len(sheetData.Items)))
	}

	logger.L().Info("Excel结构校验成功",
		zap.String("title", excelData.Title),
		zap.Time("inquiry_date", excelData.InquiryDate),
		zap.Int("total_sheets", len(excelData.Sheets)),
		zap.Int("total_markets", len(excelData.Markets)),
		zap.Int("total_suppliers", len(excelData.Suppliers)))

	return excelData, nil
}

// ImportExcelData 导入Excel数据到数据库
func (s *InquiryImportService) ImportExcelData(ctx context.Context, excelData *ExcelData, orgID string) error {
	logger.L().Info("开始导入Excel数据", zap.String("org_id", orgID),
		zap.String("title", excelData.Title), zap.Time("inquiry_date", excelData.InquiryDate))
	return s.db.Transaction(func(tx *gorm.DB) error {
		logger.L().Debug("事务已开启，准备创建询价单")

		// 1. 检查是否已存在相同的询价单
		var existingCount int64
		err := tx.Model(&inquiryDomain.BasePriceInquiry{}).
			Where("org_id = ? AND is_deleted = 0", orgID).
			Where("inquiry_title = ? OR inquiry_date = ?", excelData.Title, excelData.InquiryDate).
			Count(&existingCount).Error
		if err != nil {
			logger.L().Error("检查询价单重复失败", zap.Error(err))
			return fmt.Errorf("检查询价单重复失败: %w", err)
		}
		if existingCount > 0 {
			logger.L().Warn("检测到重复的询价单",
				zap.String("org_id", orgID),
				zap.String("inquiry_title", excelData.Title),
				zap.Time("inquiry_date", excelData.InquiryDate))
			return fmt.Errorf("已存在相同标题或日期的询价单，org_id=%s, title=%s, date=%s",
				orgID, excelData.Title, excelData.InquiryDate.Format("2006-01-02"))
		}

		// 2. 创建询价单
		inquiry := &inquiryDomain.BasePriceInquiry{
			OrgID:        orgID,
			InquiryTitle: excelData.Title,
			InquiryDate:  excelData.InquiryDate,
		}
		if err := tx.Create(inquiry).Error; err != nil {
			logger.L().Error("创建询价单失败", zap.Error(err), zap.Any("inquiry", inquiry))
			return fmt.Errorf("创建询价单失败: %w", err)
		}
		logger.L().Info("询价单创建成功", zap.String("inquiry_id", inquiry.ID))

		// 3. 处理市场
		marketIDMap := make(map[string]string)
		for _, marketName := range excelData.Markets {
			marketID, err := s.getOrCreateMarket(tx, marketName, orgID)
			if err != nil {
				logger.L().Error("处理市场失败", zap.String("market", marketName), zap.Error(err))
				return fmt.Errorf("处理市场 %s 失败: %w", marketName, err)
			}
			marketIDMap[marketName] = marketID
			logger.L().Debug("市场就绪", zap.String("market", marketName), zap.String("market_id", marketID))
		}

		// 4. 处理供应商
		supplierIDMap := make(map[string]string)
		for _, supplier := range excelData.Suppliers {
			supplierID, err := s.getOrCreateSupplier(tx, supplier.Name, supplier.FloatRatio, orgID)
			if err != nil {
				logger.L().Error("处理供应商失败", zap.String("supplier", supplier.Name), zap.Error(err))
				return fmt.Errorf("处理供应商 %s 失败: %w", supplier.Name, err)
			}
			supplierIDMap[supplier.Name] = supplierID
			logger.L().Debug("供应商就绪",
				zap.String("supplier", supplier.Name),
				zap.Float64("ratio", supplier.FloatRatio),
				zap.String("supplier_id", supplierID))
		}

		// 5. 遍历每个sheet（品类）
		for _, sheet := range excelData.Sheets {
			logger.L().Info("处理品类", zap.String("category_name", sheet.SheetName))

			// 处理品类
			categoryID, err := s.getOrCreateCategory(tx, sheet.SheetName, orgID)
			if err != nil {
				logger.L().Error("处理品类失败", zap.String("category", sheet.SheetName), zap.Error(err))
				return fmt.Errorf("处理品类 %s 失败: %w", sheet.SheetName, err)
			}
			logger.L().Debug("品类就绪", zap.String("category", sheet.SheetName), zap.String("category_id", categoryID))

			// 6. 遍历每个商品
			for idx, item := range sheet.Items {
				logger.L().Debug("处理商品", zap.Int("sort", idx), zap.String("goods", item.GoodsName))

				// 处理规格
				specID, err := s.getOrCreateSpec(tx, item.SpecName)
				if err != nil {
					logger.L().Error("处理规格失败", zap.String("spec", item.SpecName), zap.Error(err))
					return fmt.Errorf("处理规格 %s 失败: %w", item.SpecName, err)
				}

				// 处理单位
				unitID, err := s.getOrCreateUnit(tx, item.UnitName)
				if err != nil {
					logger.L().Error("处理单位失败", zap.String("unit", item.UnitName), zap.Error(err))
					return fmt.Errorf("处理单位 %s 失败: %w", item.UnitName, err)
				}

				// 处理商品
				goodsID, err := s.getOrCreateGoods(tx, item.GoodsName, categoryID, specID, unitID, orgID)
				if err != nil {
					logger.L().Error("处理商品失败", zap.String("goods", item.GoodsName), zap.Error(err))
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
					logger.L().Error("创建询价商品明细失败", zap.Error(err), zap.Any("item", inquiryItem))
					return fmt.Errorf("创建询价商品明细失败: %w", err)
				}
				logger.L().Debug("明细创建成功", zap.String("item_id", inquiryItem.ID), zap.String("goods_id", goodsID))

				// 创建市场报价
				for marketName, price := range item.MarketPrices {
					if price == nil {
						logger.L().Debug("跳过空市场报价", zap.String("goods", item.GoodsName), zap.String("market", marketName))
						continue
					}
					marketID := marketIDMap[marketName]
					marketInquiry := &inquiryDomain.PriceMarketInquiry{
						InquiryID:      inquiry.ID,
						ItemID:         inquiryItem.ID,
						MarketID:       &marketID,
						MarketNameSnap: marketName,
						Price:          price,
					}
					if err := tx.Create(marketInquiry).Error; err != nil {
						logger.L().Error("创建市场报价失败",
							zap.Error(err),
							zap.String("item_id", inquiryItem.ID),
							zap.String("market", marketName),
							zap.Float64("price", safeDeref(price)))
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
						logger.L().Error("创建供应商结算失败",
							zap.Error(err),
							zap.String("item_id", inquiryItem.ID),
							zap.String("supplier", supplier.Name),
							zap.Float64("ratio", supplier.FloatRatio),
							zap.Float64("settlement", safeDeref(settlementPrice)))
						return fmt.Errorf("创建供应商结算失败: %w", err)
					}
					logger.L().Debug("供应商结算创建成功",
						zap.String("item_id", inquiryItem.ID),
						zap.String("supplier", supplier.Name),
						zap.Float64("ratio", supplier.FloatRatio),
						zap.Float64("settlement", safeDeref(settlementPrice)))
				}
			}
		}

		logger.L().Info("导入完成，准备提交事务", zap.String("inquiry_id", inquiry.ID))
		return nil
	})
}

// Helper functions

// extractDateFromTitle 从标题中提取日期
func extractDateFromTitle(title string) (time.Time, error) {
	// 匹配格式：2025年9月上旬 或 2025年9月中旬 或 2025年9月下旬
	re := regexp.MustCompile(`(\d{4})年(\d{1,2})月([上中下])旬`)
	matches := re.FindStringSubmatch(title)
	logger.L().Debug("标题日期匹配", zap.String("title", title), zap.Any("matches", matches))
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

	dt := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	logger.L().Info("返回的询价日期", zap.Time("inquiry_date", dt), zap.Int("year", year), zap.Int("month", month), zap.Int("day", day))
	return dt, nil
}

// parseHeader 解析表头，获取列索引、市场列表和供应商列表
func parseHeader(headerRow []string) (map[string]int, []string, []SupplierInfo, error) {
	logger.L().Debug("开始解析表头", zap.Strings("header_row_raw", headerRow))

	colMap := make(map[string]int)
	markets := []string{}
	suppliers := []SupplierInfo{}

	requiredCols := []string{"品名", "规格标准", "单位", "本期均价"}

	for idx, cell := range headerRow {
		cellRaw := cell
		cell = strings.TrimSpace(cell)
		cellNoSpace := utils.RemoveAllSpaces(cell)

		logger.L().Debug("表头单元格",
			zap.Int("index", idx),
			zap.String("raw", cellRaw),
			zap.String("trim", cell),
			zap.String("no_space", cellNoSpace))

		if cellNoSpace == "" {
			continue
		}

		// 检查必需列（用去空白的版本，容忍空格/NBSP）
		for _, required := range requiredCols {
			if strings.Contains(cellNoSpace, utils.RemoveAllSpaces(required)) {
				colMap[required] = idx
			}
		}

		// 检查上月均价
		if strings.Contains(cellNoSpace, utils.RemoveAllSpaces("上月均价")) || strings.Contains(cellNoSpace, utils.RemoveAllSpaces("上期均价")) {
			colMap["上月均价"] = idx
		}

		// 检查发改委指导价（目前仅记录列位，不参与后续运算）
		if strings.Contains(cellNoSpace, utils.RemoveAllSpaces("发改委")) && strings.Contains(cellNoSpace, utils.RemoveAllSpaces("指导价")) {
			colMap["发改委指导价"] = idx
		}

		// 识别询价市场（过滤关键词后，认为是市场）
		if !strings.Contains(cellNoSpace, "序号") &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("品名")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("规格")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("单位")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("发改委")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("上月")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("本期")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("结算")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("供应商")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("浮动")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("下浮")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("上浮")) &&
			!strings.Contains(cellNoSpace, utils.RemoveAllSpaces("均价")) {
			markets = append(markets, cell) // 保留原样显示
			colMap[cell] = idx
			logger.L().Debug("识别到市场列", zap.String("market", cell), zap.Int("col", idx))
		}

		// 识别供应商及浮动比例（如：胡坤本期结算价（下浮12%））
		if strings.Contains(cellNoSpace, utils.RemoveAllSpaces("结算价")) || strings.Contains(cellNoSpace, utils.RemoveAllSpaces("供应商")) {
			supplier, ratio, err := parseSupplierInfo(cell)
			if err == nil && supplier != "" {
				suppliers = append(suppliers, SupplierInfo{Name: supplier, FloatRatio: ratio})
				logger.L().Debug("识别到供应商列", zap.String("supplier", supplier), zap.Float64("ratio", ratio), zap.Int("col", idx))
			} else {
				logger.L().Warn("供应商列解析失败", zap.String("cell", cell), zap.Error(err))
			}
		}
	}

	// 检查必需列
	for _, required := range requiredCols {
		if _, ok := colMap[required]; !ok {
			return nil, nil, nil, fmt.Errorf("缺少必需列: %s", required)
		}
	}

	logger.L().Debug("表头解析完成", zap.Any("col_map", colMap), zap.Strings("markets", markets), zap.Any("suppliers", suppliers))
	return colMap, markets, suppliers, nil
}

// parseSupplierInfo 解析供应商信息
// 如："胡坤本期结算价（下浮12%）" -> ("胡坤", 0.88)
// 如："贵海本期结算价（下浮14%）" -> ("贵海", 0.86)
func parseSupplierInfo(cell string) (string, float64, error) {
	// 容忍空白差异
	noSpace := utils.RemoveAllSpaces(cell)
	// 匹配格式：XXX本期结算价（下浮/上浮 XX%）
	re := regexp.MustCompile(`(.+?)(?:本期)?结算价.*?([下上])浮\s*(\d+(?:\.\d+)?)%`)
	matches := re.FindStringSubmatch(noSpace)
	logger.L().Debug("供应商列匹配", zap.String("cell_no_space", noSpace), zap.Any("matches", matches))
	if len(matches) != 4 {
		return "", 1.0, fmt.Errorf("无法解析供应商信息")
	}

	name := strings.TrimSpace(matches[1])
	direction := matches[2] // "下" 或 "上"
	percent, _ := strconv.ParseFloat(matches[3], 64)

	var ratio float64
	if direction == "下" {
		ratio = (percent / 100.0)
	} else {
		ratio = 0 - (percent / 100.0)
	}
	logger.L().Debug("供应商解析结果", zap.String("name", name), zap.String("direction", direction), zap.Float64("percent", percent), zap.Float64("ratio", ratio))

	return name, ratio, nil
}

// getCellValue 安全获取单元格值
func getCellValue(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	val := strings.TrimSpace(row[idx])
	return val
}

// 统一判空/占位
func isBlankOrPlaceholder(s string) bool {
	if s == "" {
		return true
	}
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "-", "—", "N/A", "NA", "无", "无货", "空", "缺货", "缺", "/":
		return true
	}
	return false
}

// parseFloat 解析浮点数（兼容空白 / 占位 / 人民币符号 / 千分位）
func parseFloat(s string) *float64 {
	raw := s
	s = strings.TrimSpace(s)
	// 先清理各种空白（包含 NBSP/全角空格）
	s = utils.RemoveAllSpaces(s)
	if isBlankOrPlaceholder(s) {
		logger.L().Debug("检测到占位符/空值，返回 nil",
			zap.String("raw", raw),
			zap.String("normalized", s))
		return nil
	}
	// 去除常见的非数值字符（人民币/中文单位/逗号等）
	replacer := strings.NewReplacer("￥", "", "¥", "", ",", "", "元", "", "块", "", "人民币", "")
	s = replacer.Replace(s)

	if isBlankOrPlaceholder(s) {
		logger.L().Debug("检测到占位符/空值，返回 nil",
			zap.String("raw", raw),
			zap.String("normalized", s))
		return nil
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.L().Warn("数值解析失败，按无值处理", zap.String("raw", raw), zap.String("normalized", s), zap.Error(err))
		return nil
	}
	return &val
}

// getOrCreateCategory 获取或创建品类
func (s *InquiryImportService) getOrCreateCategory(tx *gorm.DB, name string, orgID string) (string, error) {
	logger.L().Debug("获取或创建品类", zap.String("name", name), zap.String("org_id", orgID))
	var category categoryDomain.Category
	err := tx.Where("name = ? AND org_id = ? AND is_deleted = 0", name, orgID).First(&category).Error
	if err == nil {
		logger.L().Debug("命中已有品类", zap.String("category_id", category.ID))
		return category.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		logger.L().Error("查询品类失败", zap.Error(err))
		return "", err
	}

	// 创建新品类
	category = categoryDomain.Category{
		Name:  name,
		OrgID: orgID,
	}
	if err := tx.Create(&category).Error; err != nil {
		logger.L().Error("创建品类失败", zap.Error(err), zap.Any("category", category))
		return "", err
	}
	logger.L().Info("创建品类成功", zap.String("category_id", category.ID), zap.String("name", name))
	return category.ID, nil
}

// getOrCreateSpec 获取或创建规格
func (s *InquiryImportService) getOrCreateSpec(tx *gorm.DB, name string) (string, error) {
	if name == "" {
		name = "默认"
	}
	logger.L().Debug("获取或创建规格", zap.String("name", name))
	var spec dictDomain.Spec
	err := tx.Where("name = ? AND is_deleted = 0", name).First(&spec).Error
	if err == nil {
		logger.L().Debug("命中已有规格", zap.String("spec_id", spec.ID))
		return spec.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		logger.L().Error("查询规格失败", zap.Error(err))
		return "", err
	}

	// 创建新规格
	spec = dictDomain.Spec{
		Name: name,
	}
	if err := tx.Create(&spec).Error; err != nil {
		logger.L().Error("创建规格失败", zap.Error(err), zap.Any("spec", spec))
		return "", err
	}
	logger.L().Info("创建规格成功", zap.String("spec_id", spec.ID), zap.String("name", name))
	return spec.ID, nil
}

// getOrCreateUnit 获取或创建单位
func (s *InquiryImportService) getOrCreateUnit(tx *gorm.DB, name string) (string, error) {
	if name == "" {
		name = "个"
	}
	logger.L().Debug("获取或创建单位", zap.String("name", name))
	var unit dictDomain.Unit
	err := tx.Where("name = ? AND is_deleted = 0", name).First(&unit).Error
	if err == nil {
		logger.L().Debug("命中已有单位", zap.String("unit_id", unit.ID))
		return unit.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		logger.L().Error("查询单位失败", zap.Error(err))
		return "", err
	}

	// 创建新单位
	unit = dictDomain.Unit{
		Name: name,
	}
	if err := tx.Create(&unit).Error; err != nil {
		logger.L().Error("创建单位失败", zap.Error(err), zap.Any("unit", unit))
		return "", err
	}
	logger.L().Info("创建单位成功", zap.String("unit_id", unit.ID), zap.String("name", name))
	return unit.ID, nil
}

// getOrCreateGoods 获取或创建商品
func (s *InquiryImportService) getOrCreateGoods(tx *gorm.DB, name string, categoryID, specID, unitID, orgID string) (string, error) {
	logger.L().Debug("获取或创建商品",
		zap.String("name", name),
		zap.String("category_id", categoryID),
		zap.String("spec_id", specID),
		zap.String("unit_id", unitID),
		zap.String("org_id", orgID))
	var goods goodsDomain.Goods
	err := tx.Where("name = ? AND category_id = ? AND spec_id = ? AND unit_id = ? AND org_id = ? AND is_deleted = 0",
		name, categoryID, specID, unitID, orgID).First(&goods).Error
	if err == nil {
		logger.L().Debug("命中已有商品", zap.String("goods_id", goods.ID))
		return goods.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		logger.L().Error("查询商品失败", zap.Error(err))
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
		logger.L().Error("创建商品失败", zap.Error(err), zap.Any("goods", goods))
		return "", err
	}
	logger.L().Info("创建商品成功", zap.String("goods_id", goods.ID), zap.String("name", name))
	return goods.ID, nil
}

// getOrCreateMarket 获取或创建市场
func (s *InquiryImportService) getOrCreateMarket(tx *gorm.DB, name string, orgID string) (string, error) {
	logger.L().Debug("获取或创建市场", zap.String("name", name), zap.String("org_id", orgID))
	var market inquiryDomain.BaseMarket
	err := tx.Where("name = ? AND org_id = ? AND is_deleted = 0", name, orgID).First(&market).Error
	if err == nil {
		logger.L().Debug("命中已有市场", zap.String("market_id", market.ID))
		return market.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		logger.L().Error("查询市场失败", zap.Error(err))
		return "", err
	}

	// 创建新市场
	market = inquiryDomain.BaseMarket{
		Name:  name,
		OrgID: orgID,
	}
	if err := tx.Create(&market).Error; err != nil {
		logger.L().Error("创建市场失败", zap.Error(err), zap.Any("market", market))
		return "", err
	}
	logger.L().Info("创建市场成功", zap.String("market_id", market.ID), zap.String("name", name))
	return market.ID, nil
}

// getOrCreateSupplier 获取或创建供应商，如果存在但浮动比例不同则更新
func (s *InquiryImportService) getOrCreateSupplier(tx *gorm.DB, name string, floatRatio float64, orgID string) (string, error) {
	logger.L().Debug("获取或创建供应商", zap.String("name", name), zap.Float64("float_ratio", floatRatio), zap.String("org_id", orgID))
	var supplier supplierDomain.Supplier
	err := tx.Where("name = ? AND org_id = ? AND is_deleted = 0", name, orgID).First(&supplier).Error

	if err == nil {
		if supplier.FloatRatio != floatRatio {
			logger.L().Info("更新供应商浮动比例", zap.String("supplier_id", supplier.ID),
				zap.Float64("old_ratio", supplier.FloatRatio), zap.Float64("new_ratio", floatRatio))
			if err := tx.Model(&supplier).Update("float_ratio", floatRatio).Error; err != nil {
				logger.L().Error("更新供应商比例失败", zap.Error(err))
				return "", err
			}
		} else {
			logger.L().Debug("供应商比例无需更新", zap.String("supplier_id", supplier.ID), zap.Float64("ratio", supplier.FloatRatio))
		}
		return supplier.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		logger.L().Error("查询供应商失败", zap.Error(err))
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
		logger.L().Error("创建供应商失败", zap.Error(err), zap.Any("supplier", supplier))
		return "", err
	}

	logger.L().Info("创建供应商成功", zap.String("supplier_id", supplier.ID), zap.String("name", name), zap.Float64("ratio", floatRatio))
	return supplier.ID, nil
}
