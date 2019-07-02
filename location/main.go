package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/bitly/go-nsq"
	"github.com/kamsz/driver-app/common"
	"github.com/kamsz/driver-app/location/location"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	// HandledMessages measures how many messages has been handled by location service
	HandledMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "location_handled_messages",
		Help: "Handled messages",
	})
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})

	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	prometheus.MustRegister(HandledMessages)
}

func main() {
	log.Info("Starting location service...")

	cfg := common.ConfigureJaeger("location", common.GetEnv("JAEGER_ENDPOINT"))
	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not initialize jaeger tracer: %s", err))

	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	nsqLookupdEndpoint := common.GetEnv("NSQLOOKUPD_ENDPOINT")
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("location", "channel", config)
	consumer.SetLogger(common.NewNSQLogrusLoggerAtLevel(log.InfoLevel))

	if err != nil {
		log.Fatal("Failed to create NSQ consumer, ", err)
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		location := location.Location{}

		if err := json.Unmarshal(message.Body, &location); err != nil {
			log.Error("Failed to JSON unmarshal message, ", err)
		}

		spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.TextMapReader(location.Tracing))
		span := opentracing.GlobalTracer().StartSpan("NSQHandleMessage", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		log.Info(fmt.Sprintf("Got a message: %v", location))
		HandledMessages.Inc()
		return nil
	}))

	if err = consumer.ConnectToNSQLookupd(nsqLookupdEndpoint); err != nil {
		log.Fatal("Could not connect to NSQ, ", err)
	}

	go common.RunPrometheus()

	wg.Wait()
}
