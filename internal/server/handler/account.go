package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	domain "hdzk.cn/foodapp/internal/domain/account"
	"hdzk.cn/foodapp/internal/security"
	"hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/account"
	types "hdzk.cn/foodapp/internal/transport"
	"hdzk.cn/foodapp/pkg/crypto"
	"hdzk.cn/foodapp/pkg/logger"
)

type AccountHandler struct{ s *svc.Service }

func NewAccountHandler(s *svc.Service) *AccountHandler { return &AccountHandler{s: s} }

// Register 统一注册 POST 接口（均为 POST）
func (h *AccountHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/accounts")

	g.POST("/create", h.create)                 // 仅管理员；禁止创建管理员
	g.POST("/get", h.get)                       // 可按需收紧为仅本人/管理员
	g.POST("/get_by_username", h.getByUsername) // 本人或管理员
	g.POST("/list", h.list)                     // 可按需收紧为仅管理员

	g.POST("/update_password", h.updatePassword) // 仅管理员重置他人（不可给自己）
	g.POST("/update_status", h.updateStatus)     // 仅管理员
	g.POST("/delete", h.softDelete)              // 仅管理员；不可删自己&管理员
	g.POST("/hard_delete", h.hardDelete)         // 同上
	g.POST("/change_password", h.changePassword) // 本人；需旧密码
}

// ---------- 请求体 ----------
type acc_createReq struct {
	Username string `json:"username" binding:"required,min=1,max=64"`
	Password string `json:"password" binding:"required,max=128"`
	Status   int    `json:"status"` // 默认启用（1）
	Role     int    `json:"role"`   // 默认普通用户（1）；禁止 0（管理员）
}
type getByUsernameReq struct {
	Username string `json:"username" binding:"required"`
}
type acc_listReq struct {
	UsernameLike string `json:"username_like"`
	Status       *int   `json:"status"`
	Role         *int   `json:"role"`
	Limit        int    `json:"limit"  binding:"omitempty,min=1,max=200"`
	Offset       int    `json:"offset" binding:"omitempty,min=0"`
}
type updatePasswordReq struct {
	ID       string `json:"id"       binding:"required,uuid4"`
	Password string `json:"password" binding:"required,max=128"`
}
type updateStatusReq struct {
	ID     string `json:"id"     binding:"required,uuid4"`
	Status *int   `json:"status" binding:"required"`
}
type changePasswordReq struct {
	Username    string `json:"username"     binding:"required"`
	OldPassword string `json:"old_password" binding:"required,min=6,max=128"`
	NewPassword string `json:"new_password" binding:"required,max=128"`
}

// ---------- 处理函数 ----------
func (h *AccountHandler) create(c *gin.Context) {
	const errTitle = "创建用户失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
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
	// 保底：禁止创建管理员账户
	role := req.Role
	if role == middleware.RoleAdmin {
		BadRequest(c, errTitle, "禁止创建管理员账户")
		return
	}
	if role != middleware.RoleUser {
		role = middleware.RoleUser
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
	status := req.Status
	if status != middleware.StatusDisabled && status != middleware.StatusEnabled {
		status = middleware.StatusEnabled
	}
	a := &domain.Account{
		Username:     req.Username,
		PasswordHash: hash,
		Status:       status,
		Role:         role,
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
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}

	// 非管理员：仅允许获取自己的信息
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
	const errTitle = "用户名获取用户失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	var req getByUsernameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	// 本人或管理员
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
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}

	// 非管理员：只返回“本人”
	if act.Role != middleware.RoleAdmin {
		// 直接用 ID 精确查，避免 UsernameLike 命中其他人
		a, err := h.s.GetByID(c, act.ID)
		if err != nil {
			// 理论上不应发生：token 中的 uid 查不到
			InternalError(c, errTitle, "无法获取当前用户信息: "+err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"total": 1,
			"items": []any{a},
		})
		return
	}

	// 管理员：保留原有分页/筛选逻辑
	var req acc_listReq
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Limit, req.Offset = 20, 0
	}
	items, total, err := h.s.List(c, domain.ListQuery{
		UsernameLike: req.UsernameLike,
		Status:       req.Status,
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
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	// 仅管理员
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可重置他人密码")
		return
	}
	var req updatePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	// 禁止管理员用此接口改“自己”的密码
	if act.ID == req.ID {
		BadRequest(c, errTitle, "修改本人密码请使用修改密码接口（需要旧密码）")
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
	if err := h.s.UpdatePasswordHash(c, req.ID, hash); err != nil {
		InternalError(c, errTitle, "更新密码失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AccountHandler) updateStatus(c *gin.Context) {
	const errTitle = "更新用户状态失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	// 仅管理员
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新用户状态")
		return
	}
	var req updateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Error(errTitle,
			zap.Error(err),
		)
		BadRequest(c, errTitle, "输入格式非法"+err.Error())
		return
	}
	status := *req.Status
	// 如需禁止把自己停用，打开：
	if act.ID == req.ID && status == middleware.StatusDisabled {
		BadRequest(c, errTitle, "不允许停用自身账号")
		return
	}
	if err := h.s.UpdateStatus(c, req.ID, status); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AccountHandler) softDelete(c *gin.Context) {
	const errTitle = "删除用户失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	// 仅管理员
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除用户")
		return
	}
	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	// 禁止删除自身
	if act.ID == req.ID {
		BadRequest(c, errTitle, "不允许删除自身账号")
		return
	}
	// 禁止删除管理员（双保险）
	if target, err := h.s.GetByID(c, req.ID); err == nil && target != nil && target.Role == middleware.RoleAdmin {
		BadRequest(c, errTitle, "不允许删除管理员账户")
		return
	}
	if err := h.s.SoftDeleteByID(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AccountHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除用户失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	// 仅管理员
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除用户")
		return
	}
	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	// 禁止删除自身
	if act.ID == req.ID {
		BadRequest(c, errTitle, "不允许删除自身账号")
		return
	}
	// 禁止删除管理员（双保险）
	if target, err := h.s.GetByID(c, req.ID); err == nil && target != nil && target.Role == middleware.RoleAdmin {
		BadRequest(c, errTitle, "不允许删除管理员账户")
		return
	}
	if err := h.s.HardDeleteByID(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AccountHandler) changePassword(c *gin.Context) {
	const errTitle = "修改密码失败"
	act := middleware.GetActor(c)
	if act.Status != middleware.StatusEnabled {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	// 仅本人
	if act.Username != req.Username {
		ForbiddenError(c, errTitle, "仅可修改本人密码")
		return
	}
	// 新密码复杂度
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
