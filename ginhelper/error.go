package ginhelper

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

type BizError struct {
	Status  int    `json:"-"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func NewBizError(status, code int, message string) *BizError {
	return &BizError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func (e *BizError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return http.StatusText(e.Status)
}

func ReplyError(ctx *gin.Context, err error) {
	e, ok := err.(*BizError)
	if ok {
		ctx.JSON(e.Status, e)
		return
	}

	// gin binding 校验错误
	ve, ok := err.(validator.ValidationErrors)
	if ok {
		// todo: 控制是否返回错误详情
		err := &BizError{
			Code:    400,
			Message: ve.Error(),
		}
		ctx.JSON(400, err)
		return
	}

	// 其他错误
	err1 := &BizError{
		Code:    500,
		Message: err.Error(),
	}
	ctx.JSON(500, err1)
}
