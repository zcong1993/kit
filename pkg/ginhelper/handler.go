package ginhelper

import "github.com/gin-gonic/gin"

// Handler is same as gin handler but can return error.
type Handler = func(c *gin.Context) error

// ErrorWrapper can convert our Handler to gin.HandlerFunc with our error handler.
func ErrorWrapper(h Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := h(c)
		if err != nil {
			ReplyError(c, err)
		}
	}
}
