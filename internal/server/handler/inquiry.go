package handler

import (
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
	g := rg.Group("/inquiries")

	g.POST("/create_inquiry", h.create)
	g.POST("/get_inquiry", h.get)
	g.POST("/list_inquiries", h.list)
	g.POST("/update_inquiry", h.update)
	g.POST("/soft_delete_inquiry", h.softDelete)
	g.POST("/hard_delete_inquiry", h.hardDelete)
}

type inquiryCreateReq struct {
	InquiryTitle     string  `json:"inquiry_title" binding:"required,min=1,max=64"`
	InquiryDate      string  `json:"inquiry_date" binding:"required"`
	Market1          *string `json:"market_1" binding:"omitempty,max=128"`
	Market2          *string `json:"market_2" binding:"omitempty,max=128"`
	Market3          *string `json:"market_3" binding:"omitempty,max=128"`
	OrgID            string  `json:"org_id" binding:"required,uuid4"`
	InquiryStartDate string  `json:"inquiry_start_date" binding:"required"`
	InquiryEndDate   string  `json:"inquiry_end_date" binding:"required"`
}

type inquiryUpdateReq struct {
	ID               string  `json:"id" binding:"required,uuid4"`
	InquiryTitle     *string `json:"inquiry_title" binding:"omitempty,min=1,max=64"`
	InquiryDate      *string `json:"inquiry_date" binding:"omitempty"`
	Market1          *string `json:"market_1" binding:"omitempty,max=128"`
	Market2          *string `json:"market_2" binding:"omitempty,max=128"`
	Market3          *string `json:"market_3" binding:"omitempty,max=128"`
	InquiryStartDate *string `json:"inquiry_start_date" binding:"omitempty"`
	InquiryEndDate   *string `json:"inquiry_end_date" binding:"omitempty"`
}

func (h *InquiryHandler) create(c *gin.Context) {
	const errTitle = "创建询价记录失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建询价记录")
		return
	}

	var req inquiryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法: "+err.Error())
		return
	}

	if strings.TrimSpace(req.InquiryTitle) == "" {
		BadRequest(c, errTitle, "inquiry_title 不能为空")
		return
	}

	// 解析日期
	inquiryDate, err := time.Parse("2006-01-02", req.InquiryDate)
	if err != nil {
		BadRequest(c, errTitle, "inquiry_date 格式错误，应为 YYYY-MM-DD")
		return
	}

	// 解析时间
	startDate, err := time.Parse(time.RFC3339, req.InquiryStartDate)
	if err != nil {
		BadRequest(c, errTitle, "inquiry_start_date 格式错误，应为 RFC3339 格式")
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.InquiryEndDate)
	if err != nil {
		BadRequest(c, errTitle, "inquiry_end_date 格式错误，应为 RFC3339 格式")
		return
	}

	params := svc.CreateParams{
		InquiryTitle:     req.InquiryTitle,
		InquiryDate:      inquiryDate,
		Market1:          req.Market1,
		Market2:          req.Market2,
		Market3:          req.Market3,
		OrgID:            req.OrgID,
		InquiryStartDate: startDate,
		InquiryEndDate:   endDate,
	}
	inquiry, err := h.s.CreateInquiry(c, params)
	if err != nil {
		ConflictError(c, errTitle, "创建询价记录失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, inquiry)
}

func (h *InquiryHandler) get(c *gin.Context) {
	const errTitle = "获取询价记录失败"
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

	inquiry, err := h.s.GetInquiry(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "询价记录不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, inquiry)
}

func (h *InquiryHandler) list(c *gin.Context) {
	const errTitle = "获取询价记录列表失败"
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

	kw := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 日期范围过滤（可选）
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	list, total, err := h.s.ListInquiries(c, kw, orgID, startDate, endDate, page, ps)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *InquiryHandler) update(c *gin.Context) {
	const errTitle = "更新询价记录失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新询价记录")
		return
	}

	var req inquiryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法: "+err.Error())
		return
	}

	params := svc.UpdateParams{
		ID:           req.ID,
		InquiryTitle: req.InquiryTitle,
		Market1:      req.Market1,
		Market2:      req.Market2,
		Market3:      req.Market3,
	}

	// 解析日期（如果提供）
	if req.InquiryDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.InquiryDate)
		if err != nil {
			BadRequest(c, errTitle, "inquiry_date 格式错误，应为 YYYY-MM-DD")
			return
		}
		params.InquiryDate = &parsed
	}

	// 解析开始时间（如果提供）
	if req.InquiryStartDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.InquiryStartDate)
		if err != nil {
			BadRequest(c, errTitle, "inquiry_start_date 格式错误，应为 RFC3339 格式")
			return
		}
		params.InquiryStartDate = &parsed
	}

	// 解析结束时间（如果提供）
	if req.InquiryEndDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.InquiryEndDate)
		if err != nil {
			BadRequest(c, errTitle, "inquiry_end_date 格式错误，应为 RFC3339 格式")
			return
		}
		params.InquiryEndDate = &parsed
	}

	if err := h.s.UpdateInquiry(c, params); err != nil {
		ConflictError(c, errTitle, "更新询价记录失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InquiryHandler) softDelete(c *gin.Context) {
	const errTitle = "删除询价记录失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价记录")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDeleteInquiry(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *InquiryHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除询价记录失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价记录")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDeleteInquiry(c, req.ID); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
