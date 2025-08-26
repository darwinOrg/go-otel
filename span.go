package dgotel

import (
	"context"

	dgctx "github.com/darwinOrg/go-common/context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func DoActionWithinNewSpan(ctx context.Context, spanName string, opts []trace.SpanStartOption, action func(ctx context.Context, span trace.Span)) {
	if Tracer == nil {
		return
	}

	if opts == nil {
		opts = []trace.SpanStartOption{}
	}
	ctx, span := Tracer.Start(ctx, spanName, opts...)
	defer span.End()

	action(ctx, span)
}

func SetSpanAttributes(ctx *dgctx.DgContext, mp map[string]string) {
	if len(mp) == 0 {
		return
	}

	span := GetSpanByDgContext(ctx)
	if span == nil {
		return
	}

	var attrs []attribute.KeyValue
	for k, v := range mp {
		attrs = append(attrs, attribute.String(k, v))
	}

	span.SetAttributes(attrs...)
}

func GetSpanByDgContext(ctx *dgctx.DgContext) trace.Span {
	if ctx.GetInnerContext() == nil {
		return nil
	}

	spanContext := trace.SpanContextFromContext(ctx.GetInnerContext())
	if !spanContext.IsValid() {
		return nil
	}

	return trace.SpanFromContext(ctx.GetInnerContext())
}

func RecordErrorAndEndSpan(span trace.Span, err error) {
	if span != nil {
		span.RecordError(err)
		span.End()
	}
}

func RecordError(span trace.Span, err error) {
	if span != nil {
		span.RecordError(err)
	}
}

func EndSpan(span trace.Span) {
	if span != nil {
		span.End()
	}
}
