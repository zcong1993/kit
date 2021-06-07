package oteltracing

import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

var (
	UnaryServerInterceptor  = otelgrpc.UnaryServerInterceptor
	StreamServerInterceptor = otelgrpc.StreamServerInterceptor
)

var (
	UnaryClientInterceptor  = otelgrpc.UnaryClientInterceptor
	StreamClientInterceptor = otelgrpc.StreamClientInterceptor
)
