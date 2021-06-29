package oteltracing

import (
	"context"
	"os"
	"strconv"

	"github.com/zcong1993/x/pkg/log"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	otelTrace "go.opentelemetry.io/otel/trace"
)

const (
	exporterTypekey    = "OTEL_EXPORTER_TYPE"
	exporterTypeJaeger = "jaeger"
	samplerTypeKey     = "OTEL_TRACES_SAMPLER"
	samplerArgKey      = "OTEL_TRACES_SAMPLER_ARG"
)

func initExporter() (trace.SpanExporter, error) {
	exporterType := os.Getenv(exporterTypekey)
	if exporterType != exporterTypeJaeger {
		return nil, errors.New("only support jaeger exporter now")
	}

	// Endpoint
	if os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT") != "" {
		return jaeger.New(jaeger.WithCollectorEndpoint())
	}

	// agent
	return jaeger.New(jaeger.WithAgentEndpoint())
}

func getSamplerArg() (float64, error) {
	arg := os.Getenv(samplerArgKey)
	return strconv.ParseFloat(arg, 64)
}

func initSampler() (trace.Sampler, error) {
	samplerType := os.Getenv(samplerTypeKey)

	switch samplerType {
	case "AlwaysOn":
		return trace.AlwaysSample(), nil
	case "AlwaysOff":
		return trace.NeverSample(), nil
	case "ParentBasedAlwaysOn":
		return trace.ParentBased(trace.AlwaysSample()), nil
	case "ParentBasedAlwaysOff":
		return trace.ParentBased(trace.NeverSample()), nil
	case "TraceIdRatio":
		arg, err := getSamplerArg()
		if err != nil {
			return nil, errors.Wrap(err, "invalid sampler arg")
		}
		return trace.TraceIDRatioBased(arg), nil
	case "ParentBasedTraceIdRatio":
		arg, err := getSamplerArg()
		if err != nil {
			return nil, errors.Wrap(err, "invalid sampler arg")
		}
		return trace.ParentBased(trace.TraceIDRatioBased(arg)), nil
	default:
		return trace.AlwaysSample(), nil
	}
}

func InitTracerFromEnv(logger *log.Logger, serviceName string, attrs ...attribute.KeyValue) error {
	c := log.Component("otel-tracing")
	if os.Getenv(exporterTypekey) == "" {
		logger.Info("disable tracing", c)
		return nil
	}

	logger.Info("enable tracing", c)

	exporter, err := initExporter()
	if err != nil {
		return errors.Wrap(err, "init exporter")
	}

	sampler, err := initSampler()
	if err != nil {
		return errors.Wrap(err, "init sampler")
	}

	attrs = append(attrs, semconv.ServiceNameKey.String(serviceName))

	tp := trace.NewTracerProvider(
		trace.WithSampler(sampler),
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewSchemaless(attrs...)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

// DoInSpan executes function doFn inside new span with `operationName` name and hooking as child to a span found within given context if any.
// It uses opentracing.Tracer propagated in context. If no found, it uses noop tracer notification.
func DoInSpan(ctx context.Context, operationName string, doFn func(context.Context), opts ...otelTrace.SpanStartOption) {
	tracer := otel.Tracer("DoInSpan")
	newCtx, span := tracer.Start(ctx, operationName, opts...)
	defer span.End()
	doFn(newCtx)
}

// DoWithSpan executes function doFn inside new span with `operationName` name and hooking as child to a span found within given context if any.
// It uses opentracing.Tracer propagated in context. If no found, it uses noop tracer notification.
func DoWithSpan(ctx context.Context, operationName string, doFn func(context.Context, otelTrace.Span), opts ...otelTrace.SpanStartOption) {
	tracer := otel.Tracer("DoWithSpan")
	newCtx, span := tracer.Start(ctx, operationName, opts...)
	defer span.End()
	doFn(newCtx, span)
}
