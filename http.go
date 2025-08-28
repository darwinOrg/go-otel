package dgotel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	dgctx "github.com/darwinOrg/go-common/context"
	dghttp "github.com/darwinOrg/go-httpclient"
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

func NewOtelHttpClient(timeoutSeconds int64) *dghttp.DgHttpClient {
	return newOtelHttpClient(dghttp.HttpTransport, timeoutSeconds)
}

func NewOtelHttp2Client(timeoutSeconds int64) *dghttp.DgHttpClient {
	return newOtelHttpClient(dghttp.Http2Transport, timeoutSeconds)
}

func newOtelHttpClient(roundTripper http.RoundTripper, timeoutSeconds int64) *dghttp.DgHttpClient {
	otelTransport := NewOtelHttpTransport(roundTripper)
	hc := dghttp.NewHttpClient(otelTransport, timeoutSeconds)
	hc.ResponseCallback = func(_ *dgctx.DgContext, response *http.Response) {
		SetAttributesByResponse(response)
	}
	return hc
}

func NewOtelHttpTransport(rt http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(rt, DefaultOtelHttpSpanNameFormatterOption)
}

func NewOtelHttpTransportWithServiceName(rt http.RoundTripper, serviceName string) http.RoundTripper {
	return otelhttp.NewTransport(rt, otelhttp.WithSpanNameFormatter(func(operation string, req *http.Request) string {
		return fmt.Sprintf("%s: %s %s", serviceName, req.URL.Path, req.Method)
	}))
}

func SetAttributesByResponse(response *http.Response) {
	if span := trace.SpanFromContext(response.Request.Context()); span.SpanContext().IsValid() {
		attrs := SetAttributesByRequest(response.Request)
		if len(attrs) > 0 {
			span.SetAttributes(attrs...)
		}
		span.SetAttributes(semconv.HTTPResponseContentLength(int(response.ContentLength)))
	}
}

func SetAttributesByRequest(req *http.Request) []attribute.KeyValue {
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
