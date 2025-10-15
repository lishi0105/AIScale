package handler

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	svc "hdzk.cn/foodapp/internal/service/account"
	"hdzk.cn/foodapp/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	s       *svc.Service
	secret  string
	ttlMins int
}

func NewAuthHandler(s *svc.Service, secret string, ttlMins int) *AuthHandler {
	return &AuthHandler{s: s, secret: secret, ttlMins: ttlMins}
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/auth/login", h.login) // 登录无需鉴权
}

func (h *AuthHandler) login(c *gin.Context) {
	var req loginReq
	err_title := "登录失败"
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Error(err_title, zap.String("请求参数无效: ", err.Error()))
		BadRequest(c, err_title, "输入格式非法")
		return
	}
	uid, role, err := h.s.Authenticate(c, req.Username, req.Password)
	if err != nil {
		UnauthorizedError(c, err_title, "用户名或密码错误")
		return
	}

	// 生成 JWT
	now := time.Now()
	exp := now.Add(time.Duration(h.ttlMins) * time.Minute)
	claims := jwt.MapClaims{
		"sub":  uid,
		"usr":  req.Username,
		"role": role,
		"iat":  now.Unix(),
		"exp":  exp.Unix(),
		"iss":  "foodapp",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := tok.SignedString([]byte(h.secret))
	if err != nil {
		BadRequest(c, err_title, "生成token失败"+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":      ss,
		"token_type": "Bearer",
		"expires_in": int(exp.Sub(now).Seconds()),
	})
}
