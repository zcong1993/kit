package ginhelper

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BizError struct {
	Status  int    `json:"-"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Err     error  `json:"-"`
}

func NewBizError(status, code int, message string) *BizError {
	return &BizError{
		Status:  status,
		Code:    code,
		Message: message,
		Err:     errors.New(message),
	}
}

func NewFromError(status, code int, err error) *BizError {
	if err == nil {
		return nil
	}
	return &BizError{
		Status:  status,
		Code:    code,
		Message: err.Error(),
		Err:     err,
	}
}

func (e *BizError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return http.StatusText(e.Status)
}

func (e *BizError) Unwrap() error {
	return e.Err
}

func ReplyError(ctx *gin.Context, err error) {
	// 业务错误
	var bizError *BizError
	if errors.As(err, &bizError) {
		ctx.JSON(bizError.Status, bizError)
		return
	}

	// gin binding 校验错误
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
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
