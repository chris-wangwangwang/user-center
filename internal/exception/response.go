package exception

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Code    Code        `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Code: CodeSuccess, Message: "ok", Data: data})
}

func Fail(c *gin.Context, ex *Exception) {
	c.JSON(httpStatusOf(ex.Code), APIResponse{Code: ex.Code, Message: ex.Message})
}

func FromError(err error) *Exception {
	var ex *Exception
	if errors.As(err, &ex) {
		return ex
	}
	return NewInternal(err)
}

func httpStatusOf(code Code) int {
	switch code {
	case CodeBadRequest, CodeValidation:
		return http.StatusBadRequest
	case CodeUnauthorized, CodeInvalidToken, CodeExpiredToken, CodeInvalidPassword:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
