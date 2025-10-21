package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/inquiry_item"
	types "hdzk.cn/foodapp/internal/transport"
)

type InquiryItemHandler struct{ s *svc.Service }

func NewInquiryItemHandler(s *svc.Service) *InquiryItemHandler { return &InquiryItemHandler{s: s} }

func (h *InquiryItemHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/inquiry_item")

	g.POST("/create_inquiry_item", h.create)
	g.POST("/get_inquiry_item", h.get)
	g.POST("/list_inquiry_items", h.list)
	g.POST("/update_inquiry_item", h.update)
	g.POST("/soft_delete_inquiry_item", h.softDelete)
	g.POST("/hard_delete_inquiry_item", h.hardDelete)
	g.POST("/batch_create_inquiry_items", h.batchCreate)
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
	GoodsNameSnap     *string  `json:"goods_name_snap" binding:"omitempty,min=1,max=128"`
	CategoryNameSnap  *string  `json:"category_name_snap" binding:"omitempty,min=1,max=64"`
	SpecNameSnap      *string  `json:"spec_name_snap" binding:"omitempty,max=32"`
	UnitNameSnap      *string  `json:"unit_name_snap" binding:"omitempty,max=32"`
	GuidePrice        *float64 `json:"guide_price" binding:"omitempty,min=0"`
	LastMonthAvgPrice *float64 `json:"last_month_avg_price" binding:"omitempty,min=0"`
	CurrentAvgPrice   *float64 `json:"current_avg_price" binding:"omitempty,min=0"`
	Sort              *int     `json:"sort" binding:"omitempty,min=0"`
}

type batchCreateInquiryItemsReq struct {
	Items []inquiryItemCreateReq `json:"items" binding:"required,min=1"`
}

func (h *InquiryItemHandler) create(c *gin.Context) {
	const errTitle = "创建询价商品失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建询价商品")
		return
	}

	var req inquiryItemCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	if strings.TrimSpace(req.GoodsNameSnap) == "" {
		BadRequest(c, errTitle, "goods_name_snap 不能为空")
		return
	}
	if strings.TrimSpace(req.CategoryNameSnap) == "" {
		BadRequest(c, errTitle, "category_name_snap 不能为空")
		return
	}

	params := svc.CreateParams{
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
	inquiryItem, err := h.s.CreateInquiryItem(c, params)
	if err != nil {
		ConflictError(c, errTitle, "创建询价商品失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, inquiryItem)
}

func (h *InquiryItemHandler) get(c *gin.Context) {
	const errTitle = "获取询价商品失败"
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

	inquiryItem, err := h.s.GetInquiryItem(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "询价商品不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, inquiryItem)
}

func (h *InquiryItemHandler) list(c *gin.Context) {
	const errTitle = "获取询价商品列表失败"
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
	goodsID := strings.TrimSpace(c.Query("goods_id"))
	var goodsPtr *string
	if goodsID != "" {
		goodsPtr = &goodsID
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	params := svc.ListParams{
		InquiryID:  inquiryID,
		CategoryID: categoryPtr,
		GoodsID:    goodsPtr,
		Page:       page,
		PageSize:   ps,
	}
	list, total, err := h.s.ListInquiryItems(c, params)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *InquiryItemHandler) update(c *gin.Context) {
	const errTitle = "更新询价商品失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新询价商品")
		return
	}

	var req inquiryItemUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.UpdateParams{
		ID:                req.ID,
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
		ConflictError(c, errTitle, "更新询价商品失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InquiryItemHandler) softDelete(c *gin.Context) {
	const errTitle = "删除询价商品失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价商品")
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
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *InquiryItemHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除询价商品失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除询价商品")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDeleteInquiryItem(c, req.ID); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InquiryItemHandler) batchCreate(c *gin.Context) {
	const errTitle = "批量创建询价商品失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可批量创建询价商品")
		return
	}

	var req batchCreateInquiryItemsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	items := make([]svc.CreateParams, 0, len(req.Items))
	for _, item := range req.Items {
		if strings.TrimSpace(item.GoodsNameSnap) == "" {
			BadRequest(c, errTitle, "goods_name_snap 不能为空")
			return
		}
		if strings.TrimSpace(item.CategoryNameSnap) == "" {
			BadRequest(c, errTitle, "category_name_snap 不能为空")
			return
		}
		items = append(items, svc.CreateParams{
			InquiryID:         item.InquiryID,
			GoodsID:           item.GoodsID,
			CategoryID:        item.CategoryID,
			SpecID:            item.SpecID,
			UnitID:            item.UnitID,
			GoodsNameSnap:     item.GoodsNameSnap,
			CategoryNameSnap:  item.CategoryNameSnap,
			SpecNameSnap:      item.SpecNameSnap,
			UnitNameSnap:      item.UnitNameSnap,
			GuidePrice:        item.GuidePrice,
			LastMonthAvgPrice: item.LastMonthAvgPrice,
			CurrentAvgPrice:   item.CurrentAvgPrice,
			Sort:              item.Sort,
		})
	}

	params := svc.BatchCreateParams{Items: items}
	if err := h.s.BatchCreateInquiryItems(c, params); err != nil {
		ConflictError(c, errTitle, "批量创建询价商品失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true, "count": len(items)})
}