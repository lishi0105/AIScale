package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/supplier"
	types "hdzk.cn/foodapp/internal/transport"
)

type SupplierHandler struct{ s *svc.Service }

func NewSupplierHandler(s *svc.Service) *SupplierHandler { return &SupplierHandler{s: s} }

func (h *SupplierHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/supplier")

	g.POST("/create_supplier", h.CreateSupplier)    // 新增供货商
	g.POST("/get_supplier", h.GetSupplier)          // 按 id 获取
	g.POST("/list_supplier", h.ListSuppliers)       // 列表（分页/条件）
	g.POST("/update_supplier", h.UpdateSupplier)    // 更新供货商
	g.POST("/soft_delete_supplier", h.softDelete)   // 软删除供货商
	g.POST("/hard_delete_supplier", h.hardDelete)   // 硬删除供货商
}

// 请求体
type supplier_createReq struct {
	Name        string     `json:"name" binding:"required,min=1,max=128"`
	Code        *string    `json:"code" binding:"omitempty,max=64"`
	Pinyin      *string    `json:"pinyin" binding:"omitempty,max=64"`
	Status      *int       `json:"status" binding:"omitempty,oneof=1 2"`
	Description string     `json:"description" binding:"required"`
	FloatRatio  float64    `json:"float_ratio" binding:"required,gt=0"`
	OrgID       *string    `json:"org_id" binding:"omitempty,uuid4"`
	StartTime   *time.Time `json:"start_time" binding:"omitempty"`
	EndTime     *time.Time `json:"end_time" binding:"omitempty"`
}

type supplier_updateReq struct {
	ID          string     `json:"id" binding:"required,uuid4"`
	Name        string     `json:"name" binding:"required,min=1,max=128"`
	Code        *string    `json:"code" binding:"omitempty,max=64"`
	Pinyin      *string    `json:"pinyin" binding:"omitempty,max=64"`
	Sort        *int       `json:"sort" binding:"omitempty,min=0"`
	Status      *int       `json:"status" binding:"omitempty,oneof=1 2"`
	Description *string    `json:"description" binding:"omitempty"`
	FloatRatio  *float64   `json:"float_ratio" binding:"omitempty,gt=0"`
	OrgID       *string    `json:"org_id" binding:"omitempty,uuid4"`
	StartTime   *time.Time `json:"start_time" binding:"omitempty"`
	EndTime     *time.Time `json:"end_time" binding:"omitempty"`
}

// ---------- Supplier ----------
func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	var req supplier_createReq
	err_title := "创建供货商失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可新增供货商")
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	
	// 验证时间范围
	if req.StartTime != nil && req.EndTime != nil && req.StartTime.After(*req.EndTime) {
		BadRequest(c, err_title, "start_time 必须小于等于 end_time")
		return
	}
	
	m, err := h.s.CreateSupplier(c, req.Name, req.Code, req.Pinyin, req.Status, req.Description, req.FloatRatio, req.OrgID, req.StartTime, req.EndTime)
	if err != nil {
		ConflictError(c, err_title, "添加供货商失败:"+err.Error())
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *SupplierHandler) GetSupplier(c *gin.Context) {
	var req types.IDReq
	err_title := "获取供货商失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}

	m, err := h.s.GetSupplier(c, req.ID)
	if err != nil {
		NotFoundError(c, err_title, "供货商不存在:"+err.Error())
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	kw := c.Query("keyword")

	err_title := "获取供货商列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	
	// org_id 是可选的
	orgID := c.Query("org_id")
	var orgIDPtr *string
	if orgID != "" {
		orgIDPtr = &orgID
	}
	
	// status 是可选的
	statusStr := c.Query("status")
	var statusPtr *int
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			statusPtr = &status
		}
	}
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListSuppliers(c, kw, orgIDPtr, statusPtr, page, ps)
	if err != nil {
		InternalError(c, err_title, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	var req supplier_updateReq
	err_title := "更新供货商失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可更新供货商")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	
	// 验证时间范围
	if req.StartTime != nil && req.EndTime != nil && req.StartTime.After(*req.EndTime) {
		BadRequest(c, err_title, "start_time 必须小于等于 end_time")
		return
	}
	
	if err := h.s.UpdateSupplier(c, req.ID, req.Name, req.Code, req.Pinyin, req.Sort, req.Status, req.Description, req.FloatRatio, req.OrgID, req.StartTime, req.EndTime); err != nil {
		ConflictError(c, err_title, "更新供货商失败:"+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SupplierHandler) softDelete(c *gin.Context) {
	err_title := "删除供货商失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可删除供货商")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, err.Error())
		return
	}
	if err := h.s.SoftDeleteSupplier(c, req.ID); err != nil {
		InternalError(c, err_title, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *SupplierHandler) hardDelete(c *gin.Context) {
	var req types.IDReq
	err_title := "删除供货商失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可删除供货商")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	if err := h.s.HardDeleteSupplier(c, req.ID); err != nil {
		ConflictError(c, err_title, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
