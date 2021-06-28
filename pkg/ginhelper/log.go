package ginhelper

import (
	"math"
	"time"

	"github.com/zcong1993/x/pkg/log"

	"github.com/gin-gonic/gin"
)

func LoggerMw(logger *log.Logger) gin.HandlerFunc {
	sl := logger.Sugar()
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		if len(c.Errors) > 0 {
			sl.Errorw(c.Errors.ByType(gin.ErrorTypePrivate).String(), "path", path, "method", c.Request.Method, "statusCode", statusCode, "latency", latency, "clientIP", clientIP, "dataLength", dataLength)
		} else {
			sl.Infow("", "path", path, "method", c.Request.Method, "statusCode", statusCode, "latency", latency, "clientIP", clientIP, "dataLength", dataLength)
		}
	}
}

func DefaultWithLogger(logger *log.Logger) *gin.Engine {
	g := gin.New()
	g.Use(LoggerMw(logger.With(log.Component("requestLog"))), gin.Recovery())
	return g
}
