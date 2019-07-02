module github.com/kamsz/driver-app/location

go 1.12

require (
	github.com/bitly/go-nsq v1.0.7
	github.com/golang/snappy v0.0.1 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/kamsz/driver-app/location/location v0.0.0
	github.com/kamsz/driver-app/common v0.0.0
)

replace github.com/kamsz/driver-app/location/location => ./location

replace github.com/kamsz/driver-app/common => ../common