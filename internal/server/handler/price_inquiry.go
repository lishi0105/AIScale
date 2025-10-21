package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/price_inquiry"
	types "hdzk.cn/foodapp/internal/transport"
)

type PriceInquiryHandler struct{ s *svc.Service }

func NewPriceInquiryHandler(s *svc.Service) *PriceInquiryHandler { return &PriceInquiryHandler{s: s} }

func (h *PriceInquiryHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/price_inquiry")

	g.POST("/create_price_inquiry", h.create)
	g.POST("/get_price_inquiry", h.get)
	g.POST("/list_price_inquiries", h.list)
	g.POST("/update_price_inquiry", h.update)
	g.POST("/soft_delete_price_inquiry", h.softDelete)
	g.POST("/hard_delete_price_inquiry", h.hardDelete)
}

type priceInquiryCreateReq struct {
	OrgID        string `json:"org_id" binding:"required,uuid4"`
	InquiryTitle string `json:"inquiry_title" binding:"required,min=1,max=64"`
	InquiryDate  string `json:"inquiry_date" binding:"required"`
}

type priceInquiryUpdateReq struct {
	ID           string  `json:"id" binding:"required,uuid4"`
	InquiryTitle *string `json:"inquiry_title" binding:"omitempty,min=1,max=64"`
	InquiryDate  *string `json:"inquiry_date" binding:"omitempty"`
}

func (h *PriceInquiryHandler) create(c *gin.Context) {
	const errTitle = "创建询价单失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建询价单")
		return
	}

	var req priceInquiryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	if strings.TrimSpace(req.InquiryTitle) == "" {
		BadRequest(c, errTitle, "inquiry_title 不能为空")
		return
	}

	inquiryDate, err := time.Parse("2006-01-02", req.InquiryDate)
	if err != nil {
		BadRequest(c, errTitle, "inquiry_date 格式错误，应为 YYYY-MM-DD")
		return
	}

	params := svc.CreateParams{
		OrgID:        req.OrgID,
		InquiryTitle: req.InquiryTitle,
		InquiryDate:  inquiryDate,
	}
	priceInquiry, err := h.s.CreatePriceInquiry(c, params)
	if err != nil {
		ConflictError(c, errTitle, "创建询价单失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, priceInquiry)
}

func (h *PriceInquiryHandler) get(c *gin.Context) {
	const errTitle = "获取询价单失败"
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

	priceInquiry, err := h.s.GetPriceInquiry(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "询价单不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, priceInquiry)
}

func (h *PriceInquiryHandler) list(c *gin.Context) {
	const errTitle = "获取询价单列表失败"
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
	yearStr := c.Query("year")
	monthStr := c.Query("month")
	tenDayStr := c.Query("ten_day")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var year, month, tenDay *int
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = &y
		}
	}
	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil {
			month = &m
		}
	}
	if tenDayStr != "" {
		if td, err := strconv.Atoi(tenDayStr); err == nil {
			tenDay = &td
		}
	}

	params := svc.ListParams{
		Keyword:  keyword,
		OrgID:    orgID,
		Year:     year,
		Month:    month,
		TenDay:   tenDay,
		Page:     page,
		PageSize: ps,
	}
	list, total, err := h.s.ListPriceInquiries(c, params)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *PriceInquiryHandler) update(c *gin.Context) {
	const errTitle = "更新询价单失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新询价单")
		return
	}

	var req priceInquiryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	var inquiryDate *time.Time
	if req.InquiryDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.InquiryDate); err == nil {
			inquiryDate = &parsed
		} else {
			BadRequest(c, errTitle, "inquiry_date 格式错误，应为 YYYY-MM-DD")
			return
		}
	}

	params := svc.UpdateParams{
		ID:           req.ID,
		InquiryTitle: req.InquiryTitle,
		InquiryDate:  inquiryDate,
	}
	if err := h.s.UpdatePriceInquiry(c, params); err != nil {
		ConflictError(c, errTitle, "更新询价单失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PriceInquiryHandler) softDelete(c *gin.Context) {
	const errTitle = "删除询价单失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价单")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDeletePriceInquiry(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PriceInquiryHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除询价单失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价单")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDeletePriceInquiry(c, req.ID); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}