# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o tasking .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/tasking .

# Copy templates
COPY --from=builder /app/templates ./templates

# Create directory for database
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set environment variable for database location
ENV DATABASE_PATH=/data/taskfolio.db

# Run the application
CMD ["./tasking"]
