package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"

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

var (
	requestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Http requests total count.",
	}, []string{"router"})
)

func main() {
	prometheus.MustRegister(requestTotal)

	app := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化日志
			logger := log2.MustNewLogger(cmd)
			// 初始化 gin
			r := ginhelper.DefaultWithLogger(logger)

			r.GET("/", func(c *gin.Context) {
				requestTotal.WithLabelValues("/").Inc()
				time.Sleep(2 * time.Second)
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

			var g run.Group

			// 监听退出信号
			extrun.HandleSignal(&g)

			httpServer := exthttp.NewHttpServer(r, logger, exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(":8080"))

			// 真正的业务 http server
			g.Add(func() error {
				return httpServer.Start()
			}, func(err error) {
				httpServer.Shutdown(err)
			})

			// metrics 和 profiler 服务, debug 和监控
			profileServer := exthttp.NewMuxServer(logger, exthttp.WithListen(":6060"), exthttp.WithServiceName("metrics/profiler"))
			profileServer.RegisterProfiler()
			profileServer.RegisterMetrics(nil)
			profileServer.RunGroup(&g)

			if err := g.Run(); err != nil {
				log.Fatal("start error ", err)
			}
		},
	}

	// 注册日志相关 flag
	log2.Register(app)

	if err := app.Execute(); err != nil {
		log.Fatal(err)
	}
}
