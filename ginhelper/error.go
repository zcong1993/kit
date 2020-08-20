package ginhelper

import (
	"net/http"
	"strconv"

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
	if err := http.StatusText(e.Status); err != "" {
		return err
	}
	return "E" + strconv.Itoa(e.Code)
}

func ReplyError(ctx *gin.Context, err error) {
	e, ok := err.(*BizError)
	if ok {
		ctx.JSON(e.Status, e)
		return
	}
	err1 := &BizError{
		Code:    500,
		Message: err.Error(),
	}
	ctx.JSON(500, err1)
}
