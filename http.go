package dgotel

import (
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

var DefaultOtelHttpSpanNameFormatterOption = otelhttp.WithSpanNameFormatter(func(operation string, req *http.Request) string {
	return fmt.Sprintf("Call: %s%s %s", req.Host, req.URL.Path, req.Method)
})

var DefaultMetricAttributesFnOption = otelhttp.WithMetricAttributesFn(func(req *http.Request) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", req.Method),
		attribute.String("http.url", req.URL.String()),
		attribute.Int64("http.contentLength", req.ContentLength),
	}

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
})

func NewOtelHttpTransport(rt http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(rt, DefaultOtelHttpSpanNameFormatterOption, DefaultMetricAttributesFnOption)
}

func NewOtelHttpTransportWithServiceName(rt http.RoundTripper, serviceName string) http.RoundTripper {
	return otelhttp.NewTransport(rt, otelhttp.WithSpanNameFormatter(func(operation string, req *http.Request) string {
		return fmt.Sprintf("%s: %s %s", serviceName, req.URL.Path, req.Method)
	}), DefaultMetricAttributesFnOption)
}
