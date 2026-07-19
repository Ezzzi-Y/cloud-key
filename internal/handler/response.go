package handler

import (
	"CloudKey/internal/errcode"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessPaginated(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message, Data: nil})
}

func BadRequest(c *gin.Context, code int, message string) {
	Error(c, http.StatusBadRequest, code, message)
}

func Unauthorized(c *gin.Context, code int, message string) {
	Error(c, http.StatusUnauthorized, code, message)
}

func NotFound(c *gin.Context, code int, message string) {
	Error(c, http.StatusNotFound, code, message)
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, errcode.CodeInternalError, errcode.GetMessage(errcode.CodeInternalError))
}
