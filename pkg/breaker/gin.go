package breaker

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/zcong1993/kit/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/zcong1993/kit/pkg/zero"
)

const breakerSeparator = "://"

// RegisterGinBreaker register breaker middleware based on command line parameters.
func RegisterGinBreaker(r *gin.Engine, logger *log.Logger, opt *Option) {
	logger = logger.With(log.Component("http/breaker"))
	if opt.disable {
		logger.Info("disable middleware")
		return
	}

	r.Use(GinBreakerMiddleware(logger))
}

// GinBreakerMiddleware create a gin middleware.
func GinBreakerMiddleware(logger *log.Logger) gin.HandlerFunc {
	zero.SetupMetrics()
	metrics := zero.Metrics
	brkGetter := NewBrkGetter()

	logger.Info("load middleware")

	sl := logger.Sugar()

	return func(c *gin.Context) {
		fullPath := c.FullPath()
		// 404.
		if len(fullPath) == 0 {
			c.Next()
			return
		}
		key := strings.Join([]string{c.Request.Method, fullPath}, breakerSeparator)
		brk := brkGetter.Get(key)

		// breaker logic.
		promise, err := brk.Allow()
		if err != nil {
			metrics.AddDrop()
			sl.Errorw("[http] dropped", "url", c.Request.URL.String(), "ip", c.ClientIP())
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		c.Next()
		status := c.Writer.Status()
		if status < http.StatusInternalServerError {
			promise.Accept()
		} else {
			promise.Reject(fmt.Sprintf("%d %s", status, http.StatusText(status)))
		}
	}
}
