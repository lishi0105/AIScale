package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domain "hdzk.cn/foodapp/internal/domain/organ"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/organ"
	types "hdzk.cn/foodapp/internal/transport"
)

/************* Handler *************/
type OrganHandler struct{ s *svc.Service }

func NewOrganHandler(s *svc.Service) *OrganHandler { return &OrganHandler{s: s} }

// Register 统一注册（均为 POST）
func (h *OrganHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/orgs")

	g.POST("/create_organ", h.create)          // 仅管理员
	g.POST("/get_organ", h.get)                // 所有人可查
	g.POST("/list_organ", h.list)              // 所有人可查
	g.POST("/update_organ", h.update)          // 仅管理员
	g.POST("/soft_delete_organ", h.softDelete) // 仅管理员（软删）
	g.POST("/hard_delete_organ", h.hardDelete) // 仅管理员（硬删）
}

/************* 请求体 *************/
type orgCreateReq struct {
	Name        string  `json:"name"        binding:"required,min=1,max=64"`
	ParentID    *string `json:"parent_id"`   // 为空串表示根
	Code        *string `json:"code"`        // 可空；空串将被置为 NULL
	Description *string `json:"description"` // 可空
	Sort        *int    `json:"sort"`        // 可空
}

type orgUpdateReq struct {
	ID          string  `json:"id"          binding:"required,uuid4"`
	Name        *string `json:"name"        binding:"min=1,max=64"`
	ParentID    *string `json:"parent_id"`
	Code        *string `json:"code"` // 若传入空串，将置为 NULL
	Description *string `json:"description"`
}

type orgListReq struct {
	NameLike string `json:"name_like"`
	Deleted  *int   `json:"is_deleted"`
	Limit    int    `json:"limit"  binding:"omitempty,min=1,max=200"`
	Offset   int    `json:"offset" binding:"omitempty,min=0"`
}

/************* 处理函数 *************/

func (h *OrganHandler) create(c *gin.Context) {
	const errTitle = "创建组织失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建组织")
		return
	}

	var req orgCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "请求参数无效: "+err.Error())
		return
	}

	m := &domain.Organ{
		// ID 由 BeforeCreate 生成
		Name: req.Name,
		// Parent: 若为 nil 不变；若为 "" 作为根；字段是 NOT NULL，空串合法
	}
	if req.ParentID != nil {
		m.ParentID = req.ParentID
	}
	if req.Description != nil {
		m.Description = *req.Description
	}
	// Code：空串转 NULL（以便触发自动编码）
	if req.Code != nil {
		if *req.Code == "" {
			m.Code = nil
		} else {
			m.Code = req.Code
		}
	}
	if req.Sort != nil {
		m.Sort = *req.Sort
	}

	if err := h.s.Create(c, m); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": m.ID})
}

func (h *OrganHandler) get(c *gin.Context) {
	const errTitle = "获取组织失败"
	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	o, err := h.s.GetByID(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *OrganHandler) list(c *gin.Context) {
	const errTitle = "获取组织列表失败"
	var req orgListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		// 容错默认
		req.Limit, req.Offset = 20, 0
	}

	items, total, err := h.s.List(c, domain.ListQuery{
		NameLike: req.NameLike,
		Deleted:  req.Deleted,
		Limit:    req.Limit,
		Offset:   req.Offset,
	})
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
}

func (h *OrganHandler) update(c *gin.Context) {
	const errTitle = "更新组织失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新组织")
		return
	}

	var req orgUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "请求参数无效: "+err.Error())
		return
	}

	// 组装要更新的模型（你的 Service.Update 要求 m.Name 非空）
	update_m := svc.UpdateInput{}

	update_m.ID = req.ID
	if req.Name != nil {
		update_m.Name = req.Name
	}
	if req.ParentID != nil {
		update_m.Parent = req.ParentID
	}
	if req.Description != nil {
		update_m.Description = req.Description
	}
	if req.Code != nil {
		update_m.Code = req.Code
	}

	if err := h.s.Update(c, update_m); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	// 返回最新值
	obj, err := h.s.GetByID(c, req.ID)
	if err != nil {
		InternalError(c, errTitle, "获取更新后数据失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, obj)
}

func (h *OrganHandler) softDelete(c *gin.Context) {
	const errTitle = "删除组织失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除组织")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDelete(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *OrganHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除组织失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除组织")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "请求参数无效: "+err.Error())
		return
	}
	if err := h.s.HardDelete(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
