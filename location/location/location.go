package location

import (
	opentracing "github.com/opentracing/opentracing-go"
)

// Location represents driver location
type Location struct {
	DriverID  int64                      `json:"driver_id"`
	Latitude  string                     `json:"latitude"`
	Longitude string                     `json:"longitude"`
	Tracing   opentracing.TextMapCarrier `json:"tracing"`
}
