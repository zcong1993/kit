package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/zcong1993/x/pkg/extgrpcc"

	"github.com/zcong1993/x/pkg/extapp"

	"github.com/spf13/cobra"
	"github.com/zcong1993/x/_examples/grpc/pb"
	"github.com/zcong1993/x/pkg/breaker"
	"github.com/zcong1993/x/pkg/extrun"
	"github.com/zcong1993/x/pkg/metrics"
	"github.com/zcong1993/x/pkg/prober"
	"github.com/zcong1993/x/pkg/server/extgrpc"
	"github.com/zcong1993/x/pkg/server/exthttp"
	"github.com/zcong1993/x/pkg/shedder"
	"google.golang.org/grpc"
)

type middleService struct {
	client pb.HelloClient
	pb.UnimplementedHelloServer
}

func (h *middleService) Get(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	return h.client.Get(ctx, in)
}

func (h *middleService) ServerStream(req *pb.HelloRequest, stream pb.Hello_ServerStreamServer) error {
	s, err := h.client.ServerStream(context.Background(), req)
	if err != nil {
		return err
	}

	for {
		r, err := s.Recv()
		if err == io.EOF {
			return nil
		}
		err = stream.Send(r)
		if err != nil {
			return err
		}
	}
}

func (h *middleService) ClientStream(stream pb.Hello_ClientStreamServer) error {
	s, err := h.client.ClientStream(stream.Context())
	if err != nil {
		return err
	}

	for {
		req, err := stream.Recv()
		if err != io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = s.Send(req)
		if err != nil {
			return err
		}
	}

	resp, err := s.CloseAndRecv()
	if err != nil {
		return err
	}
	return stream.SendAndClose(resp)
}

var middleCmd = &cobra.Command{
	Use:   "middle",
	Short: "sub command for middle service",
	Run: func(cmd *cobra.Command, args []string) {
		extApp := extapp.NewFromCmd(cmd)

		logger := extApp.Logger
		tracer := extApp.Tracer
		reg := extApp.Reg
		g := extApp.G

		// shedder
		sd := shedder.NewShedderFromCmd(cmd)

		// 服务健康状态
		grpcProber := prober.NewGRPC()
		httpProber := prober.NewHTTP()
		statusProber := prober.Combine(httpProber, grpcProber, prober.NewInstrumentation("middle", logger))

		// 监听退出信号
		extrun.HandleSignal(g)

		addr := mustGet(func() (interface{}, error) {
			return cmd.Flags().GetString("server-addr")
		}).(string)

		upstreamAddr := mustGet(func() (interface{}, error) {
			return cmd.Flags().GetString("grpc-server-addr")
		}).(string)

		conn, err := grpc.Dial(upstreamAddr, extgrpcc.ClientGrpcOpts(tracer, reg, false)...)
		grpcClient := pb.NewHelloClient(conn)

		if err != nil {
			log.Fatal(err)
		}

		// 真正的业务 grpc server
		grpcServer := extgrpc.NewServer(logger, grpcProber,
			extgrpc.WithGracePeriod(time.Second*5),
			extgrpc.WithListen(addr),
			extgrpc.WithServer(func(s *grpc.Server) {
				pb.RegisterHelloServer(s, &middleService{client: grpcClient})
			}),
			metrics.WithServerMetrics(logger, reg),
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
		profileServer.RegisterMetrics(reg)
		profileServer.RegisterProber(httpProber)
		profileServer.RunGroup(g)

		statusProber.Ready()
		extapp.FatalOnErrorf(g.Run(), "start error")
	},
}

func init() {
	middleCmd.Flags().String("server-addr", ":8081", "grpc server addr")
	middleCmd.Flags().String("metrics-addr", ":6061", "metrics server addr")
	middleCmd.Flags().String("grpc-server-addr", "localhost:8082", "upstream grpc server addr")
}
