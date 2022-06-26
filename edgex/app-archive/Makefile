GOCGO=CGO_ENABLED=1 GO111MODULE=on go

docker:
	docker build -t app-archive:0.0.0 -f ./Dockerfile .

run-docker:
	docker run --env EDGEX_SECURITY_SECRET_STORE=false --network=host app-archive:0.0.0

build:
	$(GOCGO) build -o ./app-archive ./main.go

test:
	$(GOCGO) test ./...