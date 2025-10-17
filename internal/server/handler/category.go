package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/category"
	types "hdzk.cn/foodapp/internal/transport"
)

type CategoryHandler struct{ s *svc.Service }

func NewCategoryHandler(s *svc.Service) *CategoryHandler { return &CategoryHandler{s: s} }

func (h *CategoryHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/category")

	g.POST("/create_category", h.CreateCategory)  // 新增品类
	g.POST("/get_category", h.GetCategory)        // 按 id 获取
	g.POST("/list_category", h.ListCategories)    // 列表（分页/条件）
	g.POST("/update_category", h.UpdateCategory)  // 更新品类
	g.POST("/udelete_category", h.DeleteCategory) // 删除品类
}

// 请求体
type category_createReq struct {
	Name   string  `json:"name" binding:"required,min=1,max=64"`
	Code   *string `json:"code" binding:"omitempty,max=64"`
	Pinyin *string `json:"pinyin" binding:"omitempty,max=64"`
}

type category_updateReq struct {
	ID     string  `json:"id"   binding:"required,uuid4"`
	Name   string  `json:"name" binding:"required,min=1,max=64"`
	Code   *string `json:"code" binding:"omitempty,max=64"`
	Pinyin *string `json:"pinyin" binding:"omitempty,max=64"`
}

// ---------- Category ----------
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req category_createReq
	err_title := "创建品类失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可新增品类")
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	m, err := h.s.CreateCategory(c, req.Name, req.Code, req.Pinyin)
	if err != nil {
		ConflictError(c, err_title, "添加品类失败:"+err.Error())
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *CategoryHandler) GetCategory(c *gin.Context) {
	var req types.IDReq
	err_title := "获取品类失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	
	m, err := h.s.GetCategory(c, req.ID)
	if err != nil {
		NotFoundError(c, err_title, "品类不存在:"+err.Error())
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *CategoryHandler) ListCategories(c *gin.Context) {
	kw := c.Query("keyword")

	err_title := "获取品类列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListCategories(c, kw, page, ps)
	if err != nil {
		InternalError(c, err_title, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	var req category_updateReq
	err_title := "更新品类失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可更新品类")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	if err := h.s.UpdateCategory(c, req.ID, req.Name, req.Code, req.Pinyin); err != nil {
		ConflictError(c, err_title, "更新品类失败:"+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	var req types.IDReq
	err_title := "删除品类失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, err_title, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, err_title, "仅管理员可删除品类")
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	if err := h.s.DeleteCategory(c, req.ID); err != nil {
		ConflictError(c, err_title, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
