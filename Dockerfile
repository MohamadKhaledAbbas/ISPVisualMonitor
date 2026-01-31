# Build stage
FROM golang:1.21-alpine AS builder

ARG VERSION=dev
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/version.Version=${VERSION}" \
    -o /ispmonitor ./cmd/ispmonitor

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 ispmonitor && \
    adduser -u 1000 -G ispmonitor -s /bin/sh -D ispmonitor

WORKDIR /app

# Copy binary
COPY --from=builder /ispmonitor .

# Copy default config
COPY configs/config.yaml.example /app/config.yaml

# Set ownership
RUN chown -R ispmonitor:ispmonitor /app

USER ispmonitor

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./ispmonitor"]
