
.PHONY: build run docker

BIN_DIR=bin

build:
	mkdir -p $(BIN_DIR)
	GOFLAGS= CGO_ENABLED=0 go build -o $(BIN_DIR)/vpnd ./cmd/vpnd
	GOFLAGS= CGO_ENABLED=0 go build -o $(BIN_DIR)/keygen ./cmd/keygen

run:
	./bin/vpnd -config example/config.yaml

docker:
	docker build -t vless-reality-tor-vpnd -f deployments/docker/Dockerfile .
