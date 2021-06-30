package breaker

import (
	"context"
	"fmt"

	"github.com/zcong1993/kit/pkg/log"
	"go.uber.org/zap"

	"github.com/zcong1993/kit/pkg/server/extgrpc"
	"github.com/zcong1993/kit/pkg/zero"

	"github.com/tal-tech/go-zero/core/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WithGrpcServerBreaker create breaker option based on command line parameters.
func WithGrpcServerBreaker(logger *log.Logger, opt *Option) extgrpc.Option {
	logger = logger.With(log.Component("grpc/shedder"))
	if opt.disable {
		logger.Info("disable middleware")
		return extgrpc.NoopOption()
	}

	zero.SetupMetrics()
	metrics := zero.Metrics

	logger.Info("load middleware")

	brkGetter := NewBrkGetter()

	return extgrpc.CombineOptions(
		extgrpc.WithUnaryServerInterceptor(unaryServerInterceptor(logger, metrics, brkGetter)),
		extgrpc.WithStreamServerInterceptor(streamServerInterceptor(logger, metrics, brkGetter)),
	)
}

func unaryServerInterceptor(logger *log.Logger, metrics *stat.Metrics, brkGetter *BrkGetter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		brk := brkGetter.Get(info.FullMethod)
		// breaker logic.
		promise, err := brk.Allow()
		if err != nil {
			metrics.AddDrop()
			logger.Error("[grpc] dropped", zap.String("type", "unary"), zap.String("method", info.FullMethod))
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

func streamServerInterceptor(logger *log.Logger, metrics *stat.Metrics, brkGetter *BrkGetter) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		brk := brkGetter.Get(info.FullMethod)
		// breaker logic.
		promise, err := brk.Allow()
		if err != nil {
			metrics.AddDrop()
			logger.Error("[grpc] dropped", zap.String("type", "stream"), zap.String("method", info.FullMethod))
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
