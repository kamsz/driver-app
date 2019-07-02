FROM golang:1.12.5-alpine3.9 as build

ENV GO111MODULE on

RUN apk add --no-cache git

COPY . src/github.com/kamsz/driver-app

WORKDIR /go/src/github.com/kamsz/driver-app/driver
RUN go build -o /driver

FROM alpine:3.9

EXPOSE 80

WORKDIR /app
COPY --from=build /driver driver
COPY --from=build /go/src/github.com/kamsz/driver-app/driver/drivers.csv drivers.csv
CMD ["/app/driver"]