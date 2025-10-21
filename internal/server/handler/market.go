package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/market"
	types "hdzk.cn/foodapp/internal/transport"
)

type MarketHandler struct{ s *svc.Service }

func NewMarketHandler(s *svc.Service) *MarketHandler { return &MarketHandler{s: s} }

func (h *MarketHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/market")

	g.POST("/create_market", h.create)
	g.POST("/get_market", h.get)
	g.POST("/list_markets", h.list)
	g.POST("/update_market", h.update)
	g.POST("/soft_delete_market", h.softDelete)
	g.POST("/hard_delete_market", h.hardDelete)
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

func (h *MarketHandler) create(c *gin.Context) {
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

	params := svc.CreateParams{
		Name:  req.Name,
		Code:  req.Code,
		OrgID: req.OrgID,
		Sort:  req.Sort,
	}
	market, err := h.s.CreateMarket(c, params)
	if err != nil {
		ConflictError(c, errTitle, "创建市场失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, market)
}

func (h *MarketHandler) get(c *gin.Context) {
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
	c.JSON(http.StatusOK, market)
}

func (h *MarketHandler) list(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *MarketHandler) update(c *gin.Context) {
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

	params := svc.UpdateParams{
		ID:   req.ID,
		Name: req.Name,
		Code: req.Code,
		Sort: req.Sort,
	}
	if err := h.s.UpdateMarket(c, params); err != nil {
		ConflictError(c, errTitle, "更新市场失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MarketHandler) softDelete(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MarketHandler) hardDelete(c *gin.Context) {
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
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}