package inquiry

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/xuri/excelize/v2"

    inqdom "hdzk.cn/foodapp/internal/domain/inquiry"
    repo "hdzk.cn/foodapp/internal/repository/inquiry"
    "gorm.io/gorm"
)

// ImportService handles importing of inquiry records from Excel.
type ImportService struct {
    r repo.ImportRepository
}

func NewImportService(r repo.ImportRepository) *ImportService { return &ImportService{r: r} }

// ImportParams describes the upload request.
type ImportParams struct {
    OrgID string
    // File bytes of the uploaded Excel
    Data []byte
}

// Validation rules per user requirements
var (
    headerTitleRe   = regexp.MustCompile(`^\d{4}年\d{1,2}月[上下]旬?都匀市主要.*参考价$`)
    supplierRatioRe = regexp.MustCompile(`^(.+?)本期结算价（下浮([0-9]+(?:\.[0-9]+)?)%）$`)
)

// Import reads the Excel, validates layout, creates/updates base tables and inserts inquiry details.
func (s *ImportService) Import(ctx context.Context, p ImportParams) (string, error) {
    if strings.TrimSpace(p.OrgID) == "" {
        return "", fmt.Errorf("org_id 不能为空")
    }
    if len(p.Data) == 0 {
        return "", fmt.Errorf("空文件")
    }

    f, err := excelize.OpenReader(bytes.NewReader(p.Data))
    if err != nil {
        return "", fmt.Errorf("打开Excel失败: %w", err)
    }
    defer f.Close()

    // 1) title in the first sheet top-left cell(s)
    sheets := f.GetSheetList()
    if len(sheets) == 0 {
        return "", errors.New("Excel 无工作表")
    }

    firstSheet := sheets[0]
    // The title usually stays in A1; fallback to scan first 3 rows, 5 cols
    title := ""
    for r := 1; r <= 3 && title == ""; r++ {
        for c := 1; c <= 5 && title == ""; c++ {
            cell, _ := excelize.CoordinatesToCellName(c, r)
            v, _ := f.GetCellValue(firstSheet, cell)
            v = strings.TrimSpace(v)
            if v != "" {
                title = v
            }
        }
    }
    if title == "" || !headerTitleRe.MatchString(title) {
        return "", fmt.Errorf("Excel 标题缺失或不符合格式: %q", title)
    }

    // 2) category sheets: each sheet represents a category (can be one)
    // for each sheet, validate required columns and markets and suppliers

    // Collect known required logical column headers
    requiredCols := []string{"品名", "规格标准", "单位", "本期均价"}
    // Market columns: exactly 3
    marketHeaders := []string{"富万家超市", "育英巷菜市场", "大润发"}

    // Suppliers: dynamic, at least 1, headers include the ratio
    // We will parse with supplierRatioRe

    // Create inquiry header using derived title and date
    // Extract date from the title: "2025年9月上旬..." -> choose first day of the month
    inqDate, derr := parseDateFromTitle(title)
    if derr != nil {
        return "", derr
    }

    // For this import we create one PriceInquiry header for the whole file (all sheets)
    inq := &inqdom.PriceInquiry{
        ID:           uuid.NewString(),
        OrgID:        p.OrgID,
        InquiryTitle: title,
        InquiryDate:  inqDate,
        Market1:      &marketHeaders[0],
        Market2:      &marketHeaders[1],
        Market3:      &marketHeaders[2],
    }

    // Use a transaction
    gormdb := getDBFromRepo(s.r)
    tx := gormdb.WithContext(ctx).Begin()
    if err := s.r.CreateInquiry(ctx, tx, inq); err != nil {
        tx.Rollback()
        return "", err
    }

    for _, sheet := range sheets {
        if strings.TrimSpace(sheet) == "" { continue }
        // find header row: scan first 10 rows to find columns we need
        hdrRowIdx, hdrMap, err := locateHeader(f, sheet)
        if err != nil {
            tx.Rollback(); return "", fmt.Errorf("[%s] 表头识别失败: %w", sheet, err)
        }
        // validate required columns
        for _, name := range requiredCols {
            if _, ok := hdrMap[name]; !ok {
                tx.Rollback(); return "", fmt.Errorf("[%s] 缺少必填列: %s", sheet, name)
            }
        }
        // markets
        for _, m := range marketHeaders {
            if _, ok := hdrMap[m]; !ok {
                tx.Rollback(); return "", fmt.Errorf("[%s] 缺少市场列: %s", sheet, m)
            }
        }
        // supplier columns: capture all that match pattern
        supplierCols := map[string]struct{ idx int; ratio float64 }{}
        for name, idx := range hdrMap {
            m := supplierRatioRe.FindStringSubmatch(name)
            if len(m) == 3 {
                ratio, _ := strconv.ParseFloat(m[2], 64)
                supplierCols[m[1]] = struct{ idx int; ratio float64 }{idx: idx, ratio: ratio / 100.0}
            }
        }
        if len(supplierCols) == 0 {
            tx.Rollback(); return "", fmt.Errorf("[%s] 供应商列缺失（至少1个，且需包含下浮比例）", sheet)
        }

        // base category for this sheet
        cat, err := s.r.GetOrCreateCategory(ctx, tx, p.OrgID, sheet)
        if err != nil { tx.Rollback(); return "", err }

        // iterate rows until blank 品名
        for row := hdrRowIdx + 1; row <= hdrRowIdx+10000; row++ {
            name := getCellTrim(f, sheet, hdrMap["品名"], row)
            if name == "" { break }
            spec := getCellTrim(f, sheet, hdrMap["规格标准"], row)
            unit := getCellTrim(f, sheet, hdrMap["单位"], row)

            // upsert dicts and goods
            specM, err := s.r.GetOrCreateSpec(ctx, tx, spec)
            if err != nil { tx.Rollback(); return "", err }
            unitM, err := s.r.GetOrCreateUnit(ctx, tx, unit)
            if err != nil { tx.Rollback(); return "", err }
            goods, err := s.r.GetOrCreateGoods(ctx, tx, p.OrgID, name, cat.ID, specM.ID, unitM.ID)
            if err != nil { tx.Rollback(); return "", err }

            // markets prices
            m1 := parseFloatPtr(getCellTrim(f, sheet, hdrMap[marketHeaders[0]], row))
            m2 := parseFloatPtr(getCellTrim(f, sheet, hdrMap[marketHeaders[1]], row))
            m3 := parseFloatPtr(getCellTrim(f, sheet, hdrMap[marketHeaders[2]], row))
            guide := parseFloatPtr(getCellTrim(f, sheet, hdrMap["发改委指导价"], row))

            gad := &inqdom.GoodsAvgDetail{
                ID:           uuid.NewString(),
                GoodsID:      goods.ID,
                GuidePrice:   guide,
                Market1Price: m1,
                Market2Price: m2,
                Market3Price: m3,
                InquiryID:    inq.ID,
            }
            if err := s.r.CreateGoodsAvgDetail(ctx, tx, gad); err != nil { tx.Rollback(); return "", err }

            // supplier prices
            for sname, col := range supplierCols {
                price := parseFloat(getCellTrim(f, sheet, col.idx, row))
                if price <= 0 { continue }
                // supplier with ratio
                sup, err := s.r.GetOrCreateSupplierWithRatio(ctx, tx, p.OrgID, sname, col.ratio)
                if err != nil { tx.Rollback(); return "", err }
                orgID := p.OrgID
                gp := &inqdom.GoodsPrice{
                    ID:         uuid.NewString(),
                    GoodsID:    goods.ID,
                    SupplierID: sup.ID,
                    InquiryID:  inq.ID,
                    UnitPrice:  price,
                    FloatRatio: col.ratio,
                    OrgID:      &orgID,
                }
                if err := s.r.UpsertGoodsPrice(ctx, tx, gp); err != nil { tx.Rollback(); return "", err }
            }
        }
    }

    if err := tx.Commit().Error; err != nil { return "", err }
    return inq.ID, nil
}

