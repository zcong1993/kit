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
	"github.com/zcong1993/x/pkg/server/extgrpc"
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

func middleCmd(app *extapp.App) *cobra.Command {
	var (
		serverAddr     string
		grpcServerAddr string
	)

	cmd := &cobra.Command{
		Use:   "middle",
		Short: "sub command for middle service",
		Run: func(cmd *cobra.Command, args []string) {
			app.InitFromCmd(cmd, "middle")

			conn, err := grpc.Dial(grpcServerAddr, extgrpcc.ClientOtelGrpcOpts(app.Reg, false)...)
			grpcClient := pb.NewHelloClient(conn)

			if err != nil {
				log.Fatal(err)
			}

			// 真正的业务 grpc server
			app.GrpcServer(
				extgrpc.WithGracePeriod(time.Second*5),
				extgrpc.WithListen(serverAddr),
				extgrpc.WithServer(func(s *grpc.Server) {
					pb.RegisterHelloServer(s, &middleService{client: grpcClient})
				}),
			)

			extapp.FatalOnErrorf(app.Start(), "start error")
		},
	}

	cmd.Flags().StringVar(&serverAddr, "server-addr", ":8081", "grpc server addr")
	cmd.Flags().StringVar(&grpcServerAddr, "grpc-server-addr", "localhost:8082", "upstream grpc server addr")

	return cmd
}
