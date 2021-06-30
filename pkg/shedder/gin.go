package shedder

import (
	"net/http"

	"github.com/zcong1993/kit/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/tal-tech/go-zero/core/load"
	"github.com/zcong1993/kit/pkg/zero"
)

func RegisterGinShedder(r *gin.Engine, shedder load.Shedder, logger *log.Logger) {
	logger = logger.With(log.Component("http/shedder"))
	if shedder == nil {
		logger.Info("disable middleware")
		return
	}

	r.Use(GinShedderMiddleware(shedder, logger))
}

func GinShedderMiddleware(shedder load.Shedder, logger *log.Logger) gin.HandlerFunc {
	// noop middleware
	if shedder == nil {
		logger.Error("shedder is nil")
		return func(c *gin.Context) {
			c.Next()
		}
	}

	logger.Info("load middleware")

	zero.SetupMetrics()
	metrics := zero.Metrics
	sheddingStat := load.NewSheddingStat("api")

	sl := logger.Sugar()

	return func(c *gin.Context) {
		sheddingStat.IncrementTotal()
		promise, err := shedder.Allow()
		if err != nil {
			metrics.AddDrop()
			sheddingStat.IncrementDrop()
			sl.Errorw("[http] dropped", "url", c.Request.URL.String(), "ip", c.ClientIP())
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
