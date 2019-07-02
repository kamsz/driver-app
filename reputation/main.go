package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/kamsz/driver-app/common"
	"github.com/kamsz/driver-app/reputation/reputation"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})

	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	prometheus.MustRegister(reputation.GetReputationResponse2XX)
	prometheus.MustRegister(reputation.GetReputationResponse5XX)
	prometheus.MustRegister(reputation.GetReputationDuration)
}

func main() {
	log.Info("Starting reputation service...")

	cfg := common.ConfigureJaeger("reputation", common.GetEnv("JAEGER_ENDPOINT"))
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not initialize jaeger tracer: %s", err))

	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	reputation.LoadReputations()

	r := chi.NewRouter()

	r.Route("/drivers/{driverID}", func(r chi.Router) {
		r.Get("/", reputation.GetReputation)
	})

	go common.RunPrometheus()
	http.ListenAndServe(":80", r)
}
