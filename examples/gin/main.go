package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcong1993/x/ginhelper"
)

func main() {
	r, _ := ginhelper.DefaultServer()
	r.GET("/", func(c *gin.Context) {
		time.Sleep(5 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	r.GET("/test", ginhelper.ErrorWrapper(func(c *gin.Context) error {
		q := c.Query("q")
		if q == "" {
			return ginhelper.NewBizError(400, 400, "query q is required")
		}
		c.JSON(200, gin.H{"success": true})
		return nil
	}))

	ginhelper.GracefulShutdown(r, ":8080", time.Second*5, func() {
		println("on shutdown")
	})
}
