package main

import (
	"io"
	"time"

	oteltracing "github.com/zcong1993/x/pkg/tracing/otel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/zcong1993/x/pkg/extapp"

	"github.com/zcong1993/x/pkg/breaker"
	"github.com/zcong1993/x/pkg/shedder"

	"github.com/zcong1993/x/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/extrun"
	"github.com/zcong1993/x/pkg/ginhelper"
	"github.com/zcong1993/x/pkg/prober"
	"github.com/zcong1993/x/pkg/server/exthttp"
)

var service2Cmd = &cobra.Command{
	Use:   "service2",
	Short: "sub command for service2",
	Run: func(cmd *cobra.Command, args []string) {
		extApp := extapp.NewFromCmd(cmd)

		logger := extApp.Logger
		reg := extApp.Reg
		g := extApp.G
		serviceName := extApp.App

		extapp.FatalOnErrorf(oteltracing.InitTracerFromEnv(serviceName), "init tracer error")

		// 服务健康状态
		httpProber := prober.NewHTTP()
		statusProber := prober.Combine(httpProber, prober.NewInstrumentation("gin", logger, reg))

		// 监听退出信号
		extrun.HandleSignal(g)

		// 真正的业务 http server
		// 初始化 gin
		r := ginhelper.DefaultWithLogger(logger)
		r.Use(metrics.NewInstrumentationMiddleware(reg))
		r.Use(oteltracing.GinMiddleware(serviceName))
		// shedder 中间件
		r.Use(shedder.GinShedderMiddleware(shedder.NewShedderFromCmd(cmd), logger))
		// breaker 中间件
		r.Use(breaker.GinBreakerMiddleware(logger))
		addRoutersV2(r)

		httpServer := exthttp.NewHttpServer(r, logger, exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(":8080"))
		httpServer.Run(g, statusProber)

		// metrics 和 profiler 服务, debug 和监控
		profileServer := exthttp.NewMuxServer(logger, exthttp.WithListen(":6060"), exthttp.WithServiceName("metrics/profiler"))
		profileServer.RegisterProfiler()
		profileServer.RegisterMetrics(extApp.Registry)
		profileServer.RegisterProber(httpProber)
		profileServer.RunGroup(g)

		statusProber.Ready()
		extapp.FatalOnErrorf(g.Run(), "start error")
	},
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
