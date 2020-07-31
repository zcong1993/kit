package ginhelper

import (
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func LoggerMw(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		logger = log.WithPrefix(logger, "component", "request", "path", path, "method", c.Request.Method, "statusCode", statusCode, "latency", latency, "clientIP", clientIP, "dataLength", dataLength, "userAgent", clientUserAgent)

		if len(c.Errors) > 0 {
			level.Error(logger).Log("message", c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			level.Info(logger).Log()
		}
	}
}

func DefaultWithLogger(logger log.Logger) *gin.Engine {
	g := gin.New()
	g.Use(LoggerMw(logger), gin.Recovery())
	return g
}
