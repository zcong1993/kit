package extgrpcc

import (
	"math"
	"sync"

	oteltracing "github.com/zcong1993/kit/pkg/tracing/otel"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"

	"google.golang.org/grpc"
)

var once sync.Once

func ClientOtelGrpcOpts(reg prometheus.Registerer, secure bool) []grpc.DialOption {
	return BuildOpts(reg, secure, []grpc.UnaryClientInterceptor{oteltracing.UnaryClientInterceptor()}, []grpc.StreamClientInterceptor{oteltracing.StreamClientInterceptor()})
}

func BuildOpts(reg prometheus.Registerer, secure bool, unaryClientInterceptors []grpc.UnaryClientInterceptor, streamClientInterceptors []grpc.StreamClientInterceptor) []grpc.DialOption {
	grpcMets := grpc_prometheus.NewClientMetrics()
	grpcMets.EnableClientHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
	)

	unaryClientInterceptors = append(unaryClientInterceptors, grpcMets.UnaryClientInterceptor())
	streamClientInterceptors = append(streamClientInterceptors, grpcMets.StreamClientInterceptor())

	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)),
		grpc.WithChainUnaryInterceptor(unaryClientInterceptors...),
		grpc.WithChainStreamInterceptor(streamClientInterceptors...),
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
