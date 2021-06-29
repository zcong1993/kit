package extgrpc

import (
	"crypto/tls"
	"time"

	oteltracing "github.com/zcong1993/x/pkg/tracing/otel"

	"google.golang.org/grpc"
)

type Options struct {
	registerServerFuncs []registerServerFunc

	gracePeriod time.Duration
	listen      string
	network     string

	tlsConfig *tls.Config

	grpcOpts []grpc.ServerOption

	grpcUnaryServerInterceptors  []grpc.UnaryServerInterceptor
	grpcStreamServerInterceptors []grpc.StreamServerInterceptor
}

type Option = func(o *Options)
type registerServerFunc func(s *grpc.Server)

// WithGRPCServer calls the passed gRPC registration functions on the created
// grpc.Server.
func WithServer(f registerServerFunc) Option {
	return func(o *Options) {
		o.registerServerFuncs = append(o.registerServerFuncs, f)
	}
}

// WithGRPCServerOption allows adding raw grpc.ServerOption's to the
// instantiated gRPC server.
func WithGRPCServerOption(opt grpc.ServerOption) Option {
	return func(o *Options) {
		o.grpcOpts = append(o.grpcOpts, opt)
	}
}

// WithGracePeriod sets shutdown grace period for gRPC server.
// Server waits connections to drain for specified amount of time.
func WithGracePeriod(t time.Duration) Option {
	return func(o *Options) {
		o.gracePeriod = t
	}
}

// WithListen sets address to listen for gRPC server.
// Server accepts incoming connections on given address.
func WithListen(s string) Option {
	return func(o *Options) {
		o.listen = s
	}
}

// WithNetwork sets network to listen for gRPC server e.g tcp, udp or unix.
func WithNetwork(s string) Option {
	return func(o *Options) {
		o.network = s
	}
}

// WithTLSConfig sets TLS configuration for gRPC server.
func WithTLSConfig(cfg *tls.Config) Option {
	return func(o *Options) {
		o.tlsConfig = cfg
	}
}

func WithUnaryServerInterceptor(interceptor grpc.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.grpcUnaryServerInterceptors = append(o.grpcUnaryServerInterceptors, interceptor)
	}
}

func WithStreamServerInterceptor(interceptor grpc.StreamServerInterceptor) Option {
	return func(o *Options) {
		o.grpcStreamServerInterceptors = append(o.grpcStreamServerInterceptors, interceptor)
	}
}

func WithOtelTracing() Option {
	return CombineOptions(
		WithUnaryServerInterceptor(oteltracing.UnaryServerInterceptor()),
		WithStreamServerInterceptor(oteltracing.StreamServerInterceptor()),
	)
}

func CombineOptions(options ...Option) Option {
	return func(o *Options) {
		for _, f := range options {
			f(o)
		}
	}
}

func NoopOption() Option {
	return func(o *Options) {}
}
