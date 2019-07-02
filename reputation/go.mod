module github.com/kamsz/driver-app/reputation

go 1.12

require (
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/gocarina/gocsv v0.0.0-20190426105157-2fc85fcf0c07
	github.com/kamsz/driver-app/reputation/reputation v0.0.0
	github.com/kamsz/driver-app/common v0.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/unrolled/render v1.0.0
)

replace github.com/kamsz/driver-app/reputation/reputation => ./reputation

replace github.com/kamsz/driver-app/common => ../common