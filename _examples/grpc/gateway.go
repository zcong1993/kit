package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/_examples/grpc/pb"
	"github.com/zcong1993/x/pkg/breaker"
	"github.com/zcong1993/x/pkg/extgrpcc"
	"github.com/zcong1993/x/pkg/extrun"
	"github.com/zcong1993/x/pkg/ginhelper"
	log2 "github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/metrics"
	"github.com/zcong1993/x/pkg/prober"
	"github.com/zcong1993/x/pkg/server/exthttp"
	"github.com/zcong1993/x/pkg/shedder"
	"github.com/zcong1993/x/pkg/tracing"
	"github.com/zcong1993/x/pkg/tracing/register"
	"google.golang.org/grpc"
)

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "sub command for gateway",
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化日志
		logger := log2.MustNewLogger(cmd)

		var g run.Group

		me := metrics.InitMetrics()

		// 初始化 tracer
		tracer := register.MustInitTracer(&g, cmd, logger, me)

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

		upstreamAddr := mustGet(func() (interface{}, error) {
			return cmd.Flags().GetString("grpc-server-addr")
		}).(string)

		conn, err := grpc.Dial(upstreamAddr, extgrpcc.ClientGrpcOpts(tracer, me, false)...)
		grpcClient := pb.NewHelloClient(conn)

		if err != nil {
			log.Fatal(err)
		}

		r.POST("/hello", ginhelper.ErrorWrapper(func(c *gin.Context) error {
			var getRequest pb.HelloRequest
			if err := c.BindJSON(&getRequest); err != nil {
				return err
			}
			resp, err := grpcClient.Get(c.Request.Context(), &getRequest)
			if err != nil {
				return err
			}
			c.JSON(200, resp)
			return nil
		}))

		addr := mustGet(func() (interface{}, error) {
			return cmd.Flags().GetString("server-addr")
		}).(string)

		httpServer := exthttp.NewHttpServer(r, logger, exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(addr))

		g.Add(func() error {
			statusProber.Healthy()
			return httpServer.Start()
		}, func(err error) {
			statusProber.NotReady(err)
			httpServer.Shutdown(err)
			statusProber.NotHealthy(err)
		})

		metricsAddr := mustGet(func() (interface{}, error) {
			return cmd.Flags().GetString("metrics-addr")
		}).(string)

		// metrics 和 profiler 服务, debug 和监控
		profileServer := exthttp.NewMuxServer(logger, exthttp.WithListen(metricsAddr), exthttp.WithServiceName("metrics/profiler"))
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

func init() {
	gatewayCmd.Flags().String("server-addr", ":8081", "grpc server addr")
	gatewayCmd.Flags().String("metrics-addr", ":6061", "metrics server addr")
	gatewayCmd.Flags().String("grpc-server-addr", "localhost:8081", "upstream grpc server addr")
}