func parseDateFromTitle(title string) (time.Time, error) {
    // very simple: extract YYYY年M月, use first day
    y, m := 0, 0
    var err error
    // find first occurrence of 年 and 月
    nIdx := strings.Index(title, "年")
    if nIdx > 0 {
        y, err = strconv.Atoi(strings.TrimSpace(title[:nIdx]))
        if err != nil { return time.Time{}, fmt.Errorf("无法解析年份") }
    }
    rest := title
    if nIdx >= 0 && nIdx+len("年") < len(title) {
        rest = title[nIdx+len("年"):]
    }
    mIdx := strings.Index(rest, "月")
    if mIdx > 0 {
        m, err = strconv.Atoi(strings.TrimSpace(rest[:mIdx]))
        if err != nil { return time.Time{}, fmt.Errorf("无法解析月份") }
    }
    if y == 0 || m == 0 { return time.Time{}, fmt.Errorf("标题缺少年月") }
    return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.Local), nil
}

func locateHeader(f *excelize.File, sheet string) (row int, hdr map[string]int, err error) {
    // scan first 20 rows for a row that contains most required markers
    hdr = map[string]int{}
    bestRow, bestCnt := 0, -1
    rows, _ := f.GetRows(sheet)
    limit := 20
    if len(rows) < limit { limit = len(rows) }
    for r := 0; r < limit; r++ {
        rowCells := rows[r]
        tmp := map[string]int{}
        for i, v := range rowCells {
            name := strings.TrimSpace(v)
            if name == "" { continue }
            tmp[name] = i + 1 // 1-based col index
        }
        // heuristic: must contain 品名
        if _, ok := tmp["品名"]; !ok { continue }
        cnt := len(tmp)
        if cnt > bestCnt {
            bestCnt = cnt
            bestRow = r + 1 // convert to 1-based row index
            hdr = tmp
        }
    }
    if bestRow == 0 { return 0, nil, fmt.Errorf("未找到表头") }
    return bestRow, hdr, nil
}

func getCellTrim(f *excelize.File, sheet string, col int, row int) string {
    if col <= 0 || row <= 0 { return "" }
    cell, _ := excelize.CoordinatesToCellName(col, row)
    v, _ := f.GetCellValue(sheet, cell)
    return strings.TrimSpace(v)
}

func parseFloatPtr(s string) *float64 {
    if s == "" { return nil }
    f, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
    if err != nil { return nil }
    return &f
}

func parseFloat(s string) float64 {
    f, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(s), ",", ""), 64)
    return f
}

// getDBFromRepo uses an internal knowledge that repoImport holds *gorm.DB.
// It allows us to start a transaction without adding a new method on interface for now.
func getDBFromRepo(r repo.ImportRepository) *gorm.DB {
    type hasDB interface{ DB() *gorm.DB }
    if v, ok := any(r).(hasDB); ok { return v.DB() }
    panic("unsupported ImportRepository implementation")
}
