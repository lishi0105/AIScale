package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/inquiry"
	types "hdzk.cn/foodapp/internal/transport"
)

type InquiryHandler struct{ s *svc.Service }

func NewInquiryHandler(s *svc.Service) *InquiryHandler { return &InquiryHandler{s: s} }

func (h *InquiryHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/inquiry")
	g.POST("/create_inquiry", h.create)
	g.POST("/get_inquiry", h.get)
	g.POST("/list_inquiry", h.list)
	g.POST("/update_inquiry", h.update)
	g.POST("/soft_delete_inquiry", h.softDelete)
	g.POST("/hard_delete_inquiry", h.hardDelete)
	g.POST("/import_excel", h.importExcel)
}

type inquiryCreateReq struct {
	OrgID        string  `json:"org_id" binding:"required,uuid4"`
	InquiryTitle string  `json:"inquiry_title" binding:"required,min=1,max=64"`
	InquiryDate  string  `json:"inquiry_date" binding:"required"` // YYYY-MM-DD
	Market1      *string `json:"market_1" binding:"omitempty,max=128"`
	Market2      *string `json:"market_2" binding:"omitempty,max=128"`
	Market3      *string `json:"market_3" binding:"omitempty,max=128"`
}

type inquiryUpdateReq struct {
	ID           string  `json:"id" binding:"required,uuid4"`
	InquiryTitle *string `json:"inquiry_title" binding:"omitempty,min=1,max=64"`
	InquiryDate  *string `json:"inquiry_date"`
	Market1      *string `json:"market_1" binding:"omitempty,max=128"`
	Market2      *string `json:"market_2" binding:"omitempty,max=128"`
	Market3      *string `json:"market_3" binding:"omitempty,max=128"`
}

func parseDate(raw string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(raw), time.Local)
}

func parseDateTime(raw string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(raw), time.Local)
}

func parseOptionalDate(p *string) (*time.Time, error) {
	if p == nil {
		return nil, nil
	}
	t, err := parseDate(*p)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func parseOptionalDateTime(p *string) (*time.Time, error) {
	if p == nil {
		return nil, nil
	}
	t, err := parseDateTime(*p)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *InquiryHandler) create(c *gin.Context) {
	const errTitle = "创建询价失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建询价")
		return
	}

	var req inquiryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	d, err := parseDate(req.InquiryDate)
	if err != nil {
		BadRequest(c, errTitle, "inquiry_date 格式应为 YYYY-MM-DD")
		return
	}

	params := svc.CreateParams{
		OrgID:        req.OrgID,
		InquiryTitle: req.InquiryTitle,
		InquiryDate:  d,
		Market1:      req.Market1,
		Market2:      req.Market2,
		Market3:      req.Market3,
	}
	out, err := h.s.Create(c, params)
	if err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *InquiryHandler) get(c *gin.Context) {
	const errTitle = "获取询价失败"
	var req types.IDReq

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	out, err := h.s.Get(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "询价不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *InquiryHandler) list(c *gin.Context) {
	const errTitle = "获取询价列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	orgID := strings.TrimSpace(c.Query("org_id"))
	if orgID == "" {
		BadRequest(c, errTitle, "参数错误：缺少 org_id")
		return
	}

	keyword := c.Query("keyword")
	var fromPtr, toPtr *time.Time
	fromRaw := strings.TrimSpace(c.Query("date_from"))
	toRaw := strings.TrimSpace(c.Query("date_to"))
	if fromRaw != "" {
		if t, err := parseDate(fromRaw); err == nil {
			fromPtr = &t
		} else {
			BadRequest(c, errTitle, "date_from 格式应为 YYYY-MM-DD")
			return
		}
	}
	if toRaw != "" {
		if t, err := parseDate(toRaw); err == nil {
			toPtr = &t
		} else {
			BadRequest(c, errTitle, "date_to 格式应为 YYYY-MM-DD")
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.List(c, orgID, keyword, fromPtr, toPtr, page, ps)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *InquiryHandler) importExcel(c *gin.Context) {
	const errTitle = "导入询价失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可导入询价")
		return
	}

	orgID := strings.TrimSpace(c.PostForm("org_id"))
	if orgID == "" {
		BadRequest(c, errTitle, "缺少 org_id")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, errTitle, "请上传 Excel 文件")
		return
	}
	fh, err := file.Open()
	if err != nil {
		InternalError(c, errTitle, "读取文件失败")
		return
	}
	defer fh.Close()

	data, err := io.ReadAll(fh)
	if err != nil {
		InternalError(c, errTitle, "读取文件内容失败")
		return
	}

	results, err := h.s.ImportExcel(c, orgID, data)
	if err != nil {
		var ve *svc.ValidationError
		if errors.As(err, &ve) {
			BadRequest(c, errTitle, err.Error())
			return
		}
		ConflictError(c, errTitle, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"results": results})
}

func (h *InquiryHandler) update(c *gin.Context) {
	const errTitle = "更新询价失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新询价")
		return
	}

	var req inquiryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	datePtr, err := parseOptionalDate(req.InquiryDate)
	if err != nil {
		BadRequest(c, errTitle, "inquiry_date 格式应为 YYYY-MM-DD")
		return
	}

	params := svc.UpdateParams{
		ID:           req.ID,
		InquiryTitle: req.InquiryTitle,
		InquiryDate:  datePtr,
		Market1:      req.Market1,
		Market2:      req.Market2,
		Market3:      req.Market3,
	}
	if err := h.s.Update(c, params); err != nil {
		ConflictError(c, errTitle, "更新询价失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InquiryHandler) softDelete(c *gin.Context) {
	const errTitle = "删除询价失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价")
		return
	}
	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDelete(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *InquiryHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除询价失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价")
		return
	}
	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDelete(c, req.ID); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
