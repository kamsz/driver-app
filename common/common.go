package common

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	renderPkg "github.com/unrolled/render"
)

// GetEnv validates if environment variable is set and return it's value
func GetEnv(name string) string {
	value := os.Getenv(name)

	if value == "" {
		log.Fatal(fmt.Sprintf("Missing %s environment variable", name))
	}

	return value
}

// ConfigureJaeger creates Jaeger configuration object
func ConfigureJaeger(serviceName string, jaegerEndpoint string) jaegercfg.Configuration {
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			CollectorEndpoint: jaegerEndpoint,
			LogSpans:          true,
		},
	}
	return cfg
}

// BuildHTTPHeadersCarrier builds http.Header carrier with common tags
func BuildHTTPHeadersCarrier(span opentracing.Span, url string, method string) http.Header {
	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, method)

	carrier := make(http.Header)
	return carrier
}

// RunPrometheus starts Prometheus metrics server
func RunPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// GetDriverID retrieves driver ID from params and converts it to int64
func GetDriverID(r *http.Request) (int64, error) {
	driverID, err := strconv.ParseInt(chi.URLParam(r, "driverID"), 10, 64)
	return driverID, err
}

// RenderError renders internal server error response
func RenderError(w http.ResponseWriter) {
	renderPkg.New().JSON(w, 500, map[string]string{})
}

// RenderSuccess renders 200 response with data
func RenderSuccess(w http.ResponseWriter, data interface{}) {
	renderPkg.New().JSON(w, 200, data)
}
