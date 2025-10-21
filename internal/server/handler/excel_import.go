package handler

import (
    "crypto/md5"
    "encoding/hex"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/xuri/excelize/v2"
    "gorm.io/gorm"

    middleware "hdzk.cn/foodapp/internal/server/middleware"
    catdom "hdzk.cn/foodapp/internal/domain/category"
    dictdom "hdzk.cn/foodapp/internal/domain/dict"
    goodsdom "hdzk.cn/foodapp/internal/domain/goods"
    marketdom "hdzk.cn/foodapp/internal/domain/market"
    supplierdom "hdzk.cn/foodapp/internal/domain/supplier"
)

// ExcelImportHandler 提供 Excel 分片上传、合并与导入数据库
// API:
// - POST /api/v1/excel/chunk  (multipart form: upload_id, index, md5, chunk)
// - POST /api/v1/excel/merge  (form: upload_id, total, filename, file_md5)
// - POST /api/v1/excel/import (json: { path, org_id, date })

type ExcelImportHandler struct{ db *gorm.DB }

func NewExcelImportHandler(db *gorm.DB) *ExcelImportHandler { return &ExcelImportHandler{db: db} }

func (h *ExcelImportHandler) Register(rg *gin.RouterGroup) {
    g := rg.Group("/excel")
    g.POST("/chunk", h.uploadChunk)
    g.POST("/merge", h.mergeChunks)
    g.POST("/import", h.importExcel)
}

// ========== 分片上传 ==========
// 表单字段：upload_id, index(int), md5(hex), chunk(file)
func (h *ExcelImportHandler) uploadChunk(c *gin.Context) {
    const errTitle = "上传分片失败"
    act := middleware.GetActor(c)
    if act.Deleted != middleware.DeletedNo {
        ForbiddenError(c, errTitle, "账户已停用，禁止操作")
        return
    }
    if act.Role != middleware.RoleAdmin {
        ForbiddenError(c, errTitle, "仅管理员可上传导入文件")
        return
    }
    uploadID := strings.TrimSpace(c.PostForm("upload_id"))
    index := c.PostForm("index")
    md5hex := strings.TrimSpace(c.PostForm("md5"))
    if uploadID == "" || index == "" || md5hex == "" {
        BadRequest(c, errTitle, "缺少 upload_id/index/md5")
        return
    }
    idx, err := strconv.Atoi(index)
    if err != nil || idx < 0 {
        BadRequest(c, errTitle, "index 非法")
        return
    }
    f, fh, err := c.Request.FormFile("chunk")
    if err != nil {
        BadRequest(c, errTitle, "读取文件失败: "+err.Error())
        return
    }
    defer f.Close()

    tmpDir := filepath.Join(os.TempDir(), "excel_chunks", uploadID)
    if err := os.MkdirAll(tmpDir, 0o755); err != nil {
        InternalError(c, errTitle, "创建临时目录失败: "+err.Error())
        return
    }
    // 计算 md5 并保存
    hsh := md5.New()
    tee := io.TeeReader(f, hsh)
    data, err := io.ReadAll(tee)
    if err != nil {
        InternalError(c, errTitle, "读取分片失败: "+err.Error())
        return
    }
    sum := hex.EncodeToString(hsh.Sum(nil))
    if !strings.EqualFold(sum, md5hex) {
        BadRequest(c, errTitle, fmt.Sprintf("MD5 不匹配，期望 %s 实际 %s", md5hex, sum))
        return
    }
    out := filepath.Join(tmpDir, fmt.Sprintf("%06d.part", idx))
    if err := os.WriteFile(out, data, 0o644); err != nil {
        InternalError(c, errTitle, "写入分片失败: "+err.Error())
        return
    }
    c.JSON(http.StatusOK, gin.H{"ok": true, "size": fh.Size})
}

