package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"

	"github.com/zcong1993/kit/pkg/extapp"

	"github.com/zcong1993/kit/pkg/server/exthttp"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/zcong1993/kit/pkg/ginhelper"
)

type Input struct {
	Name string `json:"name" binding:"required"`
}

// export OTEL_EXPORTER_TYPE=jaege
// export OTEL_EXPORTER_JAEGER_ENDPOINT=http://localhost:14268/api/traces
func main() {
	app := extapp.NewApp()

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			app.InitFromCmd(cmd, "gin")

			// 真正的业务 http server
			ginServer := app.GinServer(exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(":8080"))
			addRouters(ginServer)

			// 启动内部 http 服务, 健康检查路由, 监控指标路由, pprof
			m := app.GetInnerHTTPServer()
			// 可以在增加额外路由
			m.HandleFunc("/xxx", func(writer http.ResponseWriter, request *http.Request) {
				writer.Write([]byte("ok"))
			})

			extapp.FatalOnErrorf(app.Start(), "start error")
		},
	}

	app.Run(cmd)
}

func addRouters(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		tracer := otel.Tracer("gin")
		go func() {
			_, span := tracer.Start(c.Request.Context(), "bg work")
			defer span.End()
			time.Sleep(time.Second)
		}()
		time.Sleep(2 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	r.GET("/pets/:id", func(c *gin.Context) {
		resp, err := otelhttp.Get(c.Request.Context(), "http://localhost:8080/")
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		defer resp.Body.Close()
		io.Copy(c.Writer, resp.Body)
	})

	r.GET("/test", ginhelper.ErrorWrapper(func(c *gin.Context) error {
		q := c.Query("q")
		if q == "" {
			return ginhelper.NewBizError(400, 400, "query q is required")
		}
		c.JSON(200, gin.H{"success": true})
		return nil
	}))

	r.POST("/p", ginhelper.ErrorWrapper(func(c *gin.Context) error {
		var input Input
		err := c.ShouldBind(&input)
		if err != nil {
			fmt.Printf("%+v\n", err)
			return err
		}
		c.JSON(200, &input)
		return nil
	}))
}
