package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domain "hdzk.cn/foodapp/internal/domain/organ"
	"hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/organ"
	types "hdzk.cn/foodapp/internal/transport"
)

type OrganHandler struct {
	s *svc.Service
}

func NewOrganHandler(s *svc.Service) *OrganHandler {
	return &OrganHandler{s: s}
}

func (h *OrganHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/organs")

	g.POST("/list", h.list)
	g.POST("/create", h.create)
	g.POST("/update", h.update)
	g.POST("/delete", h.delete)
	g.POST("/get", h.get)
}

type organCreateReq struct {
	Name   string  `json:"name" binding:"required,min=1,max=64"`
	Code   *string `json:"code" binding:"omitempty,max=32"`
	Leader string  `json:"leader" binding:"omitempty,max=64"`
	Phone  string  `json:"phone" binding:"omitempty,max=32"`
	Sort   int     `json:"sort" binding:"omitempty,gte=0"`
	Status int     `json:"status" binding:"omitempty,oneof=0 1"`
	Remark string  `json:"remark" binding:"omitempty,max=255"`
}

type organUpdateReq struct {
	ID string `json:"id" binding:"required,uuid4"`
	organCreateReq
}

type organListReq struct {
	Keyword string `json:"keyword" binding:"omitempty,max=64"`
	Status  *int   `json:"status" binding:"omitempty,oneof=0 1"`
	Limit   int    `json:"limit" binding:"omitempty,min=1,max=200"`
	Offset  int    `json:"offset" binding:"omitempty,min=0"`
}

func (h *OrganHandler) create(c *gin.Context) {
	const errTitle = "创建中队失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可新增中队")
		return
	}

	var req organCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "请求参数无效: "+err.Error())
		return
	}

	organ := &domain.Organ{
		Name:   req.Name,
		Code:   req.Code,
		Leader: req.Leader,
		Phone:  req.Phone,
		Sort:   req.Sort,
		Status: req.Status,
		Remark: req.Remark,
	}
	created, err := h.s.Create(c, organ)
	if err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *OrganHandler) update(c *gin.Context) {
	const errTitle = "更新中队失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新中队")
		return
	}

	var req organUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "请求参数无效: "+err.Error())
		return
	}

	organ := &domain.Organ{
		ID:     req.ID,
		Name:   req.Name,
		Code:   req.Code,
		Leader: req.Leader,
		Phone:  req.Phone,
		Sort:   req.Sort,
		Status: req.Status,
		Remark: req.Remark,
	}
	if err := h.s.Update(c, organ); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OrganHandler) delete(c *gin.Context) {
	const errTitle = "删除中队失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除中队")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "请求参数无效: "+err.Error())
		return
	}
	if err := h.s.Delete(c, req.ID); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OrganHandler) get(c *gin.Context) {
	const errTitle = "获取中队失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "请求参数无效: "+err.Error())
		return
	}

	organ, err := h.s.Get(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, organ)
}

func (h *OrganHandler) list(c *gin.Context) {
	const errTitle = "获取中队列表失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可查看中队列表")
		return
	}

	var req organListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "请求参数无效: "+err.Error())
		return
	}

	list, total, err := h.s.List(c, domain.ListQuery{
		Keyword: req.Keyword,
		Status:  req.Status,
		Limit:   req.Limit,
		Offset:  req.Offset,
	})
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": list,
	})
}
