// internal/server/handler/account.go
package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	domain "hdzk.cn/foodapp/internal/domain/account"
	"hdzk.cn/foodapp/internal/security"
	"hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/account"
	types "hdzk.cn/foodapp/internal/transport"
	"hdzk.cn/foodapp/pkg/crypto"
)

type AccountHandler struct{ s *svc.Service }

func NewAccountHandler(s *svc.Service) *AccountHandler { return &AccountHandler{s: s} }

// Register 同你原先一致（POST）
func (h *AccountHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/accounts")

	g.POST("/create", h.create)
	g.POST("/get", h.get)
	g.POST("/get_by_username", h.getByUsername)
	g.POST("/list", h.list)

	g.POST("/update_password", h.updatePassword)
	g.POST("/update", h.update)          // 修复为调用 Service.Update
	g.POST("/delete", h.softDelete)      // 调用 Service.SoftDelete
	g.POST("/hard_delete", h.hardDelete) // 调用 Service.HardDelete
	g.POST("/change_password", h.changePassword)
}

/************ 请求体 ************/
type acc_createReq struct {
	Username    string  `json:"username" binding:"required,min=1,max=64"`
	Password    string  `json:"password" binding:"required,max=128"`
	OrgID       string  `json:"org_id"  binding:"required,uuid4"` // 组织 ID 必填（uuid4）
	Role        int     `json:"role"`                             // 可选，默认 0
	Description *string `json:"description"`
}

type getByUsernameReq struct {
	Username string `json:"username" binding:"required"`
}

type acc_listReq struct {
	UsernameLike string `json:"username_like"`
	Deleted      *int   `json:"is_deleted"`
	Role         *int   `json:"role"`
	Limit        int    `json:"limit"  binding:"omitempty,min=1,max=200"`
	Offset       int    `json:"offset" binding:"omitempty,min=0"`
}

type updatePasswordReq struct {
	ID       string `json:"id"       binding:"required,uuid4"`
	Password string `json:"password" binding:"required,max=128"`
}

type acc_updateReq struct {
	ID          string  `json:"id"          binding:"required,uuid4"`
	Username    *string `json:"username"    binding:"omitempty,min=1,max=64"`
	OrgID       *string `json:"org_id"      binding:"omitempty,uuid4"`
	Description *string `json:"description"`
	Role        *int    `json:"role"` // 仅管理员可改；0=用户 1=管理员
}

type changePasswordReq struct {
	Username    string `json:"username"     binding:"required"`
	OldPassword string `json:"old_password" binding:"required,min=6,max=128"`
	NewPassword string `json:"new_password" binding:"required,max=128"`
}

/************ 处理函数 ************/
func (h *AccountHandler) create(c *gin.Context) {
	const errTitle = "创建用户失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可新增用户")
		return
	}

	var req acc_createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "请求参数无效: "+err.Error())
		return
	}
	// 密码复杂度
	if v := security.Validate(req.Password); len(v) > 0 {
		BadRequest(c, errTitle, strings.Join(v, "；"))
		return
	}
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		InternalError(c, errTitle, "hash 密码失败: "+err.Error())
		return
	}

	a := &domain.Account{
		Username:     req.Username,
		PasswordHash: hash,
		OrgID:        req.OrgID,
		Role:         req.Role,
		Description:  req.Description,
	}
	if err := h.s.Create(c, a); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": a.ID})
}

func (h *AccountHandler) get(c *gin.Context) {
	const errTitle = "获取用户失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if act.Role != middleware.RoleAdmin && act.ID != req.ID {
		ForbiddenError(c, errTitle, "仅可查询本人信息")
		return
	}

	a, err := h.s.GetByID(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AccountHandler) getByUsername(c *gin.Context) {
	const errTitle = "获取用户失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	var req getByUsernameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if act.Role != middleware.RoleAdmin && act.Username != req.Username {
		ForbiddenError(c, errTitle, "仅可查询本人信息")
		return
	}
	a, err := h.s.GetByUsername(c, req.Username)
	if err != nil {
		NotFoundError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AccountHandler) list(c *gin.Context) {
	const errTitle = "获取用户列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}

	if act.Role != middleware.RoleAdmin {
		a, err := h.s.GetByID(c, act.ID)
		if err != nil {
			InternalError(c, errTitle, "无法获取当前用户信息: "+err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"total": 1, "items": []any{a}})
		return
	}

	var req acc_listReq
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Limit, req.Offset = 20, 0
	}
	items, total, err := h.s.List(c, domain.ListQuery{
		UsernameLike: req.UsernameLike,
		Deleted:      req.Deleted,
		Role:         req.Role,
		Limit:        req.Limit,
		Offset:       req.Offset,
	})
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
}

func (h *AccountHandler) updatePassword(c *gin.Context) {
	const errTitle = "重置密码失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可重置他人密码")
		return
	}
	var req updatePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if act.ID == req.ID {
		BadRequest(c, errTitle, "修改本人密码请使用修改密码接口（需要旧密码）")
		return
	}
	if v := security.Validate(req.Password); len(v) > 0 {
		BadRequest(c, errTitle, strings.Join(v, "；"))
		return
	}
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		InternalError(c, errTitle, "hash 密码失败: "+err.Error())
		return
	}
	if err := h.s.UpdatePasswordHash(c, req.ID, hash); err != nil {
		InternalError(c, errTitle, "更新密码失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AccountHandler) update(c *gin.Context) {
	const errTitle = "更新用户信息失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}

	var req acc_updateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	// 非管理员：仅可改“本人”
	if act.Role != middleware.RoleAdmin && act.ID != req.ID {
		ForbiddenError(c, errTitle, "仅可修改本人信息")
		return
	}
	// 非管理员：不允许改 Role
	if act.Role != middleware.RoleAdmin {
		req.Role = nil
	}

	in := svc.UpdateInput{
		ID:          req.ID,
		Username:    req.Username,
		OrgID:       req.OrgID,
		Description: req.Description,
		Role:        req.Role,
	}
	if err := h.s.Update(c, in); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AccountHandler) softDelete(c *gin.Context) {
	const errTitle = "删除用户失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除用户")
		return
	}
	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if act.ID == req.ID {
		BadRequest(c, errTitle, "不允许删除自身账号")
		return
	}
	if target, err := h.s.GetByID(c, req.ID); err == nil && target != nil && target.Role == middleware.RoleAdmin {
		BadRequest(c, errTitle, "不允许删除管理员账户")
		return
	}
	if err := h.s.SoftDelete(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AccountHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除用户失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除用户")
		return
	}
	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if act.ID == req.ID {
		BadRequest(c, errTitle, "不允许删除自身账号")
		return
	}
	if target, err := h.s.GetByID(c, req.ID); err == nil && target != nil && target.Role == middleware.RoleAdmin {
		BadRequest(c, errTitle, "不允许删除管理员账户")
		return
	}
	if err := h.s.HardDelete(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AccountHandler) changePassword(c *gin.Context) {
	const errTitle = "修改密码失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if act.Username != req.Username {
		ForbiddenError(c, errTitle, "仅可修改本人密码")
		return
	}
	if v := security.Validate(req.NewPassword); len(v) > 0 {
		BadRequest(c, errTitle, strings.Join(v, "；"))
		return
	}
	if err := h.s.ChangePassword(c, req.Username, req.OldPassword, req.NewPassword); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
