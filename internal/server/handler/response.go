// response.go（或直接在 account_handler.go 顶部）

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Ok      bool        `json:"ok"`
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse 统一成功响应结构
type SuccessResp struct {
	Ok   bool        `json:"ok"`
	Data interface{} `json:"data,omitempty"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResp{Ok: true, Data: data})
}

// badRequest 返回 400 错误
func BadRequest(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg, Ok: false}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusBadRequest, resp)
}

// internalError 返回 500 错误
func InternalError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg, Ok: false}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusInternalServerError, resp)
}

func NotFoundError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg, Ok: false}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusNotFound, resp)
}

func UnauthorizedError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg, Ok: false}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusUnauthorized, resp)
}

func ConflictError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg, Ok: false}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusConflict, resp)
}

func ForbiddenError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg, Ok: false}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusForbidden, resp)
}
