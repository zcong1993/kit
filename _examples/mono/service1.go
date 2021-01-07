package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/extrun"
	"github.com/zcong1993/x/pkg/ginhelper"
	log2 "github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/prober"
	"github.com/zcong1993/x/pkg/server/exthttp"
	"github.com/zcong1993/x/pkg/tracing"
	"github.com/zcong1993/x/pkg/tracing/register"
)

var (
	requestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Http requests total count.",
	}, []string{"router"})

	service1Cmd = &cobra.Command{
		Use:   "service1",
		Short: "sub command for service1",
		Run: func(cmd *cobra.Command, args []string) {
			prometheus.MustRegister(requestTotal)
			// 初始化日志
			logger := log2.MustNewLogger(cmd)

			var g run.Group

			// 初始化 tracer
			tracer := register.MustInitTracer(&g, cmd, logger, nil)

			// 服务健康状态
			httpProber := prober.NewHTTP()
			statusProber := prober.Combine(httpProber, prober.NewInstrumentation("gin", logger))

			// 监听退出信号
			extrun.HandleSignal(&g)

			// 真正的业务 http server
			// 初始化 gin
			r := ginhelper.DefaultWithLogger(logger)
			r.Use(tracing.GinMiddleware(tracer, "gin", logger))
			addRouters(r)

			httpServer := exthttp.NewHttpServer(r, logger, exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(":8081"))

			g.Add(func() error {
				statusProber.Healthy()
				return httpServer.Start()
			}, func(err error) {
				statusProber.NotReady(err)
				httpServer.Shutdown(err)
				statusProber.NotHealthy(err)
			})

			// metrics 和 profiler 服务, debug 和监控
			profileServer := exthttp.NewMuxServer(logger, exthttp.WithListen(":6061"), exthttp.WithServiceName("metrics/profiler"))
			profileServer.RegisterProfiler()
			profileServer.RegisterMetrics(nil)
			profileServer.RegisterProber(httpProber)
			profileServer.RunGroup(&g)

			statusProber.Ready()
			if err := g.Run(); err != nil {
				log.Fatal("start error ", err)
			}
		},
	}
)

func addRouters(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		requestTotal.WithLabelValues("/").Inc()
		go func() {
			tracing.DoInSpan(c.Request.Context(), "bg work", func(ctx context.Context) {
				time.Sleep(time.Second)
			})
		}()
		time.Sleep(2 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})
}
