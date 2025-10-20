package handler

import (
    "io"
    "net/http"

    "github.com/gin-gonic/gin"
    middleware "hdzk.cn/foodapp/internal/server/middleware"
    svc "hdzk.cn/foodapp/internal/service/inquiry"
)

// InquiryImportHandler handles Excel uploads for inquiry import.
type InquiryImportHandler struct{ s *svc.ImportService }

func NewInquiryImportHandler(s *svc.ImportService) *InquiryImportHandler { return &InquiryImportHandler{s: s} }

func (h *InquiryImportHandler) Register(rg *gin.RouterGroup) {
    g := rg.Group("/inquiry")
    g.POST("/import_excel", h.importExcel)
}

func (h *InquiryImportHandler) importExcel(c *gin.Context) {
    const errTitle = "导入询价失败"
    act := middleware.GetActor(c)
    if act.Deleted != middleware.DeletedNo {
        ForbiddenError(c, errTitle, "账户已删除，禁止操作")
        return
    }
    // admin only
    if act.Role != middleware.RoleAdmin {
        ForbiddenError(c, errTitle, "仅管理员可导入")
        return
    }

    orgID := c.PostForm("org_id")
    file, header, err := c.Request.FormFile("file")
    _ = header
    if err != nil {
        BadRequest(c, errTitle, "缺少文件")
        return
    }
    defer file.Close()
    b, err := io.ReadAll(file)
    if err != nil {
        InternalError(c, errTitle, "读取文件失败")
        return
    }
    id, err := h.s.Import(c, svc.ImportParams{OrgID: orgID, Data: b})
    if err != nil {
        ConflictError(c, errTitle, err.Error())
        return
    }
    c.JSON(http.StatusCreated, gin.H{"inquiry_id": id})
}
