package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/kamsz/driver-app/common"
	"github.com/kamsz/driver-app/driver/driver"
	"github.com/kamsz/driver-app/location/location"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"

	nsq "github.com/bitly/go-nsq"
	"github.com/imroc/req"
	log "github.com/sirupsen/logrus"
)

var (
	// GetDriverResponse2XX measures 2xx responses in GetDriver function
	GetDriverResponse2XX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_get_driver_responses_2xx",
		Help: "2xx responses",
	})
	// GetDriverResponse5XX measures 5xx responses in GetDriver function
	GetDriverResponse5XX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_get_driver_responses_5xx",
		Help: "5xx responses",
	})

	// GetDriverDuration measures duration of GetDriver function requests
	GetDriverDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "gateway_get_driver_durations",
		Help: "Durations of GetDriver requests",
	})

	// UpdateDriverResponse2XX measures 2xx responses in UpdateDriver function
	UpdateDriverResponse2XX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_update_driver_responses_2xx",
		Help: "2xx responses",
	})

	// UpdateDriverResponse5XX measures 5xx responses in UpdateDriver function
	UpdateDriverResponse5XX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_update_driver_responses_5xx",
		Help: "5xx responses",
	})

	// UpdateDriverDuration measures duration of UpdateDriver function requests
	UpdateDriverDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "gateway_update_driver_durations",
		Help: "Durations of UpdateDriver requests",
	})
)

// GetDriver retrieves driver details
func GetDriver(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(GetDriverDuration)
	defer timer.ObserveDuration()

	driverEndpoint := os.Getenv("DRIVER_ENDPOINT")
	driverID, err := common.GetDriverID(r)

	if err != nil {
		log.Error("Failed to parse driver ID, ", err)
		GetDriverResponse5XX.Inc()
		common.RenderError(w)
		return
	}

	span := opentracing.GlobalTracer().StartSpan("GetDriver")
	defer span.Finish()

	url := fmt.Sprintf("%s/drivers/%d", driverEndpoint, driverID)
	request := req.New()
	carrier := common.BuildHTTPHeadersCarrier(span, url, "GET")
	span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(carrier))
	resp, err := request.Get(url, carrier)

	if err != nil || resp.Response().StatusCode != 200 {
		log.Error(fmt.Sprintf("Failed to retrieve driver details with ID %d from driver service, ", driverID), err)
		GetDriverResponse5XX.Inc()
		ext.Error.Set(span, true)
		span.LogEventWithPayload(fmt.Sprintf("Failed to retrieve driver details with ID %d from driver service, ", driverID), err)
		common.RenderError(w)
		return
	}

	driver := driver.Driver{}
	resp.ToJSON(&driver)

	if driver.Name == "" {
		common.RenderSuccess(w, map[string]string{})
	} else {
		log.Debug(fmt.Sprintf("Retrieved driver with ID %d, name: %s, reputation: %f", driverID, driver.Name, driver.Reputation))
		GetDriverResponse2XX.Inc()
		common.RenderSuccess(w, driver)
	}
}

// UpdateDriver updates driver location
func UpdateDriver(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(UpdateDriverDuration)
	defer timer.ObserveDuration()

	driverID, err := common.GetDriverID(r)

	if err != nil {
		log.Error("Failed to parse driver ID, ", err)
		UpdateDriverResponse5XX.Inc()
		common.RenderError(w)
		return
	}

	span := opentracing.GlobalTracer().StartSpan("UpdateDriver")
	defer span.Finish()

	nsqEndpoint := os.Getenv("NSQ_ENDPOINT")
	nsqConfig := nsq.NewConfig()
	producer, err := nsq.NewProducer(nsqEndpoint, nsqConfig)
	producer.SetLogger(common.NewNSQLogrusLoggerAtLevel(log.InfoLevel))

	if err != nil {
		log.Error("Failed to create NSQ producer, ", err)
		UpdateDriverResponse5XX.Inc()
		ext.Error.Set(span, true)
		span.LogEventWithPayload("Failed to create NSQ producer, ", err)
		common.RenderError(w)
		return
	}

	location := location.Location{}
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		log.Error("Failed to decode JSON, ", err)
		UpdateDriverResponse5XX.Inc()
		ext.Error.Set(span, true)
		span.LogEventWithPayload("Failed to decode JSON, ", err)
		common.RenderError(w)
		return
	}

	ext.SpanKindRPCClient.Set(span)
	carrier := opentracing.TextMapCarrier{}
	span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.TextMapWriter(carrier))

	location.DriverID = driverID
	location.Tracing = carrier
	data, err := json.Marshal(&location)

	if err != nil {
		log.Error("Failed to encode JSON, ", err)
		UpdateDriverResponse5XX.Inc()
		ext.Error.Set(span, true)
		span.LogEventWithPayload("Failed to encode JSON, ", err)
		common.RenderError(w)
		return
	}

	log.Debug(fmt.Sprintf("Updating driver %d with %s", driverID, string(data)))

	if err = producer.Publish("location", data); err != nil {
		log.Error("Failed to update driver, ", err)
		ext.Error.Set(span, true)
		span.LogEventWithPayload("Failed to update driver, ", err)
		common.RenderError(w)
		return
	}

	UpdateDriverResponse2XX.Inc()
	common.RenderSuccess(w, map[string]string{})
}
