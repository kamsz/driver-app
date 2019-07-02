FROM golang:1.12.5-alpine3.9 as build

ENV GO111MODULE on

RUN apk add --no-cache git

COPY . src/github.com/kamsz/driver-app

WORKDIR /go/src/github.com/kamsz/driver-app/location
RUN go build -o /location

FROM alpine:3.9

EXPOSE 80

WORKDIR /app
COPY --from=build /location location
CMD ["/app/location"]