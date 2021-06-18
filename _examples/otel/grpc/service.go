package main

import (
	"context"
	"fmt"
	"io"
	"time"

	oteltracing "github.com/zcong1993/x/pkg/tracing/otel"

	"github.com/go-kit/kit/log"
	"github.com/zcong1993/x/pkg/server/extgrpc"
	"google.golang.org/grpc"

	"google.golang.org/grpc/metadata"

	"github.com/zcong1993/x/pkg/extapp"

	"github.com/spf13/cobra"
	"github.com/zcong1993/x/_examples/otel/grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type helloService struct {
	pb.UnimplementedHelloServer
	logger log.Logger
}

func req2res(req *pb.HelloRequest) *pb.HelloResponse {
	return &pb.HelloResponse{Value: fmt.Sprintf("resp-%s-%d", req.Name, req.Sleep)}
}

func (h *helloService) Get(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	if in.Sleep > 5 {
		return nil, status.Error(codes.ResourceExhausted, "test")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		fmt.Printf("%+v\n", md)
	}

	oteltracing.DoInSpan(ctx, "bg work", func(ctx context.Context) {
		time.Sleep(time.Second)
	})

	if in.Sleep > 0 {
		time.Sleep(time.Duration(int64(time.Second) * int64(in.Sleep)))
	}

	return &pb.HelloResponse{Value: "hello " + in.Name}, nil
}

func (h *helloService) ServerStream(req *pb.HelloRequest, stream pb.Hello_ServerStreamServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		fmt.Printf("%+v\n", md)
	}

	res := req2res(req)
	i := 0
	for i < 5 {
		err := stream.Send(res)
		if err != nil {
			return err
		}
		i++
		time.Sleep(time.Millisecond * 200)
	}
	return nil
}

func (h *helloService) ClientStream(stream pb.Hello_ClientStreamServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		fmt.Printf("%+v\n", md)
	}

	var r *pb.HelloRequest
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		r = req
	}

	time.Sleep(time.Millisecond * 500)

	return stream.SendAndClose(req2res(r))
}

func (h *helloService) DuplexStream(stream pb.Hello_DuplexStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = stream.Send(req2res(req))
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 200)
	}

	return nil
}

func serviceCmd(app *extapp.App) *cobra.Command {
	var serverAddr string

	cmd := &cobra.Command{
		Use:   "service",
		Short: "sub command for service",
		Run: func(cmd *cobra.Command, args []string) {
			app.InitFromCmd(cmd, "server")

			app.GrpcServer(
				extgrpc.WithGracePeriod(time.Second*5),
				extgrpc.WithListen(serverAddr),
				extgrpc.WithServer(func(s *grpc.Server) {
					pb.RegisterHelloServer(s, &helloService{logger: app.Logger})
				}),
			)

			extapp.FatalOnErrorf(app.Start(), "start error")
		},
	}

	cmd.Flags().StringVar(&serverAddr, "server-addr", ":8082", "grpc server addr")

	return cmd
}

func mustGet(f func() (interface{}, error)) interface{} {
	val, err := f()
	extapp.FatalOnError(err)
	return val
}
