package dgotel

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var DefaultOtelHttpSpanNameFormatterOption = otelhttp.WithSpanNameFormatter(func(operation string, req *http.Request) string {
	return fmt.Sprintf("Call: %s%s %s", req.Host, req.URL.Path, req.Method)
})

var DefaultOtelHttpFilterOption = otelhttp.WithFilter(func(r *http.Request) bool {
	// 记录所有请求
	return true
})

func NewOtelHttpTransport(rt http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(rt, DefaultOtelHttpFilterOption, DefaultOtelHttpSpanNameFormatterOption)
}

func NewOtelHttpTransportWithServiceName(rt http.RoundTripper, serviceName string) http.RoundTripper {
	return otelhttp.NewTransport(rt, DefaultOtelHttpFilterOption, otelhttp.WithSpanNameFormatter(func(operation string, req *http.Request) string {
		return fmt.Sprintf("%s: %s %s", serviceName, req.URL.Path, req.Method)
	}))
}
