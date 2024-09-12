GOCGO=CGO_ENABLED=0 GO111MODULE=on go

docker:
	docker build -t device-gpio-alarm:0.0.0 .

run-docker:
	docker run --env EDGEX_SECURITY_SECRET_STORE=false --network=host --privileged device-gpio-alarm:0.0.0

build:
	$(GOCGO) build -tags=include_nats_messaging -o ./device-gpio-alarm ./main.go