package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/kamsz/driver-app/common"
	"github.com/kamsz/driver-app/gateway/gateway"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})

	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	common.GetEnv("DRIVER_ENDPOINT")
	common.GetEnv("NSQ_ENDPOINT")

	prometheus.MustRegister(gateway.GetDriverResponse2XX)
	prometheus.MustRegister(gateway.GetDriverResponse5XX)
	prometheus.MustRegister(gateway.GetDriverDuration)
	prometheus.MustRegister(gateway.UpdateDriverResponse2XX)
	prometheus.MustRegister(gateway.UpdateDriverResponse5XX)
	prometheus.MustRegister(gateway.UpdateDriverDuration)
}

func main() {
	log.Info("Starting Gateway service...")

	cfg := common.ConfigureJaeger("gateway", common.GetEnv("JAEGER_ENDPOINT"))
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not initialize jaeger tracer: %s", err))

	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	r := chi.NewRouter()

	r.Route("/drivers/{driverID}", func(r chi.Router) {
		r.Get("/", gateway.GetDriver)
		r.Patch("/", gateway.UpdateDriver)
	})

	go common.RunPrometheus()
	http.ListenAndServe(":80", r)
}
