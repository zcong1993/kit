package main

import (
	"log"
	"time"

	"github.com/zcong1993/kit/pkg/extapp"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/zcong1993/kit/_examples/grpc/pb"
	"github.com/zcong1993/kit/pkg/extgrpcc"
	"github.com/zcong1993/kit/pkg/ginhelper"
	"github.com/zcong1993/kit/pkg/server/exthttp"
	"google.golang.org/grpc"
)

func gatewayCmd(app *extapp.App) *cobra.Command {
	var (
		serverAddr     string
		grpcServerAddr string
	)

	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "sub command for gateway",
		Run: func(cmd *cobra.Command, args []string) {
			app.InitFromCmd(cmd, "gateway")

			// 真正的业务 http server
			// 初始化 gin
			r := app.GinServer(exthttp.WithGracePeriod(time.Second*5), exthttp.WithListen(serverAddr))

			conn, err := grpc.Dial(grpcServerAddr, extgrpcc.ClientOtelGrpcOpts(app.Reg, false)...)
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

			extapp.FatalOnErrorf(app.Start(), "start error")
		},
	}

	cmd.Flags().StringVar(&serverAddr, "server-addr", ":8080", "grpc server addr")
	cmd.Flags().StringVar(&grpcServerAddr, "grpc-server-addr", "localhost:8081", "upstream grpc server addr")

	return cmd
}
