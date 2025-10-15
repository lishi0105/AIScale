// response.go（或直接在 account_handler.go 顶部）

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

// badRequest 返回 400 错误
func BadRequest(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusBadRequest, resp)
}

// internalError 返回 500 错误
func InternalError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusInternalServerError, resp)
}

func NotFoundError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusNotFound, resp)
}

func UnauthorizedError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusUnauthorized, resp)
}

func ConflictError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusConflict, resp)
}

func ForbiddenError(c *gin.Context, msg string, details ...interface{}) {
	resp := ErrorResponse{Error: msg}
	if len(details) > 0 {
		resp.Details = details[0]
	}
	c.JSON(http.StatusForbidden, resp)
}
