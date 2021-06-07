package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	oteltracing "github.com/zcong1993/x/pkg/tracing/otel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"

	"github.com/zcong1993/x/pkg/extapp"

	"github.com/zcong1993/x/pkg/shedder"

	"github.com/zcong1993/x/pkg/breaker"

	"github.com/zcong1993/x/pkg/metrics"

	"github.com/zcong1993/x/pkg/prober"

	"github.com/zcong1993/x/pkg/server/exthttp"

	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/extrun"

	"github.com/gin-gonic/gin"
	"github.com/zcong1993/x/pkg/ginhelper"
)

type Input struct {
	Name string `json:"name" binding:"required"`
}

// export OTEL_EXPORTER_TYPE=jaege
// export OTEL_EXPORTER_JAEGER_ENDPOINT=http://localhost:14268/api/traces
func main() {
	app := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			extApp := extapp.NewFromCmd(cmd)

			logger := extApp.Logger
			reg := extApp.Reg
			g := extApp.G
			app := extApp.App

			extapp.FatalOnErrorf(oteltracing.InitTracerFromEnv(app), "init tracer error")

			// 服务健康状态
			httpProber := prober.NewHTTP()
			statusProber := prober.Combine(httpProber, prober.NewInstrumentation(app, logger, reg))

			// 监听退出信号
			extrun.HandleSignal(g)

			// 真正的业务 http server
			// 初始化 gin
			r := ginhelper.DefaultWithLogger(logger)
			r.Use(metrics.NewInstrumentationMiddleware(reg))
			r.Use(oteltracing.GinMiddleware(app))
			// shedder 中间件
			r.Use(shedder.GinShedderMiddleware(shedder.NewShedderFromCmd(cmd), logger))
			// breaker 中间件
			r.Use(breaker.GinBreakerMiddleware(logger))
			addRouters(r)

			httpServer := exthttp.NewHttpServer(r, logger, exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(":8080"))

			g.Add(func() error {
				statusProber.Healthy()
				return httpServer.Start()
			}, func(err error) {
				statusProber.NotReady(err)
				httpServer.Shutdown(err)
				statusProber.NotHealthy(err)
			})

			// 启动内部 http 服务, 健康检查路由, 监控指标路由, pprof
			mux := extapp.StartInnerHttpServer(extApp, httpProber)
			// 可以在增加额外路由
			mux.HandleFunc("/xxx", func(writer http.ResponseWriter, request *http.Request) {
				writer.Write([]byte("ok"))
			})

			statusProber.Ready()
			extapp.FatalOnErrorf(g.Run(), "start error")
		},
	}

	extapp.RunDefaultServerApp(app)
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
		io.Copy(ioutil.Discard, resp.Body)
		c.String(200, "hello %s", c.Param("id"))
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
