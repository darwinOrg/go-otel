package dgotel

import (
	"context"
	"log"
	"time"

	dgsys "github.com/darwinOrg/go-common/sys"
	"go.opentelemetry.io/otel"
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

func InitTracer(ctx context.Context, serviceName string, exporter sdktrace.SpanExporter) func() {
	batchSpanProcessor := sdktrace.NewBatchSpanProcessor(exporter)
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
		if err := traceProvider.Shutdown(cxt); err != nil {
			log.Printf("traceProvider shutdown error: %v", err)
			otel.Handle(err)
		}
	}
}

func NewResource(ctx context.Context, serviceName string) *resource.Resource {
	r, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.HostNameKey.String(dgsys.GetHostName()),
		),
	)
	if err != nil {
		log.Fatalf("%s: %v", "Failed to create OpenTelemetry resource", err)
	}

	return r
}

func GetTracerServiceName() string {
	return tracerServiceName
}
