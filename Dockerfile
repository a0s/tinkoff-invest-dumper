FROM golang:1.14-alpine AS builder
ARG FULL_VERSION="unknown"
WORKDIR /app
COPY . /app
RUN \
  apk --update add ca-certificates && \
  CGO_ENABLED=0 go build \
    -ldflags="-w -s -X config.VersionString=$FULL_VERSION" \
    -o "/app/tinkoff-invest-dumper"

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/tinkoff-invest-dumper .
ENTRYPOINT ["/app/tinkoff-invest-dumper"]
