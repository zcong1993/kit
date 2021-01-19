package breaker

import (
	"context"
	"fmt"
	"sync"

	"github.com/zcong1993/x/pkg/server/extgrpc"
	"github.com/zcong1993/x/pkg/zero"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tal-tech/go-zero/core/breaker"
	"github.com/tal-tech/go-zero/core/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	grpcLock       sync.Mutex
	grpcBreakerMap = make(map[string]breaker.Breaker)
)

func getGrpcBreaker(key string) breaker.Breaker {
	grpcLock.Lock()
	defer grpcLock.Unlock()
	if brk, ok := grpcBreakerMap[key]; ok {
		return brk
	}
	grpcBreakerMap[key] = breaker.NewBreaker(breaker.WithName(key))
	return grpcBreakerMap[key]
}

func WithGrpcServerBreaker(logger log.Logger) extgrpc.Option {
	zero.SetupMetrics()
	metrics := zero.Metrics

	level.Info(logger).Log("component", "breaker-grpc", "msg", "load middleware")

	return extgrpc.CombineOptions(
		extgrpc.WithUnaryServerInterceptor(unaryServerInterceptor(logger, metrics)),
		extgrpc.WithStreamServerInterceptor(streamServerInterceptor(logger, metrics)),
	)
}

func unaryServerInterceptor(logger log.Logger, metrics *stat.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		brk := getGrpcBreaker(info.FullMethod)
		// breaker logic
		promise, err := brk.Allow()
		if err != nil {
			metrics.AddDrop()
			level.Error(logger).Log("component", "breaker-grpc", "msg", "[grpc] dropped", "type", "unary", "method", info.FullMethod)
			return nil, status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_breaker middleware, please retry later.", info.FullMethod)
		}

		resp, err := handler(ctx, req)

		if err == nil {
			promise.Accept()
		} else {
			errStatus, _ := status.FromError(err)
			promise.Reject(fmt.Sprintf("%s-%s", errStatus.Message(), err.Error()))
		}

		return resp, err
	}
}

func streamServerInterceptor(logger log.Logger, metrics *stat.Metrics) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		brk := getGrpcBreaker(info.FullMethod)
		// breaker logic
		promise, err := brk.Allow()
		if err != nil {
			metrics.AddDrop()
			level.Error(logger).Log("component", "breaker-grpc", "msg", "[grpc] dropped", "type", "stream", "method", info.FullMethod)
			return status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_breaker middleware, please retry later.", info.FullMethod)
		}

		err = handler(srv, stream)

		if err == nil {
			promise.Accept()
		} else {
			errStatus, _ := status.FromError(err)
			promise.Reject(fmt.Sprintf("%s-%s", errStatus.Message(), err.Error()))
		}

		return err
	}
}
