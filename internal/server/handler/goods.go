package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	svc "hdzk.cn/foodapp/internal/service/goods"
	types "hdzk.cn/foodapp/internal/transport"
)

type GoodsHandler struct{ s *svc.Service }

func NewGoodsHandler(s *svc.Service) *GoodsHandler { return &GoodsHandler{s: s} }

func (h *GoodsHandler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/goods")

	g.POST("/create_goods", h.create)
	g.POST("/get_goods", h.get)
	g.POST("/list_goods", h.list)
	g.POST("/update_goods", h.update)
	g.POST("/soft_delete_goods", h.softDelete)
	g.POST("/hard_delete_goods", h.hardDelete)
}

type goodsCreateReq struct {
	Name               string  `json:"name" binding:"required,min=1,max=128"`
	OrgID              string  `json:"org_id" binding:"required,uuid4"`
	SpecID             string  `json:"spec_id" binding:"required,uuid4"`
	CategoryID         string  `json:"category_id" binding:"required,uuid4"`
	Sort               *int    `json:"sort" binding:"omitempty,min=0"`
	Code               *string `json:"code" binding:"min=1,max=64"`
	Pinyin             *string `json:"pinyin" binding:"omitempty,max=128"`
	ImageURL           *string `json:"image_url" binding:"omitempty,max=512"`
	AcceptanceStandard *string `json:"acceptance_standard" binding:"omitempty,max=512"`
}

type goodsUpdateReq struct {
	ID                 string  `json:"id" binding:"required,uuid4"`
	Name               *string `json:"name" binding:"omitempty,min=1,max=128"`
	Code               *string `json:"code" binding:"omitempty,min=1,max=64"`
	Sort               *int    `json:"sort" binding:"omitempty,min=0"`
	SpecID             *string `json:"spec_id" binding:"omitempty,uuid4"`
	CategoryID         *string `json:"category_id" binding:"omitempty,uuid4"`
	Pinyin             *string `json:"pinyin" binding:"omitempty,max=128"`
	ImageURL           *string `json:"image_url" binding:"omitempty,max=512"`
	AcceptanceStandard *string `json:"acceptance_standard" binding:"omitempty,max=512"`
}

func (h *GoodsHandler) create(c *gin.Context) {
	const errTitle = "创建商品失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可创建商品")
		return
	}

	var req goodsCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		BadRequest(c, errTitle, "name 或 code 不能为空")
		return
	}

	params := svc.CreateParams{
		Name:               req.Name,
		Code:               req.Code,
		OrgID:              req.OrgID,
		SpecID:             req.SpecID,
		CategoryID:         req.CategoryID,
		Sort:               req.Sort,
		Pinyin:             req.Pinyin,
		ImageURL:           req.ImageURL,
		AcceptanceStandard: req.AcceptanceStandard,
	}
	goods, err := h.s.CreateGoods(c, params)
	if err != nil {
		ConflictError(c, errTitle, "创建商品失败: "+err.Error())
		return
	}
	c.JSON(http.StatusCreated, goods)
}

func (h *GoodsHandler) get(c *gin.Context) {
	const errTitle = "获取商品失败"
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

	goods, err := h.s.GetGoods(c, req.ID)
	if err != nil {
		NotFoundError(c, errTitle, "商品不存在: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, goods)
}

func (h *GoodsHandler) list(c *gin.Context) {
	const errTitle = "获取商品列表失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}

	orgID := strings.TrimSpace(c.Query("org_id"))
	if orgID == "" {
		BadRequest(c, errTitle, "参数错误：缺少 org_id")
		return
	}

	categoryID := strings.TrimSpace(c.Query("category_id"))
	var categoryPtr *string
	if categoryID != "" {
		categoryPtr = &categoryID
	}
	specID := strings.TrimSpace(c.Query("spec_id"))
	var specPtr *string
	if specID != "" {
		specPtr = &specID
	}

	kw := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.s.ListGoods(c, kw, orgID, categoryPtr, specPtr, page, ps)
	if err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "items": list})
}

func (h *GoodsHandler) update(c *gin.Context) {
	const errTitle = "更新商品失败"

	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可更新商品")
		return
	}

	var req goodsUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}

	params := svc.UpdateParams{
		ID:                 req.ID,
		Name:               req.Name,
		Code:               req.Code,
		Sort:               req.Sort,
		SpecID:             req.SpecID,
		CategoryID:         req.CategoryID,
		Pinyin:             req.Pinyin,
		ImageURL:           req.ImageURL,
		AcceptanceStandard: req.AcceptanceStandard,
	}
	if err := h.s.UpdateGoods(c, params); err != nil {
		ConflictError(c, errTitle, "更新商品失败: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *GoodsHandler) softDelete(c *gin.Context) {
	const errTitle = "删除商品失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已停用，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除商品")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, err.Error())
		return
	}
	if err := h.s.SoftDeleteGoods(c, req.ID); err != nil {
		InternalError(c, errTitle, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *GoodsHandler) hardDelete(c *gin.Context) {
	const errTitle = "删除商品失败"
	act := middleware.GetActor(c)
	if act.Deleted != middleware.DeletedNo {
		ForbiddenError(c, errTitle, "账户已删除，禁止操作")
		return
	}
	if act.Role != middleware.RoleAdmin {
		ForbiddenError(c, errTitle, "仅管理员可删除商品")
		return
	}

	var req types.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, errTitle, "输入格式非法")
		return
	}
	if err := h.s.HardDeleteGoods(c, req.ID); err != nil {
		ConflictError(c, errTitle, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
