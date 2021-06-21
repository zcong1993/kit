package breaker

import (
	"context"
	"fmt"

	"github.com/zcong1993/x/pkg/server/extgrpc"
	"github.com/zcong1993/x/pkg/zero"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tal-tech/go-zero/core/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WithGrpcServerBreaker(logger log.Logger, opt *Option) extgrpc.Option {
	if opt.disable {
		level.Info(logger).Log("component", "grpc/breaker", "msg", "disable middleware")
		return extgrpc.NoopOption()
	}

	zero.SetupMetrics()
	metrics := zero.Metrics

	level.Info(logger).Log("component", "grpc/breaker", "msg", "load middleware")

	brkGetter := NewBrkGetter()

	return extgrpc.CombineOptions(
		extgrpc.WithUnaryServerInterceptor(unaryServerInterceptor(logger, metrics, brkGetter)),
		extgrpc.WithStreamServerInterceptor(streamServerInterceptor(logger, metrics, brkGetter)),
	)
}

func unaryServerInterceptor(logger log.Logger, metrics *stat.Metrics, brkGetter *BrkGetter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		brk := brkGetter.Get(info.FullMethod)
		// breaker logic
		promise, err := brk.Allow()
		if err != nil {
			metrics.AddDrop()
			level.Error(logger).Log("component", "grpc/breaker", "msg", "[grpc] dropped", "type", "unary", "method", info.FullMethod)
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

func streamServerInterceptor(logger log.Logger, metrics *stat.Metrics, brkGetter *BrkGetter) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		brk := brkGetter.Get(info.FullMethod)
		// breaker logic
		promise, err := brk.Allow()
		if err != nil {
			metrics.AddDrop()
			level.Error(logger).Log("component", "grpc/breaker", "msg", "[grpc] dropped", "type", "stream", "method", info.FullMethod)
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