// ========== 合并 ==========
// 表单：upload_id, total(int), filename, file_md5(hex)
func (h *ExcelImportHandler) mergeChunks(c *gin.Context) {
    const errTitle = "合并文件失败"
    act := middleware.GetActor(c)
    if act.Deleted != middleware.DeletedNo {
        ForbiddenError(c, errTitle, "账户已停用，禁止操作")
        return
    }
    if act.Role != middleware.RoleAdmin {
        ForbiddenError(c, errTitle, "仅管理员可合并导入文件")
        return
    }
    uploadID := strings.TrimSpace(c.PostForm("upload_id"))
    filename := strings.TrimSpace(c.PostForm("filename"))
    totalStr := c.PostForm("total")
    fileMD5 := strings.TrimSpace(c.PostForm("file_md5"))
    if uploadID == "" || filename == "" || totalStr == "" || fileMD5 == "" {
        BadRequest(c, errTitle, "缺少参数")
        return
    }
    total, err := strconv.Atoi(totalStr)
    if err != nil || total <= 0 {
        BadRequest(c, errTitle, "total 非法")
        return
    }
    tmpDir := filepath.Join(os.TempDir(), "excel_chunks", uploadID)
    entries, err := os.ReadDir(tmpDir)
    if err != nil {
        InternalError(c, errTitle, "读取临时目录失败: "+err.Error())
        return
    }
    if len(entries) < total {
        BadRequest(c, errTitle, "分片数量不足")
        return
    }
    // 排序
    names := make([]string, 0, len(entries))
    for _, e := range entries { names = append(names, e.Name()) }
    sort.Strings(names)

    // 合并
    mergedDir := filepath.Join(os.TempDir(), "excel_merged")
    _ = os.MkdirAll(mergedDir, 0o755)
    target := filepath.Join(mergedDir, uploadID+"_"+filepath.Base(filename))
    out, err := os.Create(target)
    if err != nil { InternalError(c, errTitle, "创建目标失败: "+err.Error()); return }
    defer out.Close()
    hsh := md5.New()
    for _, name := range names {
        part, err := os.ReadFile(filepath.Join(tmpDir, name))
        if err != nil { InternalError(c, errTitle, "读取分片失败: "+err.Error()); return }
        if _, err := out.Write(part); err != nil { InternalError(c, errTitle, "写入失败: "+err.Error()); return }
        _, _ = hsh.Write(part)
    }
    sum := hex.EncodeToString(hsh.Sum(nil))
    if !strings.EqualFold(sum, fileMD5) {
        BadRequest(c, errTitle, fmt.Sprintf("整文件 MD5 不匹配，期望 %s 实际 %s", fileMD5, sum))
        return
    }
    // 清理分片
    _ = os.RemoveAll(tmpDir)
    c.JSON(http.StatusOK, gin.H{"ok": true, "path": target})
}

