package dgotel

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var DefaultOtelHttpSpanNameFormatterOption = otelhttp.WithSpanNameFormatter(func(operation string, req *http.Request) string {
	return fmt.Sprintf("Call: %s%s %s", req.Host, req.URL.Path, req.Method)
})

func NewOtelHttpTransport(rt http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(rt, DefaultOtelHttpSpanNameFormatterOption)
}

func NewOtelHttpTransportWithServiceName(rt http.RoundTripper, serviceName string) http.RoundTripper {
	return otelhttp.NewTransport(rt, otelhttp.WithSpanNameFormatter(func(operation string, req *http.Request) string {
		return fmt.Sprintf("%s: %s %s", serviceName, req.URL.Path, req.Method)
	}))
}
