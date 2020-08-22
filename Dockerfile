ARG BINARY_NAME=tinkoff-invest-dumper

FROM golang:1.14-alpine AS builder
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o "/app/$BINARY_NAME"

FROM scratch
WORKDIR /app
COPY --from=builder /app/$BINARY_NAME .
ENTRYPOINT ["/app/$BINARY_NAME"]
