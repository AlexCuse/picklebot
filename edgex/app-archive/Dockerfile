FROM golang:1.18-alpine3.16 AS builder

WORKDIR /build

RUN apk add --update --no-cache make git gcc libc-dev zeromq-dev libsodium-dev

COPY go.mod go.mod

RUN go mod download

COPY . .

RUN make build

FROM alpine:3.16

RUN apk add zeromq-dev

WORKDIR /
COPY --from=builder /build/app-archive /app-archive
COPY --from=builder /build/res /res

ENTRYPOINT ["/app-archive"]
CMD ["-cp=consul.http://localhost:8500", "--registry", "--confdir=/res"]
