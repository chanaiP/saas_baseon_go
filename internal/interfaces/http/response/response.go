package response

import "github.com/gin-gonic/gin"

const (
	CodeOK           = 0
	CodeBadRequest   = 40000
	CodeUnauthorized = 40100
	CodeForbidden    = 40300
	CodeNotFound     = 40400
	CodeInternal     = 50000
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Body{Code: CodeOK, Message: "ok", Data: data})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Body{Code: code, Message: message})
}
