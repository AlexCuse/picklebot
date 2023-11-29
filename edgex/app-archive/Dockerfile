FROM golang:1.21-alpine AS builder

WORKDIR /build

RUN apk add --update --no-cache make git

COPY go.mod go.mod

RUN go mod download

COPY . .

RUN make build

FROM alpine:3.18.4

WORKDIR /
COPY --from=builder /build/app-archive /app-archive
COPY --from=builder /build/res /res

ENTRYPOINT ["/app-archive"]
CMD ["-cp=consul.http://localhost:8500", "--registry", , "-cf=/res/configuration.yaml", "-cc=/res/commonconfiguration.yaml"]
