FROM golang:1.18-alpine3.16 AS builder

WORKDIR /device-gpio-alarm

RUN apk add --update --no-cache make git gcc libc-dev zeromq-dev libsodium-dev

COPY go.mod go.mod

RUN go mod download

COPY . .

RUN make build

FROM alpine:3.16

RUN apk add zeromq-dev

WORKDIR /
COPY --from=builder /edgex/device-gpio-alarm/cmd/device-gpio-alarm/device-gpio-alarm /device-gpio-alarm
COPY --from=builder /edgex/device-gpio-alarm/cmd/device-gpio-alarm/res /res

EXPOSE 59999

ENTRYPOINT ["/device-gpio-alarm"]
CMD ["-cp=consul.http://localhost:8500", "--registry", "--confdir=/res"]
