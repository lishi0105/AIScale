package inquiry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	categorydomain "hdzk.cn/foodapp/internal/domain/category"
	dictdomain "hdzk.cn/foodapp/internal/domain/dict"
	goodsdomain "hdzk.cn/foodapp/internal/domain/goods"
	inquirydomain "hdzk.cn/foodapp/internal/domain/inquiry"
	supplierdomain "hdzk.cn/foodapp/internal/domain/supplier"

	"gorm.io/gorm"
)

type ImportResult struct {
	SheetName        string `json:"sheet_name"`
	Title            string `json:"title"`
	InquiryID        string `json:"inquiry_id"`
	InquiryDate      string `json:"inquiry_date"`
	CategoryID       string `json:"category_id"`
	CategoryName     string `json:"category_name"`
	GoodsCreated     int    `json:"goods_created"`
	GoodsMatched     int    `json:"goods_matched"`
	SuppliersCreated int    `json:"suppliers_created"`
	SuppliersUpdated int    `json:"suppliers_updated"`
	Rows             int    `json:"rows"`
}

type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

func (s *Service) ImportExcel(ctx context.Context, orgID string, data []byte) ([]ImportResult, error) {
	trimmedOrg := strings.TrimSpace(orgID)
	if trimmedOrg == "" {
		return nil, &ValidationError{Msg: "org_id 不能为空"}
	}
	if len(data) == 0 {
		return nil, &ValidationError{Msg: "上传文件为空"}
	}

	sheets, err := readXLSXSheets(data)
	if err != nil {
		return nil, fmt.Errorf("解析 Excel 失败: %w", err)
	}
	if len(sheets) == 0 {
		return nil, &ValidationError{Msg: "Excel 中没有任何表格"}
	}

	var results []ImportResult
	for _, sheet := range sheets {
		if strings.TrimSpace(sheet.Name) == "" {
			continue
		}
		res, err := s.importSheet(ctx, sheet, trimmedOrg)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	if len(results) == 0 {
		return nil, &ValidationError{Msg: "Excel 中没有有效的表格"}
	}
	return results, nil
}

type supplierColumn struct {
	Index  int
	Name   string
	Ratio  float64
	Header string
}

type supplierPrice struct {
	Name  string
	Ratio float64
	Price *float64
}

type parsedRow struct {
	GoodsName      string
	SpecName       string
	UnitName       string
	AvgPrice       *float64
	MarketPrices   [3]*float64
	SupplierPrices []supplierPrice
}

var (
	dateRegexp      = regexp.MustCompile(`(\d{4})年(\d{1,2})月(?:(上旬|中旬|下旬)|(\d{1,2})日?)`)
	categoryRegexp  = regexp.MustCompile(`主要([^市场]+)市场`)
	percentRegexp   = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)`)
	requiredHeaders = []string{"品名", "规格标准", "单位", "本期均价"}
	requiredMarkets = []string{"富万家超市", "育英巷菜市场", "大润发"}
)

func (s *Service) importSheet(ctx context.Context, sheet xlsxSheet, orgID string) (ImportResult, error) {
	sheetName := sheet.Name
	rows := sheet.Rows
	if len(rows) == 0 {
		return ImportResult{}, fmt.Errorf("sheet %s: %w", sheetName, &ValidationError{Msg: "表格为空"})
	}

	title := ""
	maxScan := len(rows)
	if maxScan > 3 {
		maxScan = 3
	}
	for i := 0; i < maxScan; i++ {
		if len(rows[i]) == 0 {
			continue
		}
		if v := strings.TrimSpace(rows[i][0]); v != "" {
			title = v
			break
		}
	}
	if title == "" {
		return ImportResult{}, fmt.Errorf("sheet %s: %w", sheetName, &ValidationError{Msg: "未找到标题"})
	}
	inquiryDate, err := parseDateFromTitle(title)
	if err != nil {
		return ImportResult{}, fmt.Errorf("sheet %s: %w", sheetName, err)
	}

	headerIdx, headerRow, headerMap, err := findHeaderRow(rows)
	if err != nil {
		return ImportResult{}, fmt.Errorf("sheet %s: %w", sheetName, err)
	}

	nameIdx := headerMap["品名"]
	specIdx := headerMap["规格标准"]
	unitIdx := headerMap["单位"]
	avgIdx := headerMap["本期均价"]

	var marketIndexes [3]int
	for i, market := range requiredMarkets {
		idx, ok := headerMap[market]
		if !ok {
			return ImportResult{}, fmt.Errorf("sheet %s: %w", sheetName, &ValidationError{Msg: fmt.Sprintf("缺少询价列 %s", market)})
		}
		marketIndexes[i] = idx
	}

	maxColumns := longestRowLength(rows)
	var supplierCols []supplierColumn
	for col := avgIdx + 1; col < maxColumns; col++ {
		header := strings.TrimSpace(getCell(headerRow, col))
		if header == "" {
			continue
		}
		name, ratio, err := parseSupplierHeader(header)
		if err != nil {
			return ImportResult{}, fmt.Errorf("sheet %s 列[%s]: %w", sheetName, header, err)
		}
		supplierCols = append(supplierCols, supplierColumn{Index: col, Name: name, Ratio: ratio, Header: header})
	}
	if len(supplierCols) == 0 {
		return ImportResult{}, fmt.Errorf("sheet %s: %w", sheetName, &ValidationError{Msg: "未找到供应商列"})
	}

	var rowsData []parsedRow
	for rIdx := headerIdx + 1; rIdx < len(rows); rIdx++ {
		row := rows[rIdx]
		goodsName := strings.TrimSpace(getCell(row, nameIdx))
		if goodsName == "" {
			continue
		}
		specName := strings.TrimSpace(getCell(row, specIdx))
		if specName == "" {
			return ImportResult{}, fmt.Errorf("sheet %s 第%d行: %w", sheetName, rIdx+1, &ValidationError{Msg: "规格标准不能为空"})
		}
		unitName := strings.TrimSpace(getCell(row, unitIdx))
		if unitName == "" {
			return ImportResult{}, fmt.Errorf("sheet %s 第%d行: %w", sheetName, rIdx+1, &ValidationError{Msg: "单位不能为空"})
		}

		avgPrice, err := parseRequiredPrice(getCell(row, avgIdx), "本期均价")
		if err != nil {
			return ImportResult{}, fmt.Errorf("sheet %s 第%d行: %w", sheetName, rIdx+1, err)
		}

		var marketPrices [3]*float64
		for i, colIdx := range marketIndexes {
			price, err := parseRequiredPrice(getCell(row, colIdx), requiredMarkets[i])
			if err != nil {
				return ImportResult{}, fmt.Errorf("sheet %s 第%d行: %w", sheetName, rIdx+1, err)
			}
			marketPrices[i] = price
		}

		seenSuppliers := map[string]struct{}{}
		var supplierPrices []supplierPrice
		for _, col := range supplierCols {
			price, err := parsePrice(getCell(row, col.Index))
			if err != nil {
				return ImportResult{}, fmt.Errorf("sheet %s 第%d行 列[%s]: %w", sheetName, rIdx+1, col.Header, err)
			}
			if price == nil {
				continue
			}
			if _, exists := seenSuppliers[col.Name]; exists {
				return ImportResult{}, fmt.Errorf("sheet %s 第%d行: %w", sheetName, rIdx+1, &ValidationError{Msg: fmt.Sprintf("供应商 %s 重复", col.Name)})
			}
			seenSuppliers[col.Name] = struct{}{}
			supplierPrices = append(supplierPrices, supplierPrice{Name: col.Name, Ratio: col.Ratio, Price: price})
		}
		if len(supplierPrices) == 0 {
			return ImportResult{}, fmt.Errorf("sheet %s 第%d行: %w", sheetName, rIdx+1, &ValidationError{Msg: "至少需要一个供应商报价"})
		}

		rowsData = append(rowsData, parsedRow{
			GoodsName:      goodsName,
			SpecName:       specName,
			UnitName:       unitName,
			AvgPrice:       avgPrice,
			MarketPrices:   marketPrices,
			SupplierPrices: supplierPrices,
		})
	}
	if len(rowsData) == 0 {
		return ImportResult{}, fmt.Errorf("sheet %s: %w", sheetName, &ValidationError{Msg: "未找到有效的数据行"})
	}

	categoryName := strings.TrimSpace(sheetName)
	if categoryName == "" {
		categoryName = extractCategoryFromTitle(title)
		if categoryName == "" {
			return ImportResult{}, fmt.Errorf("sheet %s: %w", sheetName, &ValidationError{Msg: "无法确定品类名称"})
		}
	}

	result := ImportResult{
		SheetName:    sheetName,
		Title:        title,
		InquiryDate:  inquiryDate.Format("2006-01-02"),
		CategoryName: categoryName,
		Rows:         len(rowsData),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		category, _, err := ensureCategory(ctx, tx, orgID, categoryName)
		if err != nil {
			return err
		}
		result.CategoryID = category.ID

		inquiry := &inquirydomain.PriceInquiry{
			InquiryTitle: title,
			InquiryDate:  inquiryDate,
			OrgID:        orgID,
		}
		inquiry.Market1 = strPtr(requiredMarkets[0])
		inquiry.Market2 = strPtr(requiredMarkets[1])
		inquiry.Market3 = strPtr(requiredMarkets[2])

		var existing inquirydomain.PriceInquiry
		err = tx.WithContext(ctx).Where("org_id = ? AND inquiry_title = ? AND inquiry_date = ? AND is_deleted = 0", orgID, title, inquiryDate).First(&existing).Error
		if err == nil {
			return &ValidationError{Msg: fmt.Sprintf("询价记录已存在：%s", title)}
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := tx.WithContext(ctx).Create(inquiry).Error; err != nil {
			return fmt.Errorf("创建询价记录失败: %w", err)
		}
		result.InquiryID = inquiry.ID

		for _, row := range rowsData {
			spec, _, err := ensureSpec(ctx, tx, row.SpecName)
			if err != nil {
				return err
			}
			unit, _, err := ensureUnit(ctx, tx, row.UnitName)
			if err != nil {
				return err
			}

			goods, created, err := ensureGoods(ctx, tx, orgID, row.GoodsName, category.ID, spec.ID, unit.ID)
			if err != nil {
				return err
			}
			if created {
				result.GoodsCreated++
			} else {
				result.GoodsMatched++
			}

			detail := &inquirydomain.GoodsAvgDetail{
				GoodsID:      goods.ID,
				GuidePrice:   row.AvgPrice,
				Market1Price: row.MarketPrices[0],
				Market2Price: row.MarketPrices[1],
				Market3Price: row.MarketPrices[2],
				InquiryID:    inquiry.ID,
				OrgID:        orgID,
			}
			if err := tx.WithContext(ctx).Create(detail).Error; err != nil {
				return fmt.Errorf("保存均价明细失败（%s）: %w", row.GoodsName, err)
			}

			for _, sup := range row.SupplierPrices {
				supplier, createdSupplier, updatedSupplier, err := ensureSupplier(ctx, tx, orgID, sup.Name, sup.Ratio)
				if err != nil {
					return err
				}
				if createdSupplier {
					result.SuppliersCreated++
				}
				if updatedSupplier {
					result.SuppliersUpdated++
				}

				price := &inquirydomain.GoodsPrice{
					GoodsID:    goods.ID,
					SupplierID: supplier.ID,
					InquiryID:  inquiry.ID,
					UnitPrice:  *sup.Price,
					FloatRatio: sup.Ratio,
					OrgID:      orgID,
				}
				if err := tx.WithContext(ctx).Create(price).Error; err != nil {
					return fmt.Errorf("保存供应商报价失败（%s-%s）: %w", row.GoodsName, sup.Name, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return ImportResult{}, fmt.Errorf("sheet %s: %w", sheetName, err)
	}

	return result, nil
}

func findHeaderRow(rows [][]string) (int, []string, map[string]int, error) {
	for idx, row := range rows {
		headerMap := make(map[string]int)
		for colIdx, cell := range row {
			val := normalizeHeader(cell)
			if val != "" {
				headerMap[val] = colIdx
			}
		}
		if containsAll(headerMap, requiredHeaders) {
			return idx, row, headerMap, nil
		}
	}
	return 0, nil, nil, &ValidationError{Msg: "未找到包含必要列（品名、规格标准、单位、本期均价）的表头"}
}

func containsAll(m map[string]int, keys []string) bool {
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}

func normalizeHeader(v string) string {
	return strings.TrimSpace(strings.ReplaceAll(v, "\n", ""))
}

func longestRowLength(rows [][]string) int {
	max := 0
	for _, row := range rows {
		if len(row) > max {
			max = len(row)
		}
	}
	return max
}

func getCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func parsePrice(raw string) (*float64, error) {
	trimmed := strings.TrimSpace(strings.ReplaceAll(raw, ",", ""))
	if trimmed == "" || trimmed == "-" || trimmed == "—" || trimmed == "--" {
		return nil, nil
	}
	trimmed = strings.TrimSuffix(trimmed, "元")
	trimmed = strings.TrimSpace(trimmed)
	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return nil, &ValidationError{Msg: fmt.Sprintf("数值格式不正确：%s", raw)}
	}
	val := math.Round(v*100) / 100
	return &val, nil
}

func parseRequiredPrice(raw string, field string) (*float64, error) {
	v, err := parsePrice(raw)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, &ValidationError{Msg: fmt.Sprintf("%s 不能为空", field)}
	}
	return v, nil
}

func parseSupplierHeader(header string) (string, float64, error) {
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return "", 0, &ValidationError{Msg: "供应商列标题为空"}
	}
	openIdx := strings.IndexAny(trimmed, "(（")
	closeIdx := strings.LastIndexAny(trimmed, ")）")
	if openIdx == -1 || closeIdx == -1 || closeIdx <= openIdx {
		return "", 0, &ValidationError{Msg: fmt.Sprintf("供应商列缺少下浮比例：%s", header)}
	}
	name := strings.TrimSpace(trimmed[:openIdx])
	if name == "" {
		return "", 0, &ValidationError{Msg: fmt.Sprintf("供应商名称为空：%s", header)}
	}
	meta := strings.TrimSpace(trimmed[openIdx+1 : closeIdx])
	percentMatch := percentRegexp.FindString(meta)
	if percentMatch == "" {
		return "", 0, &ValidationError{Msg: fmt.Sprintf("供应商列未找到百分比：%s", header)}
	}
	pct, err := strconv.ParseFloat(percentMatch, 64)
	if err != nil {
		return "", 0, &ValidationError{Msg: fmt.Sprintf("供应商比例解析失败：%s", header)}
	}
	ratio := 1.0
	if strings.Contains(meta, "下浮") {
		ratio = 1 - pct/100
	} else if strings.Contains(meta, "上浮") {
		ratio = 1 + pct/100
	} else {
		ratio = pct / 100
	}
	if ratio <= 0 {
		return "", 0, &ValidationError{Msg: fmt.Sprintf("供应商浮动比例必须大于0：%s", header)}
	}
	ratio = math.Round(ratio*10000) / 10000
	return name, ratio, nil
}

func parseDateFromTitle(title string) (time.Time, error) {
	match := dateRegexp.FindStringSubmatch(title)
	if len(match) == 0 {
		return time.Time{}, &ValidationError{Msg: "标题中未找到日期"}
	}
	year, _ := strconv.Atoi(match[1])
	month, _ := strconv.Atoi(match[2])
	day := 1
	switch match[3] {
	case "上旬":
		day = 1
	case "中旬":
		day = 11
	case "下旬":
		day = 21
	default:
		if match[4] != "" {
			d, err := strconv.Atoi(match[4])
			if err == nil && d >= 1 && d <= 31 {
				day = d
			}
		}
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), nil
}

func extractCategoryFromTitle(title string) string {
	match := categoryRegexp.FindStringSubmatch(title)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func strPtr(v string) *string {
	if v == "" {
		return nil
	}
	vv := v
	return &vv
}

func ensureCategory(ctx context.Context, tx *gorm.DB, orgID string, name string) (*categorydomain.Category, bool, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, false, &ValidationError{Msg: "品类名称不能为空"}
	}
	var model categorydomain.Category
	err := tx.WithContext(ctx).Where("org_id = ? AND name = ? AND is_deleted = 0", orgID, trimmed).First(&model).Error
	if err == nil {
		return &model, false, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	model = categorydomain.Category{Name: trimmed, OrgID: orgID}
	if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, false, err
	}
	return &model, true, nil
}

func ensureSpec(ctx context.Context, tx *gorm.DB, name string) (*dictdomain.Spec, bool, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, false, &ValidationError{Msg: "规格标准不能为空"}
	}
	var model dictdomain.Spec
	err := tx.WithContext(ctx).Where("name = ? AND is_deleted = 0", trimmed).First(&model).Error
	if err == nil {
		return &model, false, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	model = dictdomain.Spec{Name: trimmed}
	if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, false, err
	}
	return &model, true, nil
}

func ensureUnit(ctx context.Context, tx *gorm.DB, name string) (*dictdomain.Unit, bool, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, false, &ValidationError{Msg: "单位不能为空"}
	}
	var model dictdomain.Unit
	err := tx.WithContext(ctx).Where("name = ? AND is_deleted = 0", trimmed).First(&model).Error
	if err == nil {
		return &model, false, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	model = dictdomain.Unit{Name: trimmed}
	if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, false, err
	}
	return &model, true, nil
}

func ensureGoods(ctx context.Context, tx *gorm.DB, orgID, name, categoryID, specID, unitID string) (*goodsdomain.Goods, bool, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, false, &ValidationError{Msg: "品名不能为空"}
	}
	var model goodsdomain.Goods
	err := tx.WithContext(ctx).
		Where("org_id = ? AND name = ? AND spec_id = ? AND unit_id = ? AND is_deleted = 0", orgID, trimmed, specID, unitID).
		First(&model).Error
	if err == nil {
		if model.CategoryID != categoryID {
			if err := tx.WithContext(ctx).Model(&goodsdomain.Goods{}).
				Where("id = ?", model.ID).
				Update("category_id", categoryID).Error; err != nil {
				return nil, false, err
			}
			model.CategoryID = categoryID
		}
		return &model, false, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	model = goodsdomain.Goods{
		Name:       trimmed,
		OrgID:      orgID,
		SpecID:     specID,
		UnitID:     unitID,
		CategoryID: categoryID,
	}
	if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, false, err
	}
	return &model, true, nil
}

func ensureSupplier(ctx context.Context, tx *gorm.DB, orgID, name string, ratio float64) (*supplierdomain.Supplier, bool, bool, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, false, false, &ValidationError{Msg: "供应商名称不能为空"}
	}
	var model supplierdomain.Supplier
	err := tx.WithContext(ctx).Where("org_id = ? AND name = ? AND is_deleted = 0", orgID, trimmed).First(&model).Error
	if err == nil {
		updated := false
		if math.Abs(model.FloatRatio-ratio) > 0.0001 {
			if err := tx.WithContext(ctx).Model(&supplierdomain.Supplier{}).
				Where("id = ?", model.ID).
				Update("float_ratio", ratio).Error; err != nil {
				return nil, false, false, err
			}
			model.FloatRatio = ratio
			updated = true
		}
		return &model, false, updated, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, false, err
	}
	desc := fmt.Sprintf("%s（自动导入）", trimmed)
	model = supplierdomain.Supplier{
		Name:        trimmed,
		OrgID:       orgID,
		Description: desc,
		FloatRatio:  ratio,
	}
	if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, false, false, err
	}
	return &model, true, false, nil
}
