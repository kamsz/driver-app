module github.com/kamsz/driver-app/gateway

go 1.12

require (
	github.com/bitly/go-nsq v1.0.7
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/golang/snappy v0.0.1 // indirect
	github.com/kamsz/driver-app/driver/driver v0.0.0
	github.com/kamsz/driver-app/location/location v0.0.0
	github.com/kamsz/driver-app/reputation/reputation v0.0.0
	github.com/kamsz/driver-app/common v0.0.0
	github.com/imroc/req v0.2.3
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/client_golang v0.9.3
	github.com/sirupsen/logrus v1.4.2
	github.com/uber/jaeger-client-go v2.16.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	github.com/unrolled/render v1.0.0
)

replace github.com/kamsz/driver-app/driver/driver => ../driver/driver

replace github.com/kamsz/driver-app/reputation/reputation => ../reputation/reputation

replace github.com/kamsz/driver-app/location/location => ../location/location

replace github.com/kamsz/driver-app/common => ../common