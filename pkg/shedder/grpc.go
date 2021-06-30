package shedder

import (
	"context"

	"github.com/zcong1993/kit/pkg/log"
	"go.uber.org/zap"

	"github.com/zcong1993/kit/pkg/server/extgrpc"
	"github.com/zcong1993/kit/pkg/zero"

	"github.com/tal-tech/go-zero/core/load"
	"github.com/tal-tech/go-zero/core/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WithGrpcShedder create shedder option based on command line parameters.
func WithGrpcShedder(logger *log.Logger, shedder load.Shedder) extgrpc.Option {
	logger = logger.With(log.Component("grpc/shedder"))
	// noop middleware.
	if shedder == nil {
		logger.Info("disable middleware")
		return extgrpc.NoopOption()
	}

	logger.Info("load middleware")

	zero.SetupMetrics()
	metrics := zero.Metrics
	sheddingStat := load.NewSheddingStat("grpc")

	return extgrpc.CombineOptions(
		extgrpc.WithUnaryServerInterceptor(unaryServerInterceptor(logger, shedder, metrics, sheddingStat)),
		extgrpc.WithStreamServerInterceptor(streamServerInterceptor(logger, shedder, metrics, sheddingStat)),
	)
}

func unaryServerInterceptor(logger *log.Logger, shedder load.Shedder, metrics *stat.Metrics, sheddingStat *load.SheddingStat) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		sheddingStat.IncrementTotal()
		promise, err := shedder.Allow()
		if err != nil {
			metrics.AddDrop()
			sheddingStat.IncrementDrop()
			logger.Error("[grpc] dropped", zap.String("type", "unary"), zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_shedder middleware, please retry later.", info.FullMethod)
		}

		resp, err := handler(ctx, req)

		if err == nil {
			sheddingStat.IncrementPass()
			promise.Pass()
		} else if errStatus, _ := status.FromError(err); errStatus.Code() == codes.ResourceExhausted {
			promise.Fail()
		}

		return resp, err
	}
}

func streamServerInterceptor(logger *log.Logger, shedder load.Shedder, metrics *stat.Metrics, sheddingStat *load.SheddingStat) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		sheddingStat.IncrementTotal()
		promise, err := shedder.Allow()
		if err != nil {
			metrics.AddDrop()
			sheddingStat.IncrementDrop()
			logger.Error("[grpc] dropped", zap.String("type", "stream"), zap.String("method", info.FullMethod))
			return status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_shedder middleware, please retry later.", info.FullMethod)
		}

		err = handler(srv, stream)

		if err == nil {
			sheddingStat.IncrementPass()
			promise.Pass()
		} else if errStatus, _ := status.FromError(err); errStatus.Code() == codes.ResourceExhausted {
			promise.Fail()
		}

		return err
	}
}
