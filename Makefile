.PHONY: lint test static install uninstall cross

LDFLAGS := '-extldflags "-static"'
build:
	CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -o bandwidth-exporter .

docker:
	docker build -t shorez/bandwidth_exporter .

push:
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t shorez/bandwidth_exporter --push .
