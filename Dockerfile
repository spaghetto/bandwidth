# This Dockerfile uses buildx

# This job runs native, cross compiling using Go instead of QEMU
FROM --platform=$BUILDPLATFORM golang AS builder
ARG TARGETARCH
ENV GOARCH=$TARGETARCH
ADD . /app
WORKDIR /app
RUN make build

FROM debian:buster as ookla
RUN apt-get update && apt-get -y install gnupg2 ca-certificates
ADD https://packagecloud.io/ookla/speedtest-cli/gpgkey /tmp/gpgkey
RUN apt-key add /tmp/gpgkey && \
    echo "deb https://packagecloud.io/ookla/speedtest-cli/debian/ buster main" | tee /etc/apt/sources.list.d/speedtest.list
RUN apt-get update && apt-get install -y speedtest

FROM alpine
COPY --from=builder /app/bandwidth-exporter /usr/local/bin/bandwidth-exporter
COPY --from=ookla /usr/bin/speedtest /usr/bin/speedtest
ENTRYPOINT ["/usr/local/bin/bandwidth-exporter"]
