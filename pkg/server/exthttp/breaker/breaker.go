package breaker

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tal-tech/go-zero/core/breaker"
	"github.com/zcong1993/x/pkg/zero"
)

const breakerSeparator = "://"

func GinBreakerMiddleware(logger log.Logger) gin.HandlerFunc {
	zero.SetupMetrics()
	metrics := zero.Metrics

	var lock sync.Mutex
	breakerMap := make(map[string]breaker.Breaker, 0)

	var getBreaker = func(key string) breaker.Breaker {
		lock.Lock()
		defer lock.Unlock()
		if breaker, ok := breakerMap[key]; ok {
			return breaker
		}
		breakerMap[key] = breaker.NewBreaker(breaker.WithName(key))
		return breakerMap[key]
	}

	return func(c *gin.Context) {
		fullPath := c.FullPath()
		// 404
		if len(fullPath) == 0 {
			c.Next()
			return
		}
		key := strings.Join([]string{c.Request.Method, fullPath}, breakerSeparator)
		brk := getBreaker(key)

		// breaker logic
		promise, err := brk.Allow()
		if err != nil {
			metrics.AddDrop()
			level.Error(logger).Log("component", "breaker", "msg", "[http] dropped", "url", c.Request.URL.String(), "ip", c.ClientIP())
			//logx.Errorf("[http] dropped, %s - %s - %s",
			//	r.RequestURI, httpx.GetRemoteAddr(r), r.UserAgent())
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
