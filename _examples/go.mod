module github.com/zcong1993/x/_examples

go 1.16

replace github.com/zcong1993/x => ../

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/golang/protobuf v1.5.2
	github.com/spf13/cobra v1.1.3
	github.com/zcong1993/x v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.21.0
	go.opentelemetry.io/otel v1.0.0-RC1
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.27.1
)
