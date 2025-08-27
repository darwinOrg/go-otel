package dgotel

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	Tracer            trace.Tracer
	tracerServiceName string
)

// InitTracer 初始化 OpenTelemetry 并配置导出
func InitTracer(serviceName string, exporter sdktrace.SpanExporter) (func(), error) {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	Tracer = otel.Tracer(serviceName)
	tracerServiceName = serviceName

	// 返回一个关闭函数，用于优雅关闭Tracer
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("TracerProvider Shutdown Error: %v", err)
		}
	}, nil
}

func GetTracerServiceName() string {
	return tracerServiceName
}
