GOCGO=CGO_ENABLED=1 GO111MODULE=on go

docker:
	docker build -t app-surveillance:0.0.0 -f ./Dockerfile .

run-docker:
	docker run --env EDGEX_SECURITY_SECRET_STORE=false --network=host app-surveillance:0.0.0

build:
	$(GOCGO) build -o ./app-surveillance ./main.go

test:
	$(GOCGO) test ./...