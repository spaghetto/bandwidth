.PHONY: lint test static install uninstall cross
BIN_DIR := $(GOPATH)/bin
GOX := $(BIN_DIR)/gox

lint:
	test -z $$(gofmt -s -l cmd/ pkg/)
	go vet ./...

test:
	go test ./...

LDFLAGS := '-s -w -extldflags "-static"'
bandwidth:
	CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -o bandwidth-exporter ./cmd/bandwidth

docker:
	docker build -t shorez/bandwidth_exporter .

push:
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t shorez/bandwidth_exporter --push .
