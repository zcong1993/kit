package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zcong1993/x/pkg/shedder"

	"github.com/zcong1993/x/pkg/server/exthttp/breaker"

	"github.com/zcong1993/x/pkg/metrics"

	"github.com/zcong1993/x/pkg/tracing"
	"github.com/zcong1993/x/pkg/tracing/register"

	"github.com/zcong1993/x/pkg/prober"

	"github.com/zcong1993/x/pkg/server/exthttp"

	"github.com/spf13/cobra"
	log2 "github.com/zcong1993/x/pkg/log"

	"github.com/oklog/run"
	"github.com/zcong1993/x/pkg/extrun"

	"github.com/gin-gonic/gin"
	"github.com/zcong1993/x/pkg/ginhelper"
)

type Input struct {
	Name string `json:"name" binding:"required"`
}

func main() {
	app := &cobra.Command{
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

	// 注册日志相关 flag
	log2.Register(app.PersistentFlags())
	register.RegisterFlags(app.PersistentFlags())
	// 注册 shedder flag
	shedder.Register(app.PersistentFlags())

	if err := app.Execute(); err != nil {
		log.Fatal(err)
	}
}

func addRouters(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		go func() {
			tracing.DoInSpan(c.Request.Context(), "bg work", func(ctx context.Context) {
				time.Sleep(time.Second)
			})
		}()
		time.Sleep(2 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	r.GET("/pets/:id", func(c *gin.Context) {
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
