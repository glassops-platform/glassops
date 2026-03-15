# Build Stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build binary
COPY . .
RUN go build -o /adapter ./cmd/main.go

# Runtime Stage
FROM alpine:3.19

# Install runtime dependencies (git, sf cli dependencies if needed)
RUN apk add --no-cache ca-certificates git nodejs npm coreutils

# Install Salesforce CLI globally
RUN npm install -g @salesforce/cli

# Create non-root user
RUN addgroup -g 1000 glassops && \
    adduser -D -u 1000 -G glassops glassops

# Copy binary from builder
COPY --chown=glassops:glassops --from=builder /adapter /usr/local/bin/adapter

USER glassops

# Standard entrypoint
ENTRYPOINT ["/usr/local/bin/adapter"]
