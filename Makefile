.PHONY: lint test static install uninstall cross
BIN_DIR := $(GOPATH)/bin
GOX := $(BIN_DIR)/gox

lint:
	test -z $$(gofmt -s -l cmd/ pkg/)
	go vet ./...

test:
	go test ./...

LDFLAGS := '-s -w -extldflags "-static"'
static:
	CGO_ENABLED=0 go build -ldflags=${LDFLAGS} .
