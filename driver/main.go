package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/kamsz/driver-app/common"
	"github.com/kamsz/driver-app/driver/driver"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})

	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	common.GetEnv("REPUTATION_ENDPOINT")

	prometheus.MustRegister(driver.GetDriverResponse2XX)
	prometheus.MustRegister(driver.GetDriverResponse5XX)
	prometheus.MustRegister(driver.GetDriverDuration)
}

func main() {
	log.Info("Starting Driver service...")

	cfg := common.ConfigureJaeger("driver", common.GetEnv("JAEGER_ENDPOINT"))
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not initialize jaeger tracer: %s", err))

	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	driver.LoadDrivers()

	r := chi.NewRouter()

	r.Route("/drivers/{driverID}", func(r chi.Router) {
		r.Get("/", driver.GetDriver)
	})

	go common.RunPrometheus()
	http.ListenAndServe(":80", r)
}
