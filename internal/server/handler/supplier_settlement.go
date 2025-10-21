package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/supplier_settlement"
	types "hdzk.cn/foodapp/internal/transport"
)

type SupplierSettlementHandler struct{ s *svc.Service }

func NewSupplierSettlementHandler(s *svc.Service) *SupplierSettlementHandler {
	return &SupplierSettlementHandler{s: s}
}

func (h *SupplierSettlementHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/supplier_settlement")

	g.POST("/create_supplier_settlement", h.create)
	g.POST("/get_supplier_settlement", h.get)
	g.POST("/list_supplier_settlements", h.list)
	g.POST("/update_supplier_settlement", h.update)
	g.POST("/delete_supplier_settlement", h.delete)
}

type supplierSettlementCreateReq struct {
	InquiryID        string   `json:"inquiry_id" binding:"required,uuid4"`
	ItemID           string   `json:"item_id" binding:"required,uuid4"`
	SupplierID       *string  `json:"supplier_id" binding:"omitempty,uuid4"`
	SupplierNameSnap string   `json:"supplier_name_snap" binding:"required,min=1,max=128"`
	FloatRatioSnap   float64  `json:"float_ratio_snap" binding:"required,min=0,max=10"`
	SettlementPrice  float64  `json:"settlement_price" binding:"required,min=0"`
}

type supplierSettlementUpdateReq struct {
	ID               string   `json:"id" binding:"required,uuid4"`
	InquiryID        *string  `json:"inquiry_id" binding:"omitempty,uuid4"`
	ItemID           *string  `json:"item_id" binding:"omitempty,uuid4"`
	SupplierID       *string  `json:"supplier_id" binding:"omitempty,uuid4"`
	SupplierNameSnap *string  `json:"supplier_name_snap" binding:"omitempty,min=1,max=128"`
	FloatRatioSnap   *float64 `json:"float_ratio_snap" binding:"omitempty,min=0,max=10"`
	SettlementPrice  *float64 `json:"settlement_price" binding:"omitempty,min=0"`
}

func (h *SupplierSettlementHandler) create(c *gin.Context) {
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

	params := svc.CreateParams{
		InquiryID:        req.InquiryID,
		ItemID:           req.ItemID,
		SupplierID:       req.SupplierID,
		SupplierNameSnap: req.SupplierNameSnap,
		FloatRatioSnap:   req.FloatRatioSnap,
		SettlementPrice:  req.SettlementPrice,
	}
	settlement, err := h.s.CreateSupplierSettlement(c, params)
	if err != nil {
		ConflictError(c, errTitle, "创建供应商结算失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, settlement)
}

func (h *SupplierSettlementHandler) get(c *gin.Context) {
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
	c.JSON(http.StatusOK, settlement)
}

func (h *SupplierSettlementHandler) list(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *SupplierSettlementHandler) update(c *gin.Context) {
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

	params := svc.UpdateParams{
		ID:               req.ID,
		InquiryID:        req.InquiryID,
		ItemID:           req.ItemID,
		SupplierID:       req.SupplierID,
		SupplierNameSnap: req.SupplierNameSnap,
		FloatRatioSnap:   req.FloatRatioSnap,
		SettlementPrice:  req.SettlementPrice,
	}
	if err := h.s.UpdateSupplierSettlement(c, params); err != nil {
		ConflictError(c, errTitle, "更新供应商结算失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SupplierSettlementHandler) delete(c *gin.Context) {
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
	if err := h.s.DeleteSupplierSettlement(c, req.ID); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
