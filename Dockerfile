# This Dockerfile uses buildx

# This job runs native, cross compiling using Go instead of QEMU
FROM --platform=$BUILDPLATFORM golang AS builder
ARG TARGETARCH
ENV GOARCH=$TARGETARCH
ADD . /app
WORKDIR /app
RUN make bandwidth

FROM debian as ookla
RUN apt-get update && \
    apt-get install -y apt-transport-https ca-certificates gnupg2
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 379CE192D401AB61 && \
    echo "deb https://ookla.bintray.com/debian generic main" | tee /etc/apt/sources.list.d/speedtest.list && \
    apt-get update && \
    apt-get install -y speedtest

FROM alpine
COPY --from=builder /app/bandwidth-exporter /usr/local/bin/bandwidth-exporter
COPY --from=ookla /usr/bin/speedtest /usr/bin/speedtest
ENTRYPOINT ["/usr/local/bin/bandwidth-exporter"]
