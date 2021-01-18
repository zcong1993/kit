package main

import (
	"context"
	"log"
	"time"

	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/_examples/grpcserver/pb"
	"github.com/zcong1993/x/pkg/extapp"
	"github.com/zcong1993/x/pkg/extrun"
	log2 "github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/prober"
	"github.com/zcong1993/x/pkg/server/extgrpc"
	"github.com/zcong1993/x/pkg/server/exthttp"
	"google.golang.org/grpc"
)

type helloService struct {
	pb.UnimplementedHelloServer
}

func (h *helloService) Get(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	if in.Sleep > 0 {
		time.Sleep(time.Duration(int64(time.Second) * int64(in.Sleep)))
	}
	return &pb.HelloResponse{Value: "hello " + in.Name}, nil
}

func main() {
	app := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化日志
			logger := log2.MustNewLogger(cmd)

			var g run.Group

			// 初始化 tracer
			//tracer := register.MustInitTracer(&g, cmd, logger, nil)

			// 服务健康状态
			grpcProber := prober.NewGRPC()
			httpProber := prober.NewHTTP()
			statusProber := prober.Combine(httpProber, grpcProber, prober.NewInstrumentation("gin", logger))

			// 监听退出信号
			extrun.HandleSignal(&g)

			// 真正的业务 grpc server
			grpcServer := extgrpc.NewServer(logger, grpcProber,
				extgrpc.WithGracePeriod(time.Second*5),
				extgrpc.WithListen(":8081"),
				extgrpc.WithServer(func(s *grpc.Server) {
					pb.RegisterHelloServer(s, &helloService{})
				}),
			)

			g.Add(func() error {
				statusProber.Healthy()
				return grpcServer.ListenAndServe()
			}, func(err error) {
				statusProber.NotReady(err)
				grpcServer.Shutdown(err)
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

	extapp.RunDefaultGrpcServerApp(app)
}
