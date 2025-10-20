package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	domain "hdzk.cn/foodapp/internal/domain/inquiry"
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
	g.POST("/list_inquiries", h.list)
	g.POST("/update_inquiry", h.update)
	g.POST("/soft_delete_inquiry", h.softDelete)
	g.POST("/hard_delete_inquiry", h.hardDelete)
}

// 创建询价记录请求
type inquiryCreateReq struct {
	InquiryTitle     string    `json:"inquiry_title" binding:"required,min=1,max=64"`
	InquiryDate      time.Time `json:"inquiry_date" binding:"required"`
	Market1          *string   `json:"market_1" binding:"omitempty,max=128"`
	Market2          *string   `json:"market_2" binding:"omitempty,max=128"`
	Market3          *string   `json:"market_3" binding:"omitempty,max=128"`
	OrgID            string    `json:"org_id" binding:"required,uuid4"`
	InquiryStartDate time.Time `json:"inquiry_start_date" binding:"required"`
	InquiryEndDate   time.Time `json:"inquiry_end_date" binding:"required"`
}

// 更新询价记录请求
type inquiryUpdateReq struct {
	ID               string     `json:"id" binding:"required,uuid4"`
	InquiryTitle     *string    `json:"inquiry_title" binding:"omitempty,min=1,max=64"`
	InquiryDate      *time.Time `json:"inquiry_date"`
	Market1          *string    `json:"market_1" binding:"omitempty,max=128"`
	Market2          *string    `json:"market_2" binding:"omitempty,max=128"`
	Market3          *string    `json:"market_3" binding:"omitempty,max=128"`
	InquiryStartDate *time.Time `json:"inquiry_start_date"`
	InquiryEndDate   *time.Time `json:"inquiry_end_date"`
}

// 查询询价记录请求
type inquiryListReq struct {
	Keyword   string     `json:"keyword" binding:"omitempty,max=100"`
	OrgID     string     `json:"org_id" binding:"required,uuid4"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Market1   *string    `json:"market_1" binding:"omitempty,max=128"`
	Market2   *string    `json:"market_2" binding:"omitempty,max=128"`
	Market3   *string    `json:"market_3" binding:"omitempty,max=128"`
	Page      int        `json:"page" binding:"omitempty,min=1"`
	PageSize  int        `json:"page_size" binding:"omitempty,min=1,max=100"`
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
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	// 验证时间逻辑
	if req.InquiryEndDate.Before(req.InquiryStartDate) {
		BadRequest(c, errTitle, "结束时间必须晚于开始时间")
		return
	}

	params := domain.CreateInquiryParams{
		InquiryTitle:     req.InquiryTitle,
		InquiryDate:      req.InquiryDate,
		Market1:          req.Market1,
		Market2:          req.Market2,
		Market3:          req.Market3,
		OrgID:            req.OrgID,
		InquiryStartDate: req.InquiryStartDate,
		InquiryEndDate:   req.InquiryEndDate,
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
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	var req types.IDReq
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

	// 支持两种方式：JSON body 或 Query 参数
	var params domain.InquiryQueryParams

	// 尝试从 JSON body 解析
	var req inquiryListReq
	if err := c.ShouldBindJSON(&req); err == nil {
		// 从 JSON body 解析成功
		params = domain.InquiryQueryParams{
			Keyword:   req.Keyword,
			OrgID:     req.OrgID,
			StartDate: req.StartDate,
			EndDate:   req.EndDate,
			Market1:   req.Market1,
			Market2:   req.Market2,
			Market3:   req.Market3,
			Page:      req.Page,
			PageSize:  req.PageSize,
		}
	} else {
		// 从 Query 参数解析
		params = domain.InquiryQueryParams{
			Keyword:   c.Query("keyword"),
			OrgID:     c.Query("org_id"),
			Page:      parseInt(c.DefaultQuery("page", "1")),
			PageSize:  parseInt(c.DefaultQuery("page_size", "20")),
		}

		// 解析日期参数
		if startDateStr := c.Query("start_date"); startDateStr != "" {
			if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
				params.StartDate = &startDate
			}
		}
		if endDateStr := c.Query("end_date"); endDateStr != "" {
			if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
				params.EndDate = &endDate
			}
		}

		// 解析市场参数
		if market1 := c.Query("market_1"); market1 != "" {
			params.Market1 = &market1
		}
		if market2 := c.Query("market_2"); market2 != "" {
			params.Market2 = &market2
		}
		if market3 := c.Query("market_3"); market3 != "" {
			params.Market3 = &market3
		}
	}

	// 验证必填参数
	if strings.TrimSpace(params.OrgID) == "" {
		BadRequest(c, errTitle, "参数错误：缺少 org_id")
		return
	}

	list, total, err := h.s.ListInquiries(c, params)
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
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := domain.UpdateInquiryParams{
		ID:               req.ID,
		InquiryTitle:     req.InquiryTitle,
		InquiryDate:      req.InquiryDate,
		Market1:          req.Market1,
		Market2:          req.Market2,
		Market3:          req.Market3,
		InquiryStartDate: req.InquiryStartDate,
		InquiryEndDate:   req.InquiryEndDate,
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

// 辅助函数：解析整数
func parseInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}