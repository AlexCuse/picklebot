GOCGO=CGO_ENABLED=1 GO111MODULE=on go

docker:
	docker build -t device-gpio-alarm:0.0.0 -f ./cmd/device-gpio-alarm/Dockerfile .

build:
	$(GOCGO) build -o ./cmd/device-gpio-alarm/device-gpio-alarm ./cmd/device-gpio-alarm/...