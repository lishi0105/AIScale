package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	chunkImport "hdzk.cn/foodapp/internal/import"
	"hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/inquiry"
	types "hdzk.cn/foodapp/internal/transport"
	"hdzk.cn/foodapp/pkg/logger"
)

// ========== BaseMarket Handler ==========

type MarketHandler struct{ s *svc.MarketService }

func NewMarketHandler(s *svc.MarketService) *MarketHandler {
	return &MarketHandler{s: s}
}

func (h *MarketHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/market")

	g.POST("/create_market", h.createMarket)
	g.POST("/get_market", h.getMarket)
	g.POST("/list_markets", h.listMarkets)
	g.POST("/update_market", h.updateMarket)
	g.POST("/soft_delete_market", h.softDeleteMarket)
	g.POST("/hard_delete_market", h.hardDeleteMarket)
}

type marketCreateReq struct {
	Name  string  `json:"name" binding:"required,min=1,max=64"`
	OrgID string  `json:"org_id" binding:"required,uuid4"`
	Code  *string `json:"code" binding:"omitempty,min=1,max=64"`
	Sort  *int    `json:"sort" binding:"omitempty,min=0"`
}

type marketUpdateReq struct {
	ID   string  `json:"id" binding:"required,uuid4"`
	Name *string `json:"name" binding:"omitempty,min=1,max=64"`
	Code *string `json:"code" binding:"omitempty,min=1,max=64"`
	Sort *int    `json:"sort" binding:"omitempty,min=0"`
}

func (h *MarketHandler) createMarket(c *gin.Context) {
	const errTitle = "创建市场失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建市场")
		return
	}

	var req marketCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		BadRequest(c, errTitle, "name 不能为空")
		return
	}

	params := svc.MarketCreateParams{
		Name:  req.Name,
		OrgID: req.OrgID,
		Code:  req.Code,
		Sort:  req.Sort,
	}
	market, err := h.s.CreateMarket(c, params)
	if err != nil {
		ForbiddenError(c, errTitle, "创建市场失败: "+err.Error())
		return
	}
	SuccessResponse(c, market)
}

func (h *MarketHandler) getMarket(c *gin.Context) {
	const errTitle = "获取市场失败"
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

	market, err := h.s.GetMarket(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "市场不存在: "+err.Error())
		return
	}
	SuccessResponse(c, market)
}

func (h *MarketHandler) listMarkets(c *gin.Context) {
	const errTitle = "获取市场列表失败"
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
	list, total, err := h.s.ListMarkets(c, kw, orgID, page, ps)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	SuccessResponse(c, map[string]any{"total": total, "items": list})
}

func (h *MarketHandler) updateMarket(c *gin.Context) {
	const errTitle = "更新市场失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新市场")
		return
	}

	var req marketUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.MarketUpdateParams{
		ID:   req.ID,
		Name: req.Name,
		Code: req.Code,
		Sort: req.Sort,
	}
	if err := h.s.UpdateMarket(c, params); err != nil {
		ForbiddenError(c, errTitle, "更新市场失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MarketHandler) softDeleteMarket(c *gin.Context) {
	const errTitle = "删除市场失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除市场")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDeleteMarket(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	SuccessResponse(c, nil)
}

func (h *MarketHandler) hardDeleteMarket(c *gin.Context) {
	const errTitle = "删除市场失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除市场")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDeleteMarket(c, req.ID); err != nil {
		ForbiddenError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ========== BasePriceInquiry Handler ==========

type InquiryHandler struct{ s *svc.InquiryService }

func NewInquiryHandler(s *svc.InquiryService) *InquiryHandler {
	return &InquiryHandler{s: s}
}

func (h *InquiryHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/inquiry")

	g.POST("/create_inquiry", h.createInquiry)
	g.POST("/get_inquiry", h.getInquiry)
	g.POST("/list_inquiries", h.listInquiries)
	g.POST("/update_inquiry", h.updateInquiry)
	g.POST("/soft_delete_inquiry", h.softDeleteInquiry)
	g.POST("/hard_delete_inquiry", h.hardDeleteInquiry)
}

type inquiryCreateReq struct {
	OrgID        string `json:"org_id" binding:"required,uuid4"`
	InquiryTitle string `json:"inquiry_title" binding:"required,min=1,max=64"`
	InquiryDate  string `json:"inquiry_date" binding:"required"` // YYYY-MM-DD format
}

type inquiryUpdateReq struct {
	ID           string  `json:"id" binding:"required,uuid4"`
	InquiryTitle *string `json:"inquiry_title" binding:"omitempty,min=1,max=64"`
	InquiryDate  *string `json:"inquiry_date" binding:"omitempty"` // YYYY-MM-DD format
}

func (h *InquiryHandler) createInquiry(c *gin.Context) {
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

	var req inquiryCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.InquiryCreateParams{
		OrgID:        req.OrgID,
		InquiryTitle: req.InquiryTitle,
		InquiryDate:  req.InquiryDate,
	}
	inquiry, err := h.s.CreateInquiry(c, params)
	if err != nil {
		ForbiddenError(c, errTitle, "创建询价单失败: "+err.Error())
		return
	}
	SuccessResponse(c, inquiry)
}

func (h *InquiryHandler) getInquiry(c *gin.Context) {
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

	inquiry, err := h.s.GetInquiry(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "询价单不存在: "+err.Error())
		return
	}
	SuccessResponse(c, inquiry)
}

func (h *InquiryHandler) listInquiries(c *gin.Context) {
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

	var year, month, tenDay *int
	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y > 0 {
			year = &y
		}
	}
	if monthStr := c.Query("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m > 0 {
			month = &m
		}
	}
	if tenDayStr := c.Query("ten_day"); tenDayStr != "" {
		if t, err := strconv.Atoi(tenDayStr); err == nil && t > 0 {
			tenDay = &t
		}
	}

	kw := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListInquiries(c, kw, orgID, year, month, tenDay, page, ps)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	SuccessResponse(c, map[string]any{"total": total, "items": list})
}

func (h *InquiryHandler) updateInquiry(c *gin.Context) {
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

	var req inquiryUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.InquiryUpdateParams{
		ID:           req.ID,
		InquiryTitle: req.InquiryTitle,
		InquiryDate:  req.InquiryDate,
	}
	if err := h.s.UpdateInquiry(c, params); err != nil {
		ForbiddenError(c, errTitle, "更新询价单失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InquiryHandler) softDeleteInquiry(c *gin.Context) {
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
	if err := h.s.SoftDeleteInquiry(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	SuccessResponse(c, nil)
}

func (h *InquiryHandler) hardDeleteInquiry(c *gin.Context) {
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
	if err := h.s.HardDeleteInquiry(c, req.ID); err != nil {
		ForbiddenError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ========== PriceInquiryItem Handler ==========

type InquiryItemHandler struct{ s *svc.InquiryItemService }

func NewInquiryItemHandler(s *svc.InquiryItemService) *InquiryItemHandler {
	return &InquiryItemHandler{s: s}
}

func (h *InquiryItemHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/inquiry_item")

	g.POST("/create_inquiry_item", h.createInquiryItem)
	g.POST("/get_inquiry_item", h.getInquiryItem)
	g.POST("/list_inquiry_items", h.listInquiryItems)
	g.POST("/update_inquiry_item", h.updateInquiryItem)
	g.POST("/soft_delete_inquiry_item", h.softDeleteInquiryItem)
	g.POST("/hard_delete_inquiry_item", h.hardDeleteInquiryItem)
}

type inquiryItemCreateReq struct {
	InquiryID         string   `json:"inquiry_id" binding:"required,uuid4"`
	GoodsID           string   `json:"goods_id" binding:"required,uuid4"`
	CategoryID        string   `json:"category_id" binding:"required,uuid4"`
	SpecID            *string  `json:"spec_id" binding:"omitempty,uuid4"`
	UnitID            *string  `json:"unit_id" binding:"omitempty,uuid4"`
	GoodsNameSnap     string   `json:"goods_name_snap" binding:"required,min=1,max=128"`
	CategoryNameSnap  string   `json:"category_name_snap" binding:"required,min=1,max=64"`
	SpecNameSnap      *string  `json:"spec_name_snap" binding:"omitempty,max=32"`
	UnitNameSnap      *string  `json:"unit_name_snap" binding:"omitempty,max=32"`
	GuidePrice        *float64 `json:"guide_price" binding:"omitempty,min=0"`
	LastMonthAvgPrice *float64 `json:"last_month_avg_price" binding:"omitempty,min=0"`
	CurrentAvgPrice   *float64 `json:"current_avg_price" binding:"omitempty,min=0"`
	Sort              *int     `json:"sort" binding:"omitempty,min=0"`
}

type inquiryItemUpdateReq struct {
	ID                string   `json:"id" binding:"required,uuid4"`
	GoodsID           *string  `json:"goods_id" binding:"omitempty,uuid4"`
	CategoryID        *string  `json:"category_id" binding:"omitempty,uuid4"`
	SpecID            *string  `json:"spec_id" binding:"omitempty,uuid4"`
	UnitID            *string  `json:"unit_id" binding:"omitempty,uuid4"`
	GoodsNameSnap     *string  `json:"goods_name_snap" binding:"omitempty,min=1,max=128"`
	CategoryNameSnap  *string  `json:"category_name_snap" binding:"omitempty,min=1,max=64"`
	SpecNameSnap      *string  `json:"spec_name_snap" binding:"omitempty,max=32"`
	UnitNameSnap      *string  `json:"unit_name_snap" binding:"omitempty,max=32"`
	GuidePrice        *float64 `json:"guide_price" binding:"omitempty,min=0"`
	LastMonthAvgPrice *float64 `json:"last_month_avg_price" binding:"omitempty,min=0"`
	CurrentAvgPrice   *float64 `json:"current_avg_price" binding:"omitempty,min=0"`
	Sort              *int     `json:"sort" binding:"omitempty,min=0"`
}

func (h *InquiryItemHandler) createInquiryItem(c *gin.Context) {
	const errTitle = "创建询价商品明细失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建询价商品明细")
		return
	}

	var req inquiryItemCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.InquiryItemCreateParams{
		InquiryID:         req.InquiryID,
		GoodsID:           req.GoodsID,
		CategoryID:        req.CategoryID,
		SpecID:            req.SpecID,
		UnitID:            req.UnitID,
		GoodsNameSnap:     req.GoodsNameSnap,
		CategoryNameSnap:  req.CategoryNameSnap,
		SpecNameSnap:      req.SpecNameSnap,
		UnitNameSnap:      req.UnitNameSnap,
		GuidePrice:        req.GuidePrice,
		LastMonthAvgPrice: req.LastMonthAvgPrice,
		CurrentAvgPrice:   req.CurrentAvgPrice,
		Sort:              req.Sort,
	}
	item, err := h.s.CreateInquiryItem(c, params)
	if err != nil {
		ForbiddenError(c, errTitle, "创建询价商品明细失败: "+err.Error())
		return
	}
	SuccessResponse(c, item)
}

func (h *InquiryItemHandler) getInquiryItem(c *gin.Context) {
	const errTitle = "获取询价商品明细失败"
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

	item, err := h.s.GetInquiryItem(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "询价商品明细不存在: "+err.Error())
		return
	}
	SuccessResponse(c, item)
}

func (h *InquiryItemHandler) listInquiryItems(c *gin.Context) {
	const errTitle = "获取询价商品明细列表失败"
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

	categoryID := strings.TrimSpace(c.Query("category_id"))
	var categoryPtr *string
	if categoryID != "" {
		categoryPtr = &categoryID
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListInquiryItems(c, inquiryID, categoryPtr, page, ps)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	SuccessResponse(c, map[string]any{"total": total, "items": list})
}

func (h *InquiryItemHandler) updateInquiryItem(c *gin.Context) {
	const errTitle = "更新询价商品明细失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新询价商品明细")
		return
	}

	var req inquiryItemUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.InquiryItemUpdateParams{
		ID:                req.ID,
		GoodsID:           req.GoodsID,
		CategoryID:        req.CategoryID,
		SpecID:            req.SpecID,
		UnitID:            req.UnitID,
		GoodsNameSnap:     req.GoodsNameSnap,
		CategoryNameSnap:  req.CategoryNameSnap,
		SpecNameSnap:      req.SpecNameSnap,
		UnitNameSnap:      req.UnitNameSnap,
		GuidePrice:        req.GuidePrice,
		LastMonthAvgPrice: req.LastMonthAvgPrice,
		CurrentAvgPrice:   req.CurrentAvgPrice,
		Sort:              req.Sort,
	}
	if err := h.s.UpdateInquiryItem(c, params); err != nil {
		ForbiddenError(c, errTitle, "更新询价商品明细失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InquiryItemHandler) softDeleteInquiryItem(c *gin.Context) {
	const errTitle = "删除询价商品明细失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价商品明细")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDeleteInquiryItem(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	SuccessResponse(c, nil)
}

func (h *InquiryItemHandler) hardDeleteInquiryItem(c *gin.Context) {
	const errTitle = "删除询价商品明细失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价商品明细")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDeleteInquiryItem(c, req.ID); err != nil {
		ForbiddenError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ========== PriceMarketInquiry Handler ==========

type MarketInquiryHandler struct{ s *svc.MarketInquiryService }

func NewMarketInquiryHandler(s *svc.MarketInquiryService) *MarketInquiryHandler {
	return &MarketInquiryHandler{s: s}
}

func (h *MarketInquiryHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/market_inquiry")

	g.POST("/create_market_inquiry", h.createMarketInquiry)
	g.POST("/get_market_inquiry", h.getMarketInquiry)
	g.POST("/list_market_inquiries", h.listMarketInquiries)
	g.POST("/update_market_inquiry", h.updateMarketInquiry)
	g.POST("/soft_delete_market_inquiry", h.softDeleteMarketInquiry)
	g.POST("/hard_delete_market_inquiry", h.hardDeleteMarketInquiry)
}

type marketInquiryCreateReq struct {
	InquiryID      string   `json:"inquiry_id" binding:"required,uuid4"`
	ItemID         string   `json:"item_id" binding:"required,uuid4"`
	MarketID       *string  `json:"market_id" binding:"omitempty,uuid4"`
	MarketNameSnap string   `json:"market_name_snap" binding:"required,min=1,max=64"`
	Price          *float64 `json:"price" binding:"omitempty,min=0"`
}

type marketInquiryUpdateReq struct {
	ID             string   `json:"id" binding:"required,uuid4"`
	MarketID       *string  `json:"market_id" binding:"omitempty,uuid4"`
	MarketNameSnap *string  `json:"market_name_snap" binding:"omitempty,min=1,max=64"`
	Price          *float64 `json:"price" binding:"omitempty,min=0"`
}

func (h *MarketInquiryHandler) createMarketInquiry(c *gin.Context) {
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

	params := svc.MarketInquiryCreateParams{
		InquiryID:      req.InquiryID,
		ItemID:         req.ItemID,
		MarketID:       req.MarketID,
		MarketNameSnap: req.MarketNameSnap,
		Price:          req.Price,
	}
	marketInquiry, err := h.s.CreateMarketInquiry(c, params)
	if err != nil {
		ForbiddenError(c, errTitle, "创建市场报价失败: "+err.Error())
		return
	}
	SuccessResponse(c, marketInquiry)
}

func (h *MarketInquiryHandler) getMarketInquiry(c *gin.Context) {
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
	SuccessResponse(c, marketInquiry)
}

func (h *MarketInquiryHandler) listMarketInquiries(c *gin.Context) {
	const errTitle = "获取市场报价列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	inquiryID := strings.TrimSpace(c.Query("inquiry_id"))
	var inquiryPtr *string
	if inquiryID != "" {
		inquiryPtr = &inquiryID
	}

	itemID := strings.TrimSpace(c.Query("item_id"))
	var itemPtr *string
	if itemID != "" {
		itemPtr = &itemID
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListMarketInquiries(c, inquiryPtr, itemPtr, page, ps)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	SuccessResponse(c, map[string]any{"total": total, "items": list})
}

func (h *MarketInquiryHandler) updateMarketInquiry(c *gin.Context) {
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

	params := svc.MarketInquiryUpdateParams{
		ID:             req.ID,
		MarketID:       req.MarketID,
		MarketNameSnap: req.MarketNameSnap,
		Price:          req.Price,
	}
	if err := h.s.UpdateMarketInquiry(c, params); err != nil {
		ForbiddenError(c, errTitle, "更新市场报价失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MarketInquiryHandler) softDeleteMarketInquiry(c *gin.Context) {
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
	SuccessResponse(c, nil)
}

func (h *MarketInquiryHandler) hardDeleteMarketInquiry(c *gin.Context) {
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
		ForbiddenError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ========== PriceSupplierSettlement Handler ==========

type SupplierSettlementHandler struct {
	s *svc.SupplierSettlementService
}

func NewSupplierSettlementHandler(s *svc.SupplierSettlementService) *SupplierSettlementHandler {
	return &SupplierSettlementHandler{s: s}
}

func (h *SupplierSettlementHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/supplier_settlement")

	g.POST("/create_supplier_settlement", h.createSupplierSettlement)
	g.POST("/get_supplier_settlement", h.getSupplierSettlement)
	g.POST("/list_supplier_settlements", h.listSupplierSettlements)
	g.POST("/update_supplier_settlement", h.updateSupplierSettlement)
	g.POST("/soft_delete_supplier_settlement", h.softDeleteSupplierSettlement)
	g.POST("/hard_delete_supplier_settlement", h.hardDeleteSupplierSettlement)
}

type supplierSettlementCreateReq struct {
	InquiryID        string   `json:"inquiry_id" binding:"required,uuid4"`
	ItemID           string   `json:"item_id" binding:"required,uuid4"`
	SupplierID       *string  `json:"supplier_id" binding:"omitempty,uuid4"`
	SupplierNameSnap string   `json:"supplier_name_snap" binding:"required,min=1,max=128"`
	FloatRatioSnap   float64  `json:"float_ratio_snap" binding:"required,min=0"`
	SettlementPrice  *float64 `json:"settlement_price" binding:"omitempty,min=0"`
}

type supplierSettlementUpdateReq struct {
	ID               string   `json:"id" binding:"required,uuid4"`
	SupplierID       *string  `json:"supplier_id" binding:"omitempty,uuid4"`
	SupplierNameSnap *string  `json:"supplier_name_snap" binding:"omitempty,min=1,max=128"`
	FloatRatioSnap   *float64 `json:"float_ratio_snap" binding:"omitempty,min=0"`
	SettlementPrice  *float64 `json:"settlement_price" binding:"omitempty,min=0"`
}

func (h *SupplierSettlementHandler) createSupplierSettlement(c *gin.Context) {
	const errTitle = "创建供应商结算失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建供应商结算")
		return
	}

	var req supplierSettlementCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.SupplierSettlementCreateParams{
		InquiryID:        req.InquiryID,
		ItemID:           req.ItemID,
		SupplierID:       req.SupplierID,
		SupplierNameSnap: req.SupplierNameSnap,
		FloatRatioSnap:   req.FloatRatioSnap,
		SettlementPrice:  req.SettlementPrice,
	}
	settlement, err := h.s.CreateSupplierSettlement(c, params)
	if err != nil {
		ForbiddenError(c, errTitle, "创建供应商结算失败: "+err.Error())
		return
	}
	SuccessResponse(c, settlement)
}

func (h *SupplierSettlementHandler) getSupplierSettlement(c *gin.Context) {
	const errTitle = "获取供应商结算失败"
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

	settlement, err := h.s.GetSupplierSettlement(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "供应商结算不存在: "+err.Error())
		return
	}
	SuccessResponse(c, settlement)
}

func (h *SupplierSettlementHandler) listSupplierSettlements(c *gin.Context) {
	const errTitle = "获取供应商结算列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	inquiryID := strings.TrimSpace(c.Query("inquiry_id"))
	var inquiryPtr *string
	if inquiryID != "" {
		inquiryPtr = &inquiryID
	}

	itemID := strings.TrimSpace(c.Query("item_id"))
	var itemPtr *string
	if itemID != "" {
		itemPtr = &itemID
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListSupplierSettlements(c, inquiryPtr, itemPtr, page, ps)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	SuccessResponse(c, map[string]any{"total": total, "items": list})
}

func (h *SupplierSettlementHandler) updateSupplierSettlement(c *gin.Context) {
	const errTitle = "更新供应商结算失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新供应商结算")
		return
	}

	var req supplierSettlementUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.SupplierSettlementUpdateParams{
		ID:               req.ID,
		SupplierID:       req.SupplierID,
		SupplierNameSnap: req.SupplierNameSnap,
		FloatRatioSnap:   req.FloatRatioSnap,
		SettlementPrice:  req.SettlementPrice,
	}
	if err := h.s.UpdateSupplierSettlement(c, params); err != nil {
		ForbiddenError(c, errTitle, "更新供应商结算失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SupplierSettlementHandler) softDeleteSupplierSettlement(c *gin.Context) {
	const errTitle = "删除供应商结算失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除供应商结算")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDeleteSupplierSettlement(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	SuccessResponse(c, nil)
}

func (h *SupplierSettlementHandler) hardDeleteSupplierSettlement(c *gin.Context) {
	const errTitle = "删除供应商结算失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除供应商结算")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDeleteSupplierSettlement(c, req.ID); err != nil {
		ForbiddenError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

type InquiryImportHandler struct {
	s         *svc.InquiryImportService
	uploadDir string
}

func NewInquiryImportHandler(s *svc.InquiryImportService, uploadDir string) *InquiryImportHandler {
	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("创建上传目录失败: %v", err))
	}
	return &InquiryImportHandler{
		s:         s,
		uploadDir: uploadDir,
	}
}

func (h *InquiryImportHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/inquiry_import")

	g.POST("/import_inquiry", h.importInquiry)
	g.POST("/validate", h.validateExcel)
	g.POST("/import_status", h.getImportStatus)
}

func (h *InquiryImportHandler) uploadChunk(c *gin.Context) {
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
	chunkIndex, err := chunkImport.UploadChunk(c, h.uploadDir)
	if err != nil {
		logger.L().Error(errTitle, zap.String("error", err.Error()))
		InternalError(c, errTitle, "上传文件切片失败")
		return
	}
	SuccessResponse(c, map[string]any{"chunk_index": chunkIndex, "message": "切片上传成功"})
}

func (h *InquiryImportHandler) mergeChunks(c *gin.Context) {
	const errTitle = "合并切片失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可合并切片")
		return
	}
	finalPath, err := chunkImport.MergeChunks(c, h.uploadDir)
	if err != nil {
		logger.L().Error(errTitle, zap.String("error", err.Error()))
		InternalError(c, errTitle, "合并切片失败")
		return
	}
	SuccessResponse(c, map[string]any{"filepath": finalPath, "message": "切片合并成功"})
}

// validateExcel 校验Excel文件结构
func (h *InquiryImportHandler) validateExcel(c *gin.Context) {
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
		if ve, ok := err.(*svc.ValidationError); ok {
			logger.L().Error(errTitle, zap.String("error", ve.Error()))
			BadRequest(c, errTitle, ve.Error())
		} else {
			logger.L().Error(errTitle, zap.String("error", ve.Error()))
			InternalError(c, errTitle, err.Error())
		}
		return
	}

	// 返回校验结果摘要
	SuccessResponse(c, map[string]any{
		"title": excelData.Title,
		"date":  excelData.InquiryDate.Format("2006-01-02"),
		"stats": map[string]any{
			"sheets":    len(excelData.Sheets),
			"markets":   len(excelData.Markets),
			"suppliers": len(excelData.Suppliers),
		},
		"message": "Excel文件校验通过",
	})
}

// importExcel 导入Excel数据（异步）
func (h *InquiryImportHandler) importInquiry(c *gin.Context) {
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

	// 1. 获取 org_id（来自表单字段）
	orgID := c.PostForm("org_id")
	if orgID == "" {
		BadRequest(c, errTitle, "缺少 org_id 参数")
		return
	}

	// 1.5 获取 force_delete 参数（可选）
	forceDelete := c.PostForm("force_delete") == "true"

	// 2. 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, errTitle, "未上传文件或文件字段名错误（应为 'file'）")
		return
	}

	// 3. 验证文件扩展名（可选但推荐）
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") &&
		!strings.HasSuffix(strings.ToLower(file.Filename), ".xls") {
		BadRequest(c, errTitle, "仅支持 .xls 或 .xlsx 格式的 Excel 文件")
		return
	}

	// 4. 保存到临时文件
	tmpFile, err := os.CreateTemp(h.uploadDir, "import_*.xlsx")
	if err != nil {
		logger.L().Error(errTitle, zap.Error(err))

		err := &svc.ImportInquiryError{
			Code:      int(svc.ErrInternal),
			Value:     "",
			Message:   "无法创建临时文件",
			InquiryID: "",
		}
		InternalError(c, errTitle, err.GetErrorDetails())
		return
	}
	tmpFilePath := tmpFile.Name()
	defer tmpFile.Close()

	// 将上传的文件写入临时文件
	srcFile, err := file.Open()
	if err != nil {
		logger.L().Error(errTitle, zap.Error(err))
		os.Remove(tmpFilePath)

		err := &svc.ImportInquiryError{
			Code:      int(svc.ErrInternal),
			Value:     "",
			Message:   "无法打开上传文件",
			InquiryID: "",
		}
		InternalError(c, errTitle, err.GetErrorDetails())
		return
	}
	defer srcFile.Close()

	if _, err := io.Copy(tmpFile, srcFile); err != nil {
		logger.L().Error(errTitle, zap.Error(err))
		os.Remove(tmpFilePath)
		err := &svc.ImportInquiryError{
			Code:      int(svc.ErrInternal),
			Value:     "",
			Message:   "保存上传文件失败",
			InquiryID: "",
		}
		InternalError(c, errTitle, err.GetErrorDetails())
		return
	}

	// 5. 校验 Excel 结构（同步）
	excelData, err := h.s.ValidateExcelStructure(tmpFilePath)
	if err != nil {
		os.Remove(tmpFilePath) // 校验失败，删除临时文件
		if ve, ok := err.(*svc.ValidationError); ok {
			logger.L().Error(errTitle, zap.String("校验Excel结构失败: ", ve.Error()))
			err := &svc.ImportInquiryError{
				Code:      int(svc.ErrValidate),
				Value:     ve.Error(),
				Message:   "校验Excel结构失败",
				InquiryID: "",
			}
			BadRequest(c, errTitle, err.GetErrorDetails())
		} else {
			logger.L().Error(errTitle, zap.String("校验Excel结构失败: ", err.Error()))
			err := &svc.ImportInquiryError{
				Code:      int(svc.ErrValidate),
				Value:     err.Error(),
				Message:   "校验Excel结构失败",
				InquiryID: "",
			}
			InternalError(c, errTitle, err.GetErrorDetails())
		}
		return
	}

	// 6. 检查重复（同步）
	if err := h.s.CheckDuplicateInquiry(orgID, excelData.Title, excelData.InquiryDate); err != nil {
		os.Remove(tmpFilePath) // 删除临时文件
		if dupErr, ok := err.(*svc.ImportInquiryError); ok {
			ConflictError(c, errTitle, dupErr.GetErrorDetails())
		} else {
			logger.L().Error(errTitle, zap.Error(err))
			err := &svc.ImportInquiryError{
				Code:      int(svc.ErrInternal),
				Value:     "",
				Message:   "检查重复失败",
				InquiryID: "",
			}
			InternalError(c, errTitle, err.GetErrorDetails())
		}
		return
	}

	// 7. 创建导入任务（内存）
	task := h.s.CreateImportTask(orgID, file.Filename, len(excelData.Sheets))

	// 8. 立即返回任务ID（异步执行导入）
	SuccessResponse(c, map[string]any{
		"ok":      true,
		"message": "Excel文件校验通过，开始异步导入",
		"task_id": task.ID,
		"stats": gin.H{
			"title":     excelData.Title,
			"sheets":    len(excelData.Sheets),
			"markets":   len(excelData.Markets),
			"suppliers": len(excelData.Suppliers),
		},
	})

	// 9. 启动异步导入goroutine
	go func() {
		defer os.Remove(tmpFilePath) // 导入完成后删除临时文件

		logger.L().Info("开始异步导入任务",
			zap.String("task_id", task.ID),
			zap.String("org_id", orgID),
			zap.String("file", file.Filename),
			zap.Bool("force_delete", forceDelete))

		if err := h.s.ImportExcelDataAsync(task.ID, excelData, orgID, forceDelete); err != nil {
			logger.L().Error("异步导入失败",
				zap.String("task_id", task.ID),
				zap.Error(err))
		} else {
			logger.L().Info("异步导入成功",
				zap.String("task_id", task.ID))
		}
	}()
}

// getImportStatus 查询导入任务状态
func (h *InquiryImportHandler) getImportStatus(c *gin.Context) {
	const errTitle = "查询导入状态失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	var req struct {
		TaskID string `json:"task_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法: "+err.Error())
		return
	}

	// 查询任务
	task, err := h.s.GetImportTask(req.TaskID)
	if err != nil {
		NotFoundError(c, errTitle, "任务不存在")
		return
	}

	// 返回任务状态
	response := gin.H{
		"ok":               true,
		"task_id":          task.ID,
		"status":           task.Status,
		"progress":         task.Progress,
		"total_sheets":     task.TotalSheets,
		"processed_sheets": task.ProcessedSheets,
		"file_name":        task.FileName,
		"created_at":       task.CreatedAt,
		"updated_at":       task.UpdatedAt,
	}

	// 根据状态添加额外信息
	switch task.Status {
	case "success":
		if task.InquiryID != "" {
			response["inquiry_id"] = task.InquiryID
		}
		response["message"] = "导入成功"
	case "failed":
		if task.ErrorMessage != "" {
			response["error_message"] = task.ErrorMessage
		}
		response["message"] = "导入失败"
	case "processing":
		response["message"] = "正在导入中..."
	case "pending":
		response["message"] = "等待处理"
	}

	SuccessResponse(c, response)
}
