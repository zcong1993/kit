package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/zcong1993/x/pkg/breaker"
	"github.com/zcong1993/x/pkg/shedder"

	"github.com/zcong1993/x/pkg/metrics"

	"github.com/gin-gonic/gin"
	klog "github.com/go-kit/kit/log"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/pkg/extrun"
	"github.com/zcong1993/x/pkg/ginhelper"
	log2 "github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/prober"
	"github.com/zcong1993/x/pkg/server/exthttp"
	"github.com/zcong1993/x/pkg/tracing"
	"github.com/zcong1993/x/pkg/tracing/register"
)

var service2Cmd = &cobra.Command{
	Use:   "service2",
	Short: "sub command for service2",
	Run: func(cmd *cobra.Command, args []string) {
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
		r.Use(metrics.NewInstrumentationMiddleware(nil))
		r.Use(tracing.GinMiddleware(tracer, "gin", logger))
		// shedder 中间件
		r.Use(shedder.GinShedderMiddleware(shedder.NewShedderFromCmd(cmd), logger))
		// breaker 中间件
		r.Use(breaker.GinBreakerMiddleware(logger))
		addRoutersV2(r, httpClient(klog.With(logger, "component", "httpClient")))

		httpServer := exthttp.NewHttpServer(r, logger, exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(":8080"))

		g.Add(func() error {
			statusProber.Healthy()
			return httpServer.Start()
		}, func(err error) {
			statusProber.NotReady(err)
			httpServer.Shutdown(err)
			statusProber.NotHealthy(err)
		})

		// metrics 和 profiler 服务, debug 和监控
		profileServer := exthttp.NewMuxServer(logger, exthttp.WithListen(":6060"), exthttp.WithServiceName("metrics/profiler"))
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

func httpClient(logger klog.Logger) *http.Client {
	c := http.DefaultClient
	c.Transport = tracing.HTTPAutoTripperware(logger, http.DefaultTransport)
	return c
}

func addRoutersV2(r *gin.Engine, client *http.Client) {
	r.GET("/", ginhelper.ErrorWrapper(func(c *gin.Context) error {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8081", nil)
		if err != nil {
			return err
		}
		//span, ctx := tracing.StartSpan(c.Request.Context(), "call localhost:8081")
		//defer span.Finish()
		req = req.WithContext(c.Request.Context())
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		io.Copy(c.Writer, resp.Body)
		return nil
	}))
}