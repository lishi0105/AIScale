package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	repo "hdzk.cn/foodapp/internal/repository/supplier"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/supplier"
	types "hdzk.cn/foodapp/internal/transport"
	"hdzk.cn/foodapp/pkg/utils"
)

type SupplierHandler struct{ s *svc.Service }

func NewSupplierHandler(s *svc.Service) *SupplierHandler { return &SupplierHandler{s: s} }

func (h *SupplierHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/supplier")

	g.POST("/create_supplier", h.create)
	g.POST("/get_supplier", h.get)
	g.POST("/list_supplier", h.list)
	g.POST("/update_supplier", h.update)
	g.POST("/soft_delete_supplier", h.softDelete)
	g.POST("/hard_delete_supplier", h.hardDelete)
}

type supplierCreateReq struct {
	Name        string  `json:"name" binding:"required,min=1,max=128"`
	OrgID       string  `json:"org_id" binding:"required,uuid4"`
	Description string  `json:"description" binding:"required"`
	FloatRatio  float64 `json:"float_ratio" binding:"required,gt=0"`
	Code        *string `json:"code" binding:"omitempty,max=64"`
	Pinyin      *string `json:"pinyin" binding:"omitempty,max=64"`
	Status      *int    `json:"status" binding:"omitempty,oneof=1 2"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
}

type supplierUpdateReq struct {
	ID          string   `json:"id" binding:"required,uuid4"`
	Name        *string  `json:"name" binding:"omitempty,min=1,max=128"`
	Code        *string  `json:"code" binding:"omitempty,max=64"`
	Pinyin      *string  `json:"pinyin" binding:"omitempty,max=64"`
	Sort        *int     `json:"sort" binding:"omitempty,min=0"`
	Status      *int     `json:"status" binding:"omitempty,oneof=1 2"`
	Description *string  `json:"description"`
	FloatRatio  *float64 `json:"float_ratio" binding:"omitempty,gt=0"`
	StartTime   *string  `json:"start_time"`
	EndTime     *string  `json:"end_time"`
}

func (h *SupplierHandler) create(c *gin.Context) {
	const errTitle = "创建供应商失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建供应商")
		return
	}

	var req supplierCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	startTime, err := parseOptionalTime(req.StartTime)
	if err != nil {
		BadRequest(c, errTitle, "start_time 格式非法，应为 YYYY-MM-DD HH:MM:SS")
		return
	}
	endTime, err := parseOptionalTime(req.EndTime)
	if err != nil {
		BadRequest(c, errTitle, "end_time 格式非法，应为 YYYY-MM-DD HH:MM:SS")
		return
	}

	params := svc.CreateParams{
		Name:        req.Name,
		OrgID:       req.OrgID,
		Description: req.Description,
		FloatRatio:  req.FloatRatio,
		Code:        req.Code,
		Pinyin:      req.Pinyin,
		Status:      req.Status,
		StartTime:   startTime,
		EndTime:     endTime,
	}
	supplier, err := h.s.CreateSupplier(c, params)
	if err != nil {
		ConflictError(c, errTitle, "创建供应商失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, supplier)
}

func (h *SupplierHandler) get(c *gin.Context) {
	const errTitle = "获取供应商失败"
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

	supplier, err := h.s.GetSupplier(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "供应商不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) list(c *gin.Context) {
	const errTitle = "获取供应商列表失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	orgID := c.Query("org_id")
	if strings.TrimSpace(orgID) == "" {
		BadRequest(c, errTitle, "参数错误：缺少 org_id")
		return
	}
	keyword := c.Query("keyword")
	statusPtr, err := utils.GetQueryIntPointer(c, "status")
	if err != nil {
		BadRequest(c, errTitle, "status 非法: "+err.Error())
		return
	}
	if statusPtr != nil && *statusPtr != 1 && *statusPtr != 2 {
		BadRequest(c, errTitle, "status 取值非法，只能为1或2")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	listParams := repo.ListParams{
		Keyword:  keyword,
		OrgID:    orgID,
		Status:   statusPtr,
		Page:     page,
		PageSize: pageSize,
	}

	items, total, err := h.s.ListSuppliers(c, listParams)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
}

func (h *SupplierHandler) update(c *gin.Context) {
	const errTitle = "更新供应商失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新供应商")
		return
	}

	var req supplierUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	startTime, updateStart, err := parseOptionalTimeWithFlag(req.StartTime)
	if err != nil {
		BadRequest(c, errTitle, "start_time 格式非法，应为 YYYY-MM-DD HH:MM:SS")
		return
	}
	endTime, updateEnd, err := parseOptionalTimeWithFlag(req.EndTime)
	if err != nil {
		BadRequest(c, errTitle, "end_time 格式非法，应为 YYYY-MM-DD HH:MM:SS")
		return
	}

	params := svc.UpdateParams{
		ID:              req.ID,
		Name:            req.Name,
		Code:            req.Code,
		Pinyin:          req.Pinyin,
		Sort:            req.Sort,
		Status:          req.Status,
		Description:     req.Description,
		FloatRatio:      req.FloatRatio,
		StartTime:       startTime,
		EndTime:         endTime,
		UpdateSort:      req.Sort != nil,
		UpdateStartTime: updateStart,
		UpdateEndTime:   updateEnd,
	}
	if err := h.s.UpdateSupplier(c, params); err != nil {
		ConflictError(c, errTitle, "更新供应商失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SupplierHandler) softDelete(c *gin.Context) {
	const errTitle = "删除供应商失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除供应商")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDeleteSupplier(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *SupplierHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除供应商失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除供应商")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDeleteSupplier(c, req.ID); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func parseOptionalTime(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", trimmed, time.Local)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func parseOptionalTimeWithFlag(raw *string) (*time.Time, bool, error) {
	if raw == nil {
		return nil, false, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, true, nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", trimmed, time.Local)
	if err != nil {
		return nil, true, err
	}
	return &t, true, nil
}
