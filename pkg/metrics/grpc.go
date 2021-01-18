package metrics

import (
	"runtime/debug"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zcong1993/x/pkg/server/extgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	panicsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})
	serverOnce sync.Once
)

func WithServerMetrics(logger log.Logger, reg prometheus.Registerer) extgrpc.Option {
	if reg == nil {
		return extgrpc.NoopOption()
	}

	met := grpc_prometheus.NewServerMetrics()
	met.EnableHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
	)

	grpcPanicRecoveryHandler := func(p interface{}) (err error) {
		panicsTotal.Inc()
		level.Error(logger).Log("msg", "recovered from panic", "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}

	return extgrpc.CombineOptions(
		extgrpc.WithUnaryServerInterceptor(met.UnaryServerInterceptor()),
		extgrpc.WithUnaryServerInterceptor(grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(grpcPanicRecoveryHandler))),
		extgrpc.WithStreamServerInterceptor(met.StreamServerInterceptor()),
		extgrpc.WithStreamServerInterceptor(grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(grpcPanicRecoveryHandler))),
		extgrpc.WithServer(func(s *grpc.Server) {
			met.InitializeMetrics(s)
			serverOnce.Do(func() {
				reg.MustRegister(met, panicsTotal)
			})
		}),
	)
}
