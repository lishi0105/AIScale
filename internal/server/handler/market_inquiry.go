package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/market_inquiry"
	types "hdzk.cn/foodapp/internal/transport"
)

type MarketInquiryHandler struct{ s *svc.Service }

func NewMarketInquiryHandler(s *svc.Service) *MarketInquiryHandler { return &MarketInquiryHandler{s: s} }

func (h *MarketInquiryHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/market_inquiry")

	g.POST("/create_market_inquiry", h.create)
	g.POST("/get_market_inquiry", h.get)
	g.POST("/list_market_inquiries", h.list)
	g.POST("/update_market_inquiry", h.update)
	g.POST("/soft_delete_market_inquiry", h.softDelete)
	g.POST("/hard_delete_market_inquiry", h.hardDelete)
	g.POST("/batch_create_market_inquiries", h.batchCreate)
	g.POST("/get_by_item_and_market", h.getByItemAndMarket)
}

type marketInquiryCreateReq struct {
	InquiryID      string   `json:"inquiry_id" binding:"required,uuid4"`
	ItemID         string   `json:"item_id" binding:"required,uuid4"`
	MarketID       *string  `json:"market_id" binding:"omitempty,uuid4"`
	MarketNameSnap string   `json:"market_name_snap" binding:"required,min=1,max=64"`
	Price          float64  `json:"price" binding:"required,min=0"`
}

type marketInquiryUpdateReq struct {
	ID             string   `json:"id" binding:"required,uuid4"`
	MarketID       *string  `json:"market_id" binding:"omitempty,uuid4"`
	MarketNameSnap *string  `json:"market_name_snap" binding:"omitempty,min=1,max=64"`
	Price          *float64 `json:"price" binding:"omitempty,min=0"`
}

type batchCreateMarketInquiriesReq struct {
	Inquiries []marketInquiryCreateReq `json:"inquiries" binding:"required,min=1"`
}

type getByItemAndMarketReq struct {
	ItemID      string `json:"item_id" binding:"required,uuid4"`
	MarketName  string `json:"market_name" binding:"required,min=1,max=64"`
}

func (h *MarketInquiryHandler) create(c *gin.Context) {
	const errTitle = "创建市场报价失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建市场报价")
		return
	}

	var req marketInquiryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	if strings.TrimSpace(req.MarketNameSnap) == "" {
		BadRequest(c, errTitle, "market_name_snap 不能为空")
		return
	}
	if req.Price < 0 {
		BadRequest(c, errTitle, "price 不能为负数")
		return
	}

	params := svc.CreateParams{
		InquiryID:      req.InquiryID,
		ItemID:         req.ItemID,
		MarketID:       req.MarketID,
		MarketNameSnap: req.MarketNameSnap,
		Price:          req.Price,
	}
	marketInquiry, err := h.s.CreateMarketInquiry(c, params)
	if err != nil {
		ConflictError(c, errTitle, "创建市场报价失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, marketInquiry)
}

func (h *MarketInquiryHandler) get(c *gin.Context) {
	const errTitle = "获取市场报价失败"
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

	marketInquiry, err := h.s.GetMarketInquiry(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "市场报价不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, marketInquiry)
}

func (h *MarketInquiryHandler) list(c *gin.Context) {
	const errTitle = "获取市场报价列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	inquiryID := strings.TrimSpace(c.Query("inquiry_id"))
	if inquiryID == "" {
		BadRequest(c, errTitle, "参数错误：缺少 inquiry_id")
		return
	}

	itemID := strings.TrimSpace(c.Query("item_id"))
	var itemPtr *string
	if itemID != "" {
		itemPtr = &itemID
	}
	marketID := strings.TrimSpace(c.Query("market_id"))
	var marketPtr *string
	if marketID != "" {
		marketPtr = &marketID
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	params := svc.ListParams{
		InquiryID: inquiryID,
		ItemID:    itemPtr,
		MarketID:  marketPtr,
		Page:      page,
		PageSize:  ps,
	}
	list, total, err := h.s.ListMarketInquiries(c, params)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *MarketInquiryHandler) update(c *gin.Context) {
	const errTitle = "更新市场报价失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新市场报价")
		return
	}

	var req marketInquiryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.UpdateParams{
		ID:             req.ID,
		MarketID:       req.MarketID,
		MarketNameSnap: req.MarketNameSnap,
		Price:          req.Price,
	}
	if err := h.s.UpdateMarketInquiry(c, params); err != nil {
		ConflictError(c, errTitle, "更新市场报价失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MarketInquiryHandler) softDelete(c *gin.Context) {
	const errTitle = "删除市场报价失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除市场报价")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDeleteMarketInquiry(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MarketInquiryHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除市场报价失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除市场报价")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDeleteMarketInquiry(c, req.ID); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MarketInquiryHandler) batchCreate(c *gin.Context) {
	const errTitle = "批量创建市场报价失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可批量创建市场报价")
		return
	}

	var req batchCreateMarketInquiriesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	inquiries := make([]svc.CreateParams, 0, len(req.Inquiries))
	for _, inquiry := range req.Inquiries {
		if strings.TrimSpace(inquiry.MarketNameSnap) == "" {
			BadRequest(c, errTitle, "market_name_snap 不能为空")
			return
		}
		if inquiry.Price < 0 {
			BadRequest(c, errTitle, "price 不能为负数")
			return
		}
		inquiries = append(inquiries, svc.CreateParams{
			InquiryID:      inquiry.InquiryID,
			ItemID:         inquiry.ItemID,
			MarketID:       inquiry.MarketID,
			MarketNameSnap: inquiry.MarketNameSnap,
			Price:          inquiry.Price,
		})
	}

	params := svc.BatchCreateParams{Inquiries: inquiries}
	if err := h.s.BatchCreateMarketInquiries(c, params); err != nil {
		ConflictError(c, errTitle, "批量创建市场报价失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true, "count": len(inquiries)})
}

func (h *MarketInquiryHandler) getByItemAndMarket(c *gin.Context) {
	const errTitle = "获取市场报价失败"
	var req getByItemAndMarketReq

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	marketInquiry, err := h.s.GetByItemAndMarket(c, req.ItemID, req.MarketName)
	if err != nil {
		NotFoundError(c, errTitle, "市场报价不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, marketInquiry)
}