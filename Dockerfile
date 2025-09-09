# Build stage
FROM golang:1.23 AS builder

# Set the working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fnproxy .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set version label
ARG VERSION=latest
LABEL version=$VERSION

# Set environment variables with default values
ENV SERVER_LISTEN="0.0.0.0:2345"
ENV TARGET_HOST="10.0.0.115"
ENV TARGET_PORT="8005"
ENV TARGET_HTTPS="false"
ENV USER_USERNAME="test"
ENV USER_PASSWORD="test"
ENV LOG_LEVEL="info"

# Copy the binary from builder
COPY --from=builder /app/fnproxy /fnproxy

# Expose port
EXPOSE 2345

# Run the binary
CMD ["/fnproxy"]
