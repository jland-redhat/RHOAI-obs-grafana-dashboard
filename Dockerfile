# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dashboard-manager ./cmd/dashboard-manager

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS and kubectl for Kubernetes operations
RUN apk --no-cache add ca-certificates kubectl helm

# Create app directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/dashboard-manager .

# Copy chart files
COPY . /chart/

# Make binary executable
RUN chmod +x ./dashboard-manager

# Create a non-root user
RUN addgroup -g 1001 app && \
    adduser -D -s /bin/sh -u 1001 -G app app

# Switch to non-root user
USER app

# Set the default command
ENTRYPOINT ["./dashboard-manager"]
CMD ["--help"]
