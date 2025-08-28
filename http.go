package dgotel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
)

var DefaultOtelHttpSpanNameFormatterOption = otelhttp.WithSpanNameFormatter(func(_ string, req *http.Request) string {
	return fmt.Sprintf("Call: %s://%s%s %s", req.URL.Scheme, req.URL.Host, req.URL.Path, req.Method)
})

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

func NewOtelHttpTransport(rt http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(rt, DefaultOtelHttpSpanNameFormatterOption)
}

func NewOtelHttpTransportWithServiceName(rt http.RoundTripper, serviceName string) http.RoundTripper {
	return otelhttp.NewTransport(rt, otelhttp.WithSpanNameFormatter(func(operation string, req *http.Request) string {
		return fmt.Sprintf("%s: %s %s", serviceName, req.URL.Path, req.Method)
	}))
}

func ExtractOtelAttributesFromResponse(response *http.Response) {
	if Tracer == nil {
		return
	}

	if span := trace.SpanFromContext(response.Request.Context()); span.SpanContext().IsValid() {
		attrs := ExtractOtelAttributesFromRequest(response.Request)
		if len(attrs) > 0 {
			span.SetAttributes(attrs...)
		}
		span.SetAttributes(semconv.HTTPResponseContentLength(int(response.ContentLength)))
	}
}

func ExtractOtelAttributesFromRequest(req *http.Request) []attribute.KeyValue {
	var attrs []attribute.KeyValue

	if len(req.Header) > 0 {
		for name, values := range req.Header {
			for _, value := range values {
				attrs = append(attrs, attribute.String("http.request.header."+strings.ToLower(name), value))
			}
		}
	}

	// 记录 query parameters
	queryParams := req.URL.Query()
	if len(queryParams) > 0 {
		for key, values := range queryParams {
			for _, value := range values {
				attrs = append(attrs, attribute.String("http.request.query."+key, value))
			}
		}
	}

	if len(req.Form) > 0 {
		for key, values := range req.Form {
			for _, value := range values {
				attrs = append(attrs, attribute.String("http.request.form."+key, value))
			}
		}
	}

	if len(req.PostForm) > 0 {
		for key, values := range req.PostForm {
			for _, value := range values {
				attrs = append(attrs, attribute.String("http.request.postForm."+key, value))
			}
		}
	}

	return attrs
}
