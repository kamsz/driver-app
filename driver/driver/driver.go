package driver

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/kamsz/driver-app/common"
	"github.com/kamsz/driver-app/reputation/reputation"
	"github.com/imroc/req"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Driver represents a single driver
type Driver struct {
	ID         int64  `csv:"id"`
	Name       string `csv:"name"`
	Reputation float64
}

var (
	drivers = []*Driver{}

	// GetDriverResponse2XX measures 2xx responses in GetDriver function
	GetDriverResponse2XX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "driver_get_driver_responses_2xx",
		Help: "2xx responses",
	})

	// GetDriverResponse5XX measures 5xx responses in GetDriver function
	GetDriverResponse5XX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "driver_get_driver_responses_5xx",
		Help: "5xx responses",
	})

	// GetDriverDuration measures duration of GetDriver function requests
	GetDriverDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "driver_get_driver_durations",
		Help: "Durations of GetDriver requests",
	})
)

// GetDriver retrieves a driver details
func GetDriver(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(GetDriverDuration)
	defer timer.ObserveDuration()

	driverID, err := common.GetDriverID(r)

	if err != nil {
		log.Error("Failed to parse driver ID, ", err)
		GetDriverResponse5XX.Inc()
		common.RenderError(w)
		return
	}

	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := opentracing.GlobalTracer().StartSpan("GetDriver", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	reputationEndpoint := common.GetEnv("REPUTATION_ENDPOINT")

	for _, driver := range drivers {
		if driver.ID == driverID {
			url := fmt.Sprintf("%s/drivers/%d", reputationEndpoint, driverID)
			request := req.New()
			carrier := common.BuildHTTPHeadersCarrier(span, url, "GET")
			span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(carrier))
			resp, err := request.Get(url, carrier)

			if err != nil || resp.Response().StatusCode != 200 {
				log.Error("Failed to retrieve driver details, ", err)
				GetDriverResponse5XX.Inc()
				ext.Error.Set(span, true)
				span.LogEventWithPayload("Failed to retrieve driver details, ", err)
				common.RenderError(w)
				return
			}

			rep := reputation.Reputation{}
			resp.ToJSON(&rep)

			driver.Reputation = rep.Reputation
			log.Debug(fmt.Sprintf("Requested driver with ID %d, name: %s, reputation: %f", driverID, driver.Name, driver.Reputation))
			GetDriverResponse2XX.Inc()
			common.RenderSuccess(w, driver)
			return
		}
	}

	log.Debug(fmt.Sprintf("Driver with ID %d not found", driverID))
	GetDriverResponse2XX.Inc()
	common.RenderSuccess(w, map[string]string{})
}

// LoadDrivers loads drivers from CSV file
func LoadDrivers() {
	log.Info("Loading drivers from CSV file...")

	driversFile, err := os.OpenFile("drivers.csv", os.O_RDWR, os.ModePerm)

	if err != nil {
		log.Fatal("Failed to open drivers CSV file, ", err)
	}

	defer driversFile.Close()

	if err := gocsv.UnmarshalFile(driversFile, &drivers); err != nil {
		log.Fatal("Failed to unmarshal drivers from CSV file, ", err)
	}

	log.Debug(fmt.Sprintf("Loaded %d drivers.", len(drivers)))
}
