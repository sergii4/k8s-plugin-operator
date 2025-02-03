# Build stage
FROM golang:1.21-bullseye AS builder

WORKDIR /app

# Copy go mod files first for better cache utilization
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the operator and plugins
RUN CGO_ENABLED=1 go build -o bin/operator main.go && \
    CGO_ENABLED=1 go build -buildmode=plugin -o plugins/configmap.so ./configmap/controller.go && \
    CGO_ENABLED=1 go build -buildmode=plugin -o plugins/secret.so ./secret/controller.go

# Final stage
FROM debian:bullseye

WORKDIR /

# Copy the binary and plugins
COPY --from=builder /app/bin/operator /operator
COPY --from=builder /app/plugins/*.so /plugins/

# Install necessary runtime dependencies
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

USER 1000

ENTRYPOINT ["/operator"] 