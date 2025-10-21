package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	excelSvc "hdzk.cn/foodapp/internal/service/excel"
)

type ExcelHandler struct {
	s         *excelSvc.ExcelImportService
	uploadDir string
}

func NewExcelHandler(s *excelSvc.ExcelImportService, uploadDir string) *ExcelHandler {
	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("创建上传目录失败: %v", err))
	}
	return &ExcelHandler{
		s:         s,
		uploadDir: uploadDir,
	}
}

func (h *ExcelHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/excel")

	g.POST("/upload_chunk", h.uploadChunk)
	g.POST("/merge_chunks", h.mergeChunks)
	g.POST("/import", h.importExcel)
	g.POST("/validate", h.validateExcel)
}

// uploadChunk 上传文件切片
func (h *ExcelHandler) uploadChunk(c *gin.Context) {
	const errTitle = "上传文件切片失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可上传文件")
		return
	}

	// 获取参数
	filename := c.PostForm("filename")
	chunkIndexStr := c.PostForm("chunk_index")
	
	if filename == "" {
		BadRequest(c, errTitle, "缺少文件名")
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		BadRequest(c, errTitle, "切片索引格式错误")
		return
	}

	// 获取文件数据
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, errTitle, "获取文件失败: "+err.Error())
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		InternalError(c, errTitle, "打开文件失败: "+err.Error())
		return
	}
	defer src.Close()

	// 读取文件数据
	chunkData := make([]byte, file.Size)
	if _, err := src.Read(chunkData); err != nil {
		InternalError(c, errTitle, "读取文件数据失败: "+err.Error())
		return
	}

	// 保存切片
	if err := excelSvc.SaveUploadedChunk(h.uploadDir, filename, chunkIndex, chunkData); err != nil {
		InternalError(c, errTitle, "保存切片失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"chunk_index": chunkIndex,
		"message":     "切片上传成功",
	})
}

// mergeChunks 合并文件切片
func (h *ExcelHandler) mergeChunks(c *gin.Context) {
	const errTitle = "合并文件切片失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可操作")
		return
	}

	var req struct {
		Filename     string `json:"filename" binding:"required"`
		TotalChunks  int    `json:"total_chunks" binding:"required,min=1"`
		MD5          string `json:"md5" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法: "+err.Error())
		return
	}

	// 合并切片
	finalPath, err := excelSvc.MergeChunks(h.uploadDir, req.Filename, req.TotalChunks)
	if err != nil {
		InternalError(c, errTitle, "合并切片失败: "+err.Error())
		return
	}

	// 校验MD5
	if err := excelSvc.ValidateFile(finalPath, req.MD5); err != nil {
		// 删除文件
		os.Remove(finalPath)
		ConflictError(c, errTitle, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"filepath": finalPath,
		"message":  "文件合并成功",
	})
}

// validateExcel 校验Excel文件结构
func (h *ExcelHandler) validateExcel(c *gin.Context) {
	const errTitle = "校验Excel文件失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	var req struct {
		Filepath string `json:"filepath" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法: "+err.Error())
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(req.Filepath); os.IsNotExist(err) {
		NotFoundError(c, errTitle, "文件不存在")
		return
	}

	// 校验Excel结构
	excelData, err := h.s.ValidateExcelStructure(req.Filepath)
	if err != nil {
		if ve, ok := err.(*excelSvc.ValidationError); ok {
			BadRequest(c, errTitle, ve.Error())
		} else {
			InternalError(c, errTitle, err.Error())
		}
		return
	}

	// 返回校验结果摘要
	c.JSON(http.StatusOK, gin.H{
		"ok":    true,
		"title": excelData.Title,
		"date":  excelData.InquiryDate.Format("2006-01-02"),
		"stats": gin.H{
			"sheets":    len(excelData.Sheets),
			"markets":   len(excelData.Markets),
			"suppliers": len(excelData.Suppliers),
		},
		"message": "Excel文件校验通过",
	})
}

// importExcel 导入Excel数据
func (h *ExcelHandler) importExcel(c *gin.Context) {
	const errTitle = "导入Excel失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可导入数据")
		return
	}

	var req struct {
		Filepath string `json:"filepath" binding:"required"`
		OrgID    string `json:"org_id" binding:"required,uuid4"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法: "+err.Error())
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(req.Filepath); os.IsNotExist(err) {
		NotFoundError(c, errTitle, "文件不存在")
		return
	}

	// 校验Excel结构
	excelData, err := h.s.ValidateExcelStructure(req.Filepath)
	if err != nil {
		if ve, ok := err.(*excelSvc.ValidationError); ok {
			BadRequest(c, errTitle, ve.Error())
		} else {
			InternalError(c, errTitle, err.Error())
		}
		return
	}

	// 导入数据
	if err := h.s.ImportExcelData(c, excelData, req.OrgID); err != nil {
		InternalError(c, errTitle, "导入数据失败: "+err.Error())
		return
	}

	// 删除临时文件
	os.Remove(req.Filepath)
	// 清理临时目录
	tmpDir := filepath.Join(h.uploadDir, "tmp")
	os.RemoveAll(tmpDir)

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Excel数据导入成功",
		"stats": gin.H{
			"title":     excelData.Title,
			"sheets":    len(excelData.Sheets),
			"markets":   len(excelData.Markets),
			"suppliers": len(excelData.Suppliers),
		},
	})
}
