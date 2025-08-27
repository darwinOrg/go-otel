package dgotel

import (
	"context"
	"log"
	"time"

	dgsys "github.com/darwinOrg/go-common/sys"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.15.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	Tracer            trace.Tracer
	tracerServiceName string
)

func InitTracer(serviceName, httpEndpoint, httpUrlPath string) func() {
	ctx := context.Background()
	traceExporter := NewHTTPExporter(ctx, httpEndpoint, httpUrlPath)
	batchSpanProcessor := sdktrace.NewBatchSpanProcessor(traceExporter)
	otelResource := NewResource(ctx, serviceName)
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(otelResource),
		sdktrace.WithSpanProcessor(batchSpanProcessor))
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	Tracer = otel.Tracer(serviceName)
	tracerServiceName = serviceName

	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExporter.Shutdown(cxt); err != nil {
			log.Printf("traceExporter Shutdown Error: %v", err)
			otel.Handle(err)
		}
	}
}

func NewResource(ctx context.Context, serviceName string) *resource.Resource {
	r, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.DeploymentEnvironmentKey.String(dgsys.GetProfile()),
			semconv.HostNameKey.String(dgsys.GetHostName()),
		),
	)
	if err != nil {
		log.Fatalf("%s: %v", "Failed to create OpenTelemetry resource", err)
	}

	return r
}

func NewHTTPExporter(ctx context.Context, httpEndpoint, httpUrlPath string) *otlptrace.Exporter {
	traceExporter, err := otlptrace.New(ctx, otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(httpEndpoint),
		otlptracehttp.WithURLPath(httpUrlPath),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithCompression(1)))
	if err != nil {
		log.Fatalf("%s: %v", "Failed to create the OpenTelemetry trace exporter", err)
	}

	return traceExporter
}

func GetTracerServiceName() string {
	return tracerServiceName
}
