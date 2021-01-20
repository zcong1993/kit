package main

import (
	"context"
	"log"
	"time"

	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/_examples/grpc/pb"
	"github.com/zcong1993/x/pkg/breaker"
	"github.com/zcong1993/x/pkg/extrun"
	log2 "github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/metrics"
	"github.com/zcong1993/x/pkg/prober"
	"github.com/zcong1993/x/pkg/server/extgrpc"
	"github.com/zcong1993/x/pkg/server/exthttp"
	"github.com/zcong1993/x/pkg/shedder"
	"github.com/zcong1993/x/pkg/tracing/register"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type helloService struct {
	pb.UnimplementedHelloServer
}

func (h *helloService) Get(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	if in.Sleep > 5 {
		return nil, status.Error(codes.ResourceExhausted, "test")
	}

	if in.Sleep > 0 {
		time.Sleep(time.Duration(int64(time.Second) * int64(in.Sleep)))
	}

	return &pb.HelloResponse{Value: "hello " + in.Name}, nil
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "sub command for service",
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化日志
		logger := log2.MustNewLogger(cmd)

		var g run.Group

		me := metrics.InitMetrics()

		// 初始化 tracer
		tracer := register.MustInitTracer(&g, cmd, logger, me)

		// shedder
		sd := shedder.NewShedderFromCmd(cmd)

		// 服务健康状态
		grpcProber := prober.NewGRPC()
		httpProber := prober.NewHTTP()
		statusProber := prober.Combine(httpProber, grpcProber, prober.NewInstrumentation("grpc", logger))

		// 监听退出信号
		extrun.HandleSignal(&g)

		addr := mustGet(func() (interface{}, error) {
			return cmd.Flags().GetString("server-addr")
		}).(string)

		// 真正的业务 grpc server
		grpcServer := extgrpc.NewServer(logger, grpcProber,
			extgrpc.WithGracePeriod(time.Second*5),
			extgrpc.WithListen(addr),
			extgrpc.WithServer(func(s *grpc.Server) {
				pb.RegisterHelloServer(s, &helloService{})
			}),
			metrics.WithServerMetrics(logger, me),
			extgrpc.WithServerTracing(tracer),
			shedder.WithGrpcShedder(logger, sd),
			breaker.WithGrpcServerBreaker(logger),
		)

		g.Add(func() error {
			statusProber.Healthy()
			return grpcServer.ListenAndServe()
		}, func(err error) {
			statusProber.NotReady(err)
			grpcServer.Shutdown(err)
			statusProber.NotHealthy(err)
		})

		metricsAddr := mustGet(func() (interface{}, error) {
			return cmd.Flags().GetString("metrics-addr")
		}).(string)

		// metrics 和 profiler 服务, debug 和监控
		profileServer := exthttp.NewMuxServer(logger, exthttp.WithListen(metricsAddr), exthttp.WithServiceName("metrics/profiler"))
		profileServer.RegisterProfiler()
		profileServer.RegisterMetrics(me)
		profileServer.RegisterProber(httpProber)
		profileServer.RunGroup(&g)

		statusProber.Ready()
		if err := g.Run(); err != nil {
			log.Fatal("start error ", err)
		}
	},
}

func mustGet(f func() (interface{}, error)) interface{} {
	val, err := f()
	if err != nil {
		log.Fatal(err)
	}
	return val
}

func init() {
	serviceCmd.Flags().String("server-addr", ":8081", "grpc server addr")
	serviceCmd.Flags().String("metrics-addr", ":6061", "metrics server addr")
}
