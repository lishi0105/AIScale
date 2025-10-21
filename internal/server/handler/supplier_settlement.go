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

func NewSupplierSettlementHandler(s *svc.Service) *SupplierSettlementHandler { return &SupplierSettlementHandler{s: s} }

func (h *SupplierSettlementHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/supplier_settlement")

	g.POST("/create_supplier_settlement", h.create)
	g.POST("/get_supplier_settlement", h.get)
	g.POST("/list_supplier_settlements", h.list)
	g.POST("/update_supplier_settlement", h.update)
	g.POST("/soft_delete_supplier_settlement", h.softDelete)
	g.POST("/hard_delete_supplier_settlement", h.hardDelete)
	g.POST("/batch_create_supplier_settlements", h.batchCreate)
	g.POST("/get_by_item_and_supplier", h.getByItemAndSupplier)
}

type supplierSettlementCreateReq struct {
	InquiryID        string   `json:"inquiry_id" binding:"required,uuid4"`
	ItemID           string   `json:"item_id" binding:"required,uuid4"`
	SupplierID       *string  `json:"supplier_id" binding:"omitempty,uuid4"`
	SupplierNameSnap string   `json:"supplier_name_snap" binding:"required,min=1,max=128"`
	FloatRatioSnap   float64  `json:"float_ratio_snap" binding:"required,min=0,max=1"`
	SettlementPrice  float64  `json:"settlement_price" binding:"required,min=0"`
}

type supplierSettlementUpdateReq struct {
	ID               string   `json:"id" binding:"required,uuid4"`
	SupplierID       *string  `json:"supplier_id" binding:"omitempty,uuid4"`
	SupplierNameSnap *string  `json:"supplier_name_snap" binding:"omitempty,min=1,max=128"`
	FloatRatioSnap   *float64 `json:"float_ratio_snap" binding:"omitempty,min=0,max=1"`
	SettlementPrice  *float64 `json:"settlement_price" binding:"omitempty,min=0"`
}

type batchCreateSupplierSettlementsReq struct {
	Settlements []supplierSettlementCreateReq `json:"settlements" binding:"required,min=1"`
}

type getByItemAndSupplierReq struct {
	ItemID       string `json:"item_id" binding:"required,uuid4"`
	SupplierName string `json:"supplier_name" binding:"required,min=1,max=128"`
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

	if strings.TrimSpace(req.SupplierNameSnap) == "" {
		BadRequest(c, errTitle, "supplier_name_snap 不能为空")
		return
	}
	if req.FloatRatioSnap <= 0 || req.FloatRatioSnap > 1 {
		BadRequest(c, errTitle, "float_ratio_snap 必须在 0-1 之间")
		return
	}
	if req.SettlementPrice < 0 {
		BadRequest(c, errTitle, "settlement_price 不能为负数")
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
	supplierSettlement, err := h.s.CreateSupplierSettlement(c, params)
	if err != nil {
		ConflictError(c, errTitle, "创建供应商结算失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, supplierSettlement)
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

	supplierSettlement, err := h.s.GetSupplierSettlement(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "供应商结算不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, supplierSettlement)
}

func (h *SupplierSettlementHandler) list(c *gin.Context) {
	const errTitle = "获取供应商结算列表失败"
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
	supplierID := strings.TrimSpace(c.Query("supplier_id"))
	var supplierPtr *string
	if supplierID != "" {
		supplierPtr = &supplierID
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	params := svc.ListParams{
		InquiryID:  inquiryID,
		ItemID:     itemPtr,
		SupplierID: supplierPtr,
		Page:       page,
		PageSize:   ps,
	}
	list, total, err := h.s.ListSupplierSettlements(c, params)
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

func (h *SupplierSettlementHandler) softDelete(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *SupplierSettlementHandler) hardDelete(c *gin.Context) {
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
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SupplierSettlementHandler) batchCreate(c *gin.Context) {
	const errTitle = "批量创建供应商结算失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可批量创建供应商结算")
		return
	}

	var req batchCreateSupplierSettlementsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	settlements := make([]svc.CreateParams, 0, len(req.Settlements))
	for _, settlement := range req.Settlements {
		if strings.TrimSpace(settlement.SupplierNameSnap) == "" {
			BadRequest(c, errTitle, "supplier_name_snap 不能为空")
			return
		}
		if settlement.FloatRatioSnap <= 0 || settlement.FloatRatioSnap > 1 {
			BadRequest(c, errTitle, "float_ratio_snap 必须在 0-1 之间")
			return
		}
		if settlement.SettlementPrice < 0 {
			BadRequest(c, errTitle, "settlement_price 不能为负数")
			return
		}
		settlements = append(settlements, svc.CreateParams{
			InquiryID:        settlement.InquiryID,
			ItemID:           settlement.ItemID,
			SupplierID:       settlement.SupplierID,
			SupplierNameSnap: settlement.SupplierNameSnap,
			FloatRatioSnap:   settlement.FloatRatioSnap,
			SettlementPrice:  settlement.SettlementPrice,
		})
	}

	params := svc.BatchCreateParams{Settlements: settlements}
	if err := h.s.BatchCreateSupplierSettlements(c, params); err != nil {
		ConflictError(c, errTitle, "批量创建供应商结算失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true, "count": len(settlements)})
}

func (h *SupplierSettlementHandler) getByItemAndSupplier(c *gin.Context) {
	const errTitle = "获取供应商结算失败"
	var req getByItemAndSupplierReq

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	supplierSettlement, err := h.s.GetByItemAndSupplier(c, req.ItemID, req.SupplierName)
	if err != nil {
		NotFoundError(c, errTitle, "供应商结算不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, supplierSettlement)
}