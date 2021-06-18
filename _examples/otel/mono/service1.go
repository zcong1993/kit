package main

import (
	"context"
	"net/http"
	"time"

	oteltracing "github.com/zcong1993/x/pkg/tracing/otel"

	"github.com/zcong1993/x/pkg/extapp"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/server/exthttp"
)

func service1Cmd(app *extapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "service1",
		Short: "sub command for service1",
		Run: func(cmd *cobra.Command, args []string) {
			app.InitFromCmd(cmd, "service1")

			// 真正的业务 http server
			// 初始化 gin
			r := app.GinServer(exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(":8081"))
			addRouters(r)

			extapp.FatalOnErrorf(app.Start(), "start error")
		},
	}
}

func addRouters(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		go func() {
			oteltracing.DoInSpan(c.Request.Context(), "bg work", func(ctx context.Context) {
				time.Sleep(time.Second)
			})
		}()
		time.Sleep(2 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})
}
