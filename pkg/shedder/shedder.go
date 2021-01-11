package shedder

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tal-tech/go-zero/core/load"
	"github.com/zcong1993/x/pkg/zero"
)

func GinShedderMiddleware(shedder load.Shedder, logger log.Logger) gin.HandlerFunc {
	// noop middleware
	if shedder == nil {
		level.Info(logger).Log("component", "shedder", "msg", "disable middleware")
		return func(c *gin.Context) {
			c.Next()
		}
	}

	level.Info(logger).Log("component", "shedder", "msg", "load middleware")

	zero.SetupMetrics()
	metrics := zero.Metrics
	sheddingStat := load.NewSheddingStat("api")

	return func(c *gin.Context) {
		sheddingStat.IncrementTotal()
		promise, err := shedder.Allow()
		if err != nil {
			metrics.AddDrop()
			sheddingStat.IncrementDrop()
			level.Error(logger).Log("component", "shedder", "msg", "[http] dropped", "url", c.Request.URL.String(), "ip", c.ClientIP())
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}

		c.Next()
		status := c.Writer.Status()
		if status == http.StatusInternalServerError {
			promise.Fail()
		} else {
			sheddingStat.IncrementPass()
			promise.Pass()
		}
	}
}
