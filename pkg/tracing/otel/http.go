package oteltracing

import (
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	// GinMiddleware is alias for otelgin.Middleware.
	GinMiddleware = otelgin.Middleware
	// HTTPMiddleware is alias for telhttp.NewHandler.
	HTTPMiddleware = otelhttp.NewHandler
)

// HttpTransport is alias for otelhttp.NewTransport.
var HttpTransport = otelhttp.NewTransport
