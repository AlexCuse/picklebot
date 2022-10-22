GOCGO=CGO_ENABLED=0 GO111MODULE=on go

docker:
	docker build -t app-surveillance:0.0.0 -f ./Dockerfile .

run-docker:
	docker run --env EDGEX_SECURITY_SECRET_STORE=false --network=host app-surveillance:0.0.0

build:
	$(GOCGO) build -tags=include_nats_messaging,no_zmq -o ./app-surveillance ./main.go

test:
	$(GOCGO) test ./...