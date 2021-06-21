package shedder

import (
	"context"

	"github.com/zcong1993/x/pkg/server/extgrpc"
	"github.com/zcong1993/x/pkg/zero"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/tal-tech/go-zero/core/load"
	"github.com/tal-tech/go-zero/core/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WithGrpcShedder(logger log.Logger, shedder load.Shedder) extgrpc.Option {
	// noop middleware
	if shedder == nil {
		level.Info(logger).Log("component", "grpc/shedder", "msg", "disable middleware")
		return extgrpc.NoopOption()
	}

	level.Info(logger).Log("component", "grpc/shedder", "msg", "load middleware")

	zero.SetupMetrics()
	metrics := zero.Metrics
	sheddingStat := load.NewSheddingStat("grpc")

	return extgrpc.CombineOptions(
		extgrpc.WithUnaryServerInterceptor(unaryServerInterceptor(logger, shedder, metrics, sheddingStat)),
		extgrpc.WithStreamServerInterceptor(streamServerInterceptor(logger, shedder, metrics, sheddingStat)),
	)
}

func unaryServerInterceptor(logger log.Logger, shedder load.Shedder, metrics *stat.Metrics, sheddingStat *load.SheddingStat) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		sheddingStat.IncrementTotal()
		promise, err := shedder.Allow()
		if err != nil {
			metrics.AddDrop()
			sheddingStat.IncrementDrop()
			level.Error(logger).Log("component", "grpc/shedder", "msg", "[grpc] dropped", "type", "unary", "method", info.FullMethod)
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

func streamServerInterceptor(logger log.Logger, shedder load.Shedder, metrics *stat.Metrics, sheddingStat *load.SheddingStat) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		sheddingStat.IncrementTotal()
		promise, err := shedder.Allow()
		if err != nil {
			metrics.AddDrop()
			sheddingStat.IncrementDrop()
			level.Error(logger).Log("component", "grpc/shedder", "msg", "[grpc] dropped", "type", "stream", "method", info.FullMethod)
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
