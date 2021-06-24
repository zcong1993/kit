package main

import (
	"io"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/zcong1993/x/pkg/extapp"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/ginhelper"
	"github.com/zcong1993/x/pkg/server/exthttp"
)

func service2Cmd(app *extapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "service2",
		Short: "sub command for service2",
		Run: func(cmd *cobra.Command, args []string) {
			app.InitFromCmd(cmd, "service2")

			// 真正的业务 http server
			// 初始化 gin
			r := app.GinServer(exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(":8080"))
			addRoutersV2(r)

			extapp.FatalOnErrorf(app.Start(), "start error")
		},
	}
}

func addRoutersV2(r *gin.Engine) {
	r.GET("/", ginhelper.ErrorWrapper(func(c *gin.Context) error {
		resp, err := otelhttp.Get(c.Request.Context(), "http://localhost:8081")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		io.Copy(c.Writer, resp.Body)
		return nil
	}))
}
