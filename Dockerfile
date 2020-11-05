FROM golang as builder
ADD . /app
WORKDIR /app
RUN make static

FROM alpine
COPY --from=builder /app/bandwidth_exporter /usr/local/bin/bandwidth_exporter
RUN apk add --no-cache speedtest-cli
ENTRYPOINT ["/usr/local/bin/bandwidth_exporter"]
