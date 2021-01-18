package extgrpcc

import (
	"math"
	"sync"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/zcong1993/x/pkg/tracing"
	"google.golang.org/grpc"
)

var once sync.Once

// ClientGrpcOpts return return options with tracing and metrics.
func ClientGrpcOpts(tracer opentracing.Tracer, reg prometheus.Registerer, secure bool) []grpc.DialOption {
	grpcMets := grpc_prometheus.NewClientMetrics()
	grpcMets.EnableClientHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				grpcMets.UnaryClientInterceptor(),
				tracing.UnaryClientInterceptor(tracer),
			),
		),
		grpc.WithStreamInterceptor(
			grpc_middleware.ChainStreamClient(
				grpcMets.StreamClientInterceptor(),
				tracing.StreamClientInterceptor(tracer),
			),
		),
	}

	if reg != nil {
		once.Do(func() {
			reg.MustRegister(grpcMets)
		})
	}

	if !secure {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	return dialOpts
}
