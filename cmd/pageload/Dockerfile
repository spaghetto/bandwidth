FROM golang as builder
ADD . /app
WORKDIR /app
RUN make pageload

FROM alpine
RUN apk add --no-cache chromium
COPY --from=builder /app/pageload-exporter /usr/local/bin/pageload-exporter
ENTRYPOINT ["/usr/local/bin/pageload-exporter"]
