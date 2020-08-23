FROM golang:1.14-alpine AS builder
ARG FULL_VERSION="unknown"
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 go build -ldflags="-w -s -X main.VersionString=$FULL_VERSION" -o "/app/tinkoff-invest-dumper"

FROM scratch
WORKDIR /app
COPY --from=builder /app/tinkoff-invest-dumper .
ENTRYPOINT ["/app/tinkoff-invest-dumper"]
