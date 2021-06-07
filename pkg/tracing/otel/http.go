package oteltracing

import (
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	GinMiddleware  = otelgin.Middleware
	HTTPMiddleware = otelhttp.NewHandler
)

var HttpTransport = otelhttp.NewTransport