// ========== 解析并入库 ==========
// JSON: { path, org_id, date(YYYY-MM-DD) }
func (h *ExcelImportHandler) importExcel(c *gin.Context) {
    act := middleware.GetActor(c)
    if act.Deleted != middleware.DeletedNo {
        ForbiddenError(c, "导入Excel失败", "账户已停用，禁止操作")
        return
    }
    if act.Role != middleware.RoleAdmin {
        ForbiddenError(c, "导入Excel失败", "仅管理员可导入")
        return
    }
    type reqBody struct {
        Path  string `json:"path" binding:"required"`
        OrgID string `json:"org_id" binding:"required,uuid4"`
        Date  string `json:"date" binding:"required"`
    }
    const errTitle = "导入Excel失败"
    var req reqBody
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, errTitle, "请求参数无效: "+err.Error())
        return
    }

    // 打开 excel
    xf, err := excelize.OpenFile(req.Path)
    if err != nil {
        BadRequest(c, errTitle, "无法打开Excel: "+err.Error())
        return
    }
    defer xf.Close()

    // 1) 标题校验：第一张 sheet 的 A1，包含“参考价”字样
    firstSheet := xf.GetSheetName(0)
    title, _ := xf.GetCellValue(firstSheet, "A1")
    if strings.TrimSpace(title) == "" || !strings.Contains(title, "参考价") {
        BadRequest(c, errTitle, "Excel 必须包含标题（A1，含‘参考价’）")
        return
    }

    // 解析日期
    date, err := time.Parse("2006-01-02", strings.TrimSpace(req.Date))
    if err != nil { BadRequest(c, errTitle, "date 格式应为 YYYY-MM-DD"); return }

    // 供应商列：如 “胡坤本期结算价（下浮12%）/（上浮10%）”
    supplierRe := regexp.MustCompile(`^(?s)(.+?)本期结算价（([上下])浮(\d+(?:\.\d+)?)%）$`)

    // 事务导入
    var inquiryID string
    if err := h.db.Transaction(func(tx *gorm.DB) error {
        // 创建询价单
        inq := &marketdom.BasePriceInquiry{ OrgID: req.OrgID, InquiryTitle: strings.TrimSpace(title), InquiryDate: date }
        if err := tx.Create(inq).Error; err != nil { return fmt.Errorf("创建询价单失败: %w", err) }
        inquiryID = inq.ID

        // name->id 本地缓存
        categoryIDs := map[string]string{}
        specIDs := map[string]string{}
        unitIDs := map[string]string{}
        goodsIDs := map[string]string{}
        marketIDs := map[string]string{}
        supplierIDs := map[string]string{}

        // helpers
        findOrCreateCategory := func(name string) (string, error) {
            name = strings.TrimSpace(name)
            if name == "" { return "", errors.New("品类名称为空") }
            if id, ok := categoryIDs[name]; ok { return id, nil }
            var m catdom.Category
            if err := tx.Where("is_deleted=0 AND org_id=? AND name=?", req.OrgID, name).First(&m).Error; err == nil {
                categoryIDs[name] = m.ID; return m.ID, nil
            }
            m = catdom.Category{Name: name, OrgID: req.OrgID}
            if err := tx.Create(&m).Error; err != nil { return "", err }
            categoryIDs[name] = m.ID
            return m.ID, nil
        }
        findOrCreateSpec := func(name string) (string, error) {
            name = strings.TrimSpace(name)
            if name == "" { return "", errors.New("规格标准为空") }
            if id, ok := specIDs[name]; ok { return id, nil }
            var m dictdom.Spec
            if err := tx.Where("is_deleted=0 AND name=?", name).First(&m).Error; err == nil {
                specIDs[name] = m.ID; return m.ID, nil
            }
            m = dictdom.Spec{Name: name}
            if err := tx.Create(&m).Error; err != nil { return "", err }
            specIDs[name] = m.ID
            return m.ID, nil
        }
        findOrCreateUnit := func(name string) (string, error) {
            name = strings.TrimSpace(name)
            if name == "" { return "", errors.New("单位为空") }
            if id, ok := unitIDs[name]; ok { return id, nil }
            var m dictdom.Unit
            if err := tx.Where("is_deleted=0 AND name=?", name).First(&m).Error; err == nil {
                unitIDs[name] = m.ID; return m.ID, nil
            }
            m = dictdom.Unit{Name: name}
            if err := tx.Create(&m).Error; err != nil { return "", err }
            unitIDs[name] = m.ID
            return m.ID, nil
        }
        goodsKey := func(name, specID, unitID string) string { return name+"|"+specID+"|"+unitID }
        findOrCreateGoods := func(name, categoryID, specID, unitID string) (string, error) {
            name = strings.TrimSpace(name)
            if name == "" { return "", errors.New("品名为空") }
            key := goodsKey(name, specID, unitID)
            if id, ok := goodsIDs[key]; ok { return id, nil }
            var m goodsdom.Goods
            if err := tx.Where("is_deleted=0 AND org_id=? AND name=? AND spec_id=? AND unit_id=?", req.OrgID, name, specID, unitID).First(&m).Error; err == nil {
                goodsIDs[key] = m.ID; return m.ID, nil
            }
            m = goodsdom.Goods{Name: name, OrgID: req.OrgID, CategoryID: categoryID, SpecID: specID, UnitID: unitID}
            if err := tx.Create(&m).Error; err != nil { return "", err }
            goodsIDs[key] = m.ID
            return m.ID, nil
        }
        findOrCreateMarket := func(name string) (string, error) {
            name = strings.TrimSpace(name)
            if name == "" { return "", errors.New("市场名称为空") }
            if id, ok := marketIDs[name]; ok { return id, nil }
            var m marketdom.BaseMarket
            if err := tx.Where("is_deleted=0 AND org_id=? AND name=?", req.OrgID, name).First(&m).Error; err == nil {
                marketIDs[name] = m.ID; return m.ID, nil
            }
            m = marketdom.BaseMarket{Name: name, OrgID: req.OrgID}
            if err := tx.Create(&m).Error; err != nil { return "", err }
            marketIDs[name] = m.ID
            return m.ID, nil
        }
        findOrCreateSupplier := func(name string, ratio float64) (string, float64, error) {
            name = strings.TrimSpace(name)
            if name == "" { return "", 0, errors.New("供应商名称为空") }
            if id, ok := supplierIDs[name]; ok { return id, ratio, nil }
            var m supplierdom.Supplier
            if err := tx.Where("is_deleted=0 AND org_id=? AND name=?", req.OrgID, name).First(&m).Error; err == nil {
                if ratio > 0 && m.FloatRatio != ratio {
                    if err := tx.Model(&supplierdom.Supplier{}).Where("id=?", m.ID).Update("float_ratio", ratio).Error; err != nil { return "", 0, err }
                }
                supplierIDs[name] = m.ID
                return m.ID, ratio, nil
            }
            // 新建
            m = supplierdom.Supplier{Name: name, OrgID: req.OrgID, Description: "", FloatRatio: ratio}
            if err := tx.Create(&m).Error; err != nil { return "", 0, err }
            supplierIDs[name] = m.ID
            return m.ID, ratio, nil
        }

        // 遍历所有 sheet
        for _, sheet := range xf.GetSheetList() {
            rows, err := xf.GetRows(sheet)
            if err != nil { return fmt.Errorf("读取 sheet %s 失败: %w", sheet, err) }
            if len(rows) == 0 { continue }

            // 寻找表头行：含 品名/单位/本期均价
            headerIdx, header := findHeader(rows)
            if headerIdx < 0 {
                return fmt.Errorf("sheet %s 缺少必需列：品名/规格标准/单位/本期均价", sheet)
            }

            // 列索引
            idxName := indexOf(header, "品名")
            idxSpec := indexOf(header, "规格标准")
            idxUnit := indexOf(header, "单位")
            idxLast := indexOf(header, "上月均价")
            idxCurr := indexOf(header, "本期均价")

            // 识别市场列 & 供应商列
            marketCols := map[int]string{}
            type supCol struct{ idx int; name string; ratio float64 }
            var supCols []supCol

            for i, raw := range header {
                name := strings.TrimSpace(raw)
                if name == "" { continue }
                if i == idxName || i == idxSpec || i == idxUnit || i == idxLast || i == idxCurr { continue }
                switch name { // 排除常见非市场字段
                case "序号", "发改委指导价", "上月均价", "本期均价", "品名", "规格标准", "单位":
                    continue
                }
                // 供应商列
                if m := supplierRe.FindStringSubmatch(name); len(m) == 4 {
                    pct, _ := strconv.ParseFloat(m[3], 64)
                    ratio := 1.0
                    if m[2] == "下" {
                        ratio = 1.0 - pct/100.0
                    } else {
                        ratio = 1.0 + pct/100.0
                    }
                    supCols = append(supCols, supCol{idx: i, name: m[1], ratio: ratio})
                    continue
                }
                // 其他未识别列视为市场列
                marketCols[i] = name
            }
            if len(marketCols) == 0 {
                return fmt.Errorf("sheet %s 缺少询价市场列", sheet)
            }
            if len(supCols) == 0 {
                return fmt.Errorf("sheet %s 需至少1个供应商列(带上下浮比例)", sheet)
            }

            // 确保品类（用 sheet 名）
            categoryID, err := findOrCreateCategory(sheet)
            if err != nil { return fmt.Errorf("创建/获取品类失败[%s]: %w", sheet, err) }

            // 逐行数据
            for r := headerIdx + 1; r < len(rows); r++ {
                row := rows[r]
                nm := getCell(row, idxName)
                if nm == "" { continue }
                specName := getCell(row, idxSpec)
                unitName := getCell(row, idxUnit)
                var lastPtr, currPtr *float64
                if idxLast >= 0 { if v := parseFloat(getCell(row, idxLast)); v != nil { lastPtr = v } }
                if idxCurr >= 0 { if v := parseFloat(getCell(row, idxCurr)); v != nil { currPtr = v } }

                specID, err := findOrCreateSpec(specName)
                if err != nil { return fmt.Errorf("创建/获取规格失败[%s]: %w", specName, err) }
                unitID, err := findOrCreateUnit(unitName)
                if err != nil { return fmt.Errorf("创建/获取单位失败[%s]: %w", unitName, err) }
                goodsID, err := findOrCreateGoods(nm, categoryID, specID, unitID)
                if err != nil { return fmt.Errorf("创建/获取商品失败[%s]: %w", nm, err) }

                item := &marketdom.PriceInquiryItem{
                    InquiryID:          inquiryID,
                    GoodsID:            goodsID,
                    CategoryID:         categoryID,
                    SpecID:             &specID,
                    UnitID:             &unitID,
                    GoodsNameSnap:      nm,
                    CategoryNameSnap:   sheet,
                    SpecNameSnap:       &specName,
                    UnitNameSnap:       &unitName,
                    LastMonthAvgPrice:  lastPtr,
                    CurrentAvgPrice:    currPtr,
                }
                if err := tx.Create(item).Error; err != nil { return fmt.Errorf("写入询价明细失败: %w", err) }

                // 市场报价
                for idx, marketName := range marketCols {
                    price := parseFloat(getCell(row, idx))
                    // 即使价格为空，也保留名称快照以体现询价范围
                    mid, err := findOrCreateMarket(marketName)
                    if err != nil { return fmt.Errorf("创建/获取市场失败[%s]: %w", marketName, err) }
                    mi := &marketdom.PriceMarketInquiry{ InquiryID: inquiryID, ItemID: item.ID, MarketID: &mid, MarketNameSnap: marketName, Price: price }
                    if err := tx.Create(mi).Error; err != nil { return fmt.Errorf("写市场报价失败: %w", err) }
                }
                // 供应商结算（仅保存比例与名称快照；结算价不存储）
                for _, sc := range supCols {
                    if _, _, err := findOrCreateSupplier(sc.name, sc.ratio); err != nil { return fmt.Errorf("创建/获取供应商失败[%s]: %w", sc.name, err) }
                    ss := &marketdom.PriceSupplierSettlement{ InquiryID: inquiryID, ItemID: item.ID, SupplierNameSnap: sc.name, FloatRatioSnap: sc.ratio }
                    if err := tx.Create(ss).Error; err != nil { return fmt.Errorf("写供应商结算失败: %w", err) }
                }
            }
        }
        return nil
    }); err != nil {
        InternalError(c, errTitle, err.Error())
        return
    }

    c.JSON(http.StatusOK, gin.H{"ok": true, "inquiry_id": inquiryID})
}

// findHeader 在若干行内查找包含“品名/规格标准/单位/本期均价”的表头行
func findHeader(rows [][]string) (int, []string) {
    for i := 0; i < len(rows) && i < 20; i++ {
        row := rows[i]
        if indexOf(row, "品名") >= 0 && indexOf(row, "规格标准") >= 0 && indexOf(row, "单位") >= 0 && indexOf(row, "本期均价") >= 0 {
            return i, row
        }
    }
    return -1, nil
}

func indexOf(row []string, name string) int {
    for i, v := range row { if strings.TrimSpace(v) == name { return i } }
    return -1
}

func getCell(row []string, idx int) string { if idx < 0 || idx >= len(row) { return "" }; return strings.TrimSpace(row[idx]) }

func parseFloat(s string) *float64 {
    s = strings.TrimSpace(s)
    if s == "" { return nil }
    s = strings.ReplaceAll(s, ",", "")
    if v, err := strconv.ParseFloat(s, 64); err == nil { return &v }
    return nil
}
