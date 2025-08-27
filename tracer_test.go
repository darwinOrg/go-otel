package dgotel

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func TestTracer(t *testing.T) {
	shutdown := InitTracer("test-service", "", "")
	defer shutdown()

	for i := 0; i < 10; i++ {
		ctx := context.Background()
		parentMethod(ctx)
	}
	time.Sleep(10 * time.Second)
}

func parentMethod(ctx context.Context) {
	tracer := otel.Tracer("otel-go-tracer")
	ctx, span := tracer.Start(ctx, "parent span")
	fmt.Println(span.SpanContext().TraceID()) // 打印 TraceId
	span.SetAttributes(attribute.String("key", "value"))
	span.SetStatus(codes.Ok, "Success")
	childMethod(ctx)
	span.End()
}

func childMethod(ctx context.Context) {
	tracer := otel.Tracer("otel-go-tracer")
	ctx, span := tracer.Start(ctx, "child span")
	span.SetStatus(codes.Ok, "Success")
	grandChildMethod(ctx)
	span.End()
}

func grandChildMethod(ctx context.Context) {
	tracer := otel.Tracer("otel-go-tracer")
	ctx, span := tracer.Start(ctx, "grandchild span")
	span.SetStatus(codes.Error, "error")

	// 业务代码...

	span.End()
}
