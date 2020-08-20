package ginhelper

import "github.com/gin-gonic/gin"

type Handler = func(c *gin.Context) error

func ErrorWrapper(h Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := h(c)
		if err != nil {
			ReplyError(c, err)
		}
	}
}
