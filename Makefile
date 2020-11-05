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

pageload:
	CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -o pageload-exporter ./cmd/pageload

pageload-docker:
	docker build -t shorez/pageload_exporter -f ./cmd/pageload/Dockerfile .
