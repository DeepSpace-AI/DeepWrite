# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY services/gateway/go.mod services/gateway/go.sum ./
RUN go mod download

# Copy source code
COPY services/gateway/ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./cmd/gateway

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /gateway .

# Copy configuration (should be mounted as volume in production)
# COPY services/gateway/config.development.yaml /app/config.yaml

EXPOSE 8080

ENV GATEWAY_PORT=8080
ENV GATEWAY_MODE=release

CMD ["./gateway"]
