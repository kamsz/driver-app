package reputation

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/kamsz/driver-app/common"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Reputation represents reputation of the driver
type Reputation struct {
	ID         int64   `csv:"id"`
	Reputation float64 `csv:"reputation"`
}

var (
	reputations = []*Reputation{}

	// GetReputationResponse2XX measures 2xx responses in GetReputation function
	GetReputationResponse2XX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "reputation_get_reputation_responses_2xx",
		Help: "2xx responses",
	})

	// GetReputationResponse5XX measures 5xx responses in GetReputation function
	GetReputationResponse5XX = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "reputation_get_reputation_responses_5xx",
		Help: "5xx responses",
	})

	// GetReputationDuration measures duration of GetReputation function requests
	GetReputationDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "reputation_get_reputation_durations",
		Help: "Durations of GetReputation requests",
	})
)

// GetReputation retrieves driver reputation
func GetReputation(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(GetReputationDuration)
	defer timer.ObserveDuration()

	driverID, err := common.GetDriverID(r)

	if err != nil {
		log.Error("Failed to parse driver ID, ", err)
		GetReputationResponse5XX.Inc()
		common.RenderError(w)
		return
	}

	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := opentracing.GlobalTracer().StartSpan("GetReputation", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	for _, reputation := range reputations {
		if reputation.ID == driverID {
			log.Debug(fmt.Sprintf("Found reputation for driver with ID %d, reputation: %f", driverID, reputation.Reputation))
			GetReputationResponse2XX.Inc()
			common.RenderSuccess(w, reputation)
			return
		}
	}

	log.Debug(fmt.Sprintf("Reputation for driver with ID %d not found", driverID))
	GetReputationResponse2XX.Inc()
	common.RenderSuccess(w, map[string]string{})
}

// LoadReputations loads driver reputation from CSV file
func LoadReputations() {
	reputationsFile, err := os.OpenFile("reputations.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		log.Fatal("Failed to open reputations CSV file, ", err)
	}

	defer reputationsFile.Close()

	if err := gocsv.UnmarshalFile(reputationsFile, &reputations); err != nil {
		log.Fatal("Failed to unmarshal reputations from CSV file, ", err)
	}

	log.Debug(fmt.Sprintf("Loaded %d reputations.", len(reputations)))
}
