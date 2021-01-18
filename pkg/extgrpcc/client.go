package extgrpcc

import (
	"math"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/zcong1993/x/pkg/tracing"
	"google.golang.org/grpc"
)

// ClientGrpcOpts return return options with tracing and metrics.
func ClientGrpcOpts(tracer opentracing.Tracer, secure bool) []grpc.DialOption {
	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				tracing.UnaryClientInterceptor(tracer),
			),
		),
		grpc.WithStreamInterceptor(
			grpc_middleware.ChainStreamClient(
				tracing.StreamClientInterceptor(tracer),
			),
		),
	}

	if !secure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	return dialOpts
}
