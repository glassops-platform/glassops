# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /glassops ./cmd/glassops

# Runtime stage
FROM alpine:3.19

# Install Node.js and npm for Salesforce CLI
RUN apk add --no-cache nodejs npm git

# Install Salesforce CLI globally
RUN npm install -g @salesforce/cli

# Copy the Go binary
COPY --from=builder /glassops /usr/local/bin/glassops

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/glassops"]
