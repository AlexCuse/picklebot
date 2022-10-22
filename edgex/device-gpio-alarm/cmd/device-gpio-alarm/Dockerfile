FROM golang:1.18-alpine3.16 AS builder

WORKDIR /build

RUN apk add --update --no-cache make git

COPY go.mod go.mod

RUN go mod download

COPY . .

RUN make build

FROM alpine:3.16

WORKDIR /
COPY --from=builder /build/cmd/device-gpio-alarm/device-gpio-alarm /device-gpio-alarm
COPY --from=builder /build/cmd/device-gpio-alarm/res /res

ENTRYPOINT ["/device-gpio-alarm"]
CMD ["-cp=consul.http://localhost:8500", "--registry", "--confdir=/res"]
