package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/zcong1993/x/_examples/grpc/pb"
	"github.com/zcong1993/x/pkg/extapp"
	"github.com/zcong1993/x/pkg/extgrpcc"
	"github.com/zcong1993/x/pkg/log"
	"github.com/zcong1993/x/pkg/tracing/register"
	"google.golang.org/grpc"
)

func main() {
	app := &cobra.Command{
		Use:   "grpc client test",
		Short: "grpc client",
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.MustNewLogger(cmd)
			reg := prometheus.NewRegistry()
			tracer, _, err := register.NewTracer(cmd, logger, reg)
			if err != nil {
				level.Error(logger).Log("msg", err.Error())
				os.Exit(2)
			}

			conn, err := grpc.Dial(os.Getenv("server"), extgrpcc.ClientGrpcOpts(tracer, reg, false)...)
			grpcClient := pb.NewHelloClient(conn)

			resp, err := grpcClient.Get(context.Background(), &pb.HelloRequest{
				Name:  "hello",
				Sleep: 2,
			})

			if err != nil {
				level.Error(logger).Log("msg", err.Error())
				os.Exit(2)
			} else {
				fmt.Printf("%+v\n", resp)
			}

			s, err := grpcClient.ServerStream(context.Background(), &pb.HelloRequest{
				Name:  "serverStream",
				Sleep: 1,
			})

			if err != nil {
				level.Error(logger).Log("msg", err.Error())
				os.Exit(2)
			} else {
				for {
					r, err := s.Recv()
					if err != nil {
						break
					}
					fmt.Printf("%+v\n", r)
				}
			}

			// tracing not work for client side stream

			//ss, err := grpcClient.ClientStream(context.Background())
			//if err != nil {
			//	extapp.FatalOnError(err)
			//} else {
			//	i := 0
			//	for i < 4 {
			//		req := &pb.HelloRequest{
			//			Name:  "clientStream",
			//			Sleep: 1,
			//		}
			//		err := ss.Send(req)
			//		extapp.FatalOnError(err)
			//		i++
			//	}
			//	ss.CloseSend()
			//	r, err := ss.CloseAndRecv()
			//	extapp.FatalOnError(err)
			//	fmt.Printf("%+v\n", r)
			//}
		},
	}

	log.Register(app.PersistentFlags())
	register.RegisterFlags(app.PersistentFlags())

	extapp.FatalOnError(app.Execute())
}
