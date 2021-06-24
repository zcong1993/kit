package breaker

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/zcong1993/x/pkg/zero"
)

const breakerSeparator = "://"

func RegisterGinBreaker(r *gin.Engine, logger log.Logger, opt *Option) {
	if opt.disable {
		level.Info(logger).Log("component", "http/breaker", "msg", "disable middleware")
		return
	}

	r.Use(GinBreakerMiddleware(logger))
}

func GinBreakerMiddleware(logger log.Logger) gin.HandlerFunc {
	zero.SetupMetrics()
	metrics := zero.Metrics
	brkGetter := NewBrkGetter()

	level.Info(logger).Log("component", "http/breaker", "msg", "load middleware")

	return func(c *gin.Context) {
		fullPath := c.FullPath()
		// 404
		if len(fullPath) == 0 {
			c.Next()
			return
		}
		key := strings.Join([]string{c.Request.Method, fullPath}, breakerSeparator)
		brk := brkGetter.Get(key)

		// breaker logic
		promise, err := brk.Allow()
		if err != nil {
			metrics.AddDrop()
			level.Error(logger).Log("component", "http/breaker", "msg", "[http] dropped", "url", c.Request.URL.String(), "ip", c.ClientIP())
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
