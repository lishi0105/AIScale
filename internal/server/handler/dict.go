package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/dict"
	types "hdzk.cn/foodapp/internal/transport"
)

type DictHandler struct{ s *svc.Service }

func NewDictHandler(s *svc.Service) *DictHandler { return &DictHandler{s: s} }

func (h *DictHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/dict")

	g.POST("/create_unit", h.CreateUnit)  // 新增单位
	g.POST("/get_unit", h.GetUnit)        // 按 id 获取
	g.POST("/list_unit", h.ListUnits)     // 列表（分页/条件）
	g.POST("/update_unit", h.UpdateUnit)  // 更新单位
	g.POST("/udelete_unit", h.DeleteUnit) // 删除单位

	g.POST("/create_spec", h.CreateSpec)  // 新增规格
	g.POST("/get_spec", h.GetSpec)        // 按 id 获取
	g.POST("/list_spec", h.ListSpecs)     // 列表（分页/条件）
	g.POST("/update_spec", h.UpdateSpec)  // 更新规格
	g.POST("/udelete_spec", h.DeleteSpec) // 删除规格

	g.POST("/create_mealTime", h.CreateMealTime)  // 新增餐次
	g.POST("/get_mealTime", h.GetMealTime)        // 按 id 获取
	g.POST("/list_mealTime", h.ListMealTimes)     // 列表（分页/条件）
	g.POST("/update_mealTime", h.UpdateMealTime)  // 更新餐次
	g.POST("/udelete_mealTime", h.DeleteMealTime) // 删除规格
}

// 通用请求体
type dict_createReq struct {
	Name string  `json:"name" binding:"required,min=1,max=32"`
	Code *string `json:"code" binding:"omitempty,max=32"`
	Sort int     `json:"sort" binding:"gte=0"`
}

type dict_updateReq struct {
	ID   string  `json:"id"   binding:"required,uuid4"`
	Name string  `json:"name" binding:"required,min=1,max=32"`
	Code *string `json:"code" binding:"omitempty,max=32"`
	Sort int     `json:"sort" binding:"gte=0"`
}

// ---------- Unit ----------
func (h *DictHandler) CreateUnit(c *gin.Context) {
	var req dict_createReq
	err_title := "创建单位失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可新增单位")
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	m, err := h.s.CreateUnit(c, req.Name, req.Code, req.Sort)
	if err != nil {
		ConflictError(c, err_title, "添加单位失败:"+err.Error())
		return
	}
	SuccessResponse(c, m)
}
func (h *DictHandler) GetUnit(c *gin.Context) {
	var req types.IDReq
	err_title := "获取单位失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	m, err := h.s.GetUnit(c, req.ID)
	if err != nil {
		NotFoundError(c, err_title, "单位不存在:"+err.Error())
		return
	}
	SuccessResponse(c, m)
}
func (h *DictHandler) ListUnits(c *gin.Context) {
	kw := c.Query("keyword")

	err_title := "获取单位列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListUnits(c, kw, page, ps)
	if err != nil {
		InternalError(c, err_title, err.Error())
		return
	}
	SuccessResponse(c, map[string]any{"total": total, "items": list})
}

func (h *DictHandler) UpdateUnit(c *gin.Context) {
	var req dict_updateReq
	err_title := "更新单位失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可更新单位")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	if err := h.s.UpdateUnit(c, req.ID, req.Name, req.Code, req.Sort); err != nil {
		ConflictError(c, err_title, "更新单位失败:"+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *DictHandler) DeleteUnit(c *gin.Context) {
	var req types.IDReq
	err_title := "删除单位失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可删除单位")
		return
	}
	if err := h.s.DeleteUnit(c, req.ID); err != nil {
		ConflictError(c, err_title, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------- Spec ----------
func (h *DictHandler) CreateSpec(c *gin.Context) {
	var req dict_createReq
	err_title := "创建规格失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可新增规格")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	m, err := h.s.CreateSpec(c, req.Name, req.Code, req.Sort)
	if err != nil {
		ConflictError(c, err_title, "添加规格失败:"+err.Error())
		return
	}
	SuccessResponse(c, m)
}
func (h *DictHandler) GetSpec(c *gin.Context) {
	var req types.IDReq
	err_title := "获取规格失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	m, err := h.s.GetSpec(c, req.ID)
	if err != nil {
		NotFoundError(c, err_title, "规格不存在:"+err.Error())
		return
	}
	SuccessResponse(c, m)
}
func (h *DictHandler) ListSpecs(c *gin.Context) {
	kw := c.Query("keyword")
	err_title := "获取规格列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListSpecs(c, kw, page, ps)
	if err != nil {
		InternalError(c, err_title, err.Error())
		return
	}
	SuccessResponse(c, map[string]any{"total": total, "items": list})
}
func (h *DictHandler) UpdateSpec(c *gin.Context) {
	var req dict_updateReq
	err_title := "更新规格失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可编辑规格")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	if err := h.s.UpdateSpec(c, req.ID, req.Name, req.Code, req.Sort); err != nil {
		ConflictError(c, err_title, "更新规格失败:"+err.Error())
		return
	}
	SuccessResponse(c, nil)
}
func (h *DictHandler) DeleteSpec(c *gin.Context) {
	var req types.IDReq
	err_title := "删除规格失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可删除规格")
		return
	}
	if err := h.s.DeleteSpec(c, req.ID); err != nil {
		ConflictError(c, err_title, err.Error())
		return
	}
	SuccessResponse(c, nil)
}

// ---------- MealTime ----------
func (h *DictHandler) CreateMealTime(c *gin.Context) {
	var req dict_createReq
	err_title := "创建就餐时段失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可新增就餐时段")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	m, err := h.s.CreateMealTime(c, req.Name, req.Code, req.Sort)
	if err != nil {
		ConflictError(c, err_title, "添加就餐时段失败:"+err.Error())
		return
	}
	SuccessResponse(c, m)
}
func (h *DictHandler) GetMealTime(c *gin.Context) {
	var req types.IDReq
	err_title := "获取就餐时段失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	m, err := h.s.GetMealTime(c, req.ID)
	if err != nil {
		NotFoundError(c, err_title, "就餐时段不存在:"+err.Error())
		return
	}
	SuccessResponse(c, m)
}
func (h *DictHandler) ListMealTimes(c *gin.Context) {
	kw := c.Query("keyword")
	err_title := "获取就餐时段列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListMealTimes(c, kw, page, ps)
	if err != nil {
		InternalError(c, err_title, err.Error())
		return
	}
	SuccessResponse(c, map[string]any{"total": total, "items": list})
}
func (h *DictHandler) UpdateMealTime(c *gin.Context) {
	var req dict_updateReq
	err_title := "更新就餐时段失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可编辑就餐时段")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	if err := h.s.UpdateMealTime(c, req.ID, req.Name, req.Code, req.Sort); err != nil {
		ConflictError(c, err_title, "更新就餐时段失败:"+err.Error())
		return
	}
	SuccessResponse(c, nil)
}
func (h *DictHandler) DeleteMealTime(c *gin.Context) {
	var req types.IDReq
	err_title := "删除就餐时段失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可删除就餐时段")
		return
	}
	if err := h.s.DeleteMealTime(c, req.ID); err != nil {
		ConflictError(c, err_title, err.Error())
		return
	}
	SuccessResponse(c, nil)
}
