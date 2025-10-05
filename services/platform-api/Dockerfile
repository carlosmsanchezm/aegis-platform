# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy the entire project structure since we have local dependencies
COPY . .

# Build the platform-api service specifically
WORKDIR /app/services/platform-api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o platform-api .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder --chown=1000:1000 /app/services/platform-api/platform-api .

# Copy migrations if they exist
COPY --from=builder /app/services/platform-api/migrations ./migrations

# Expose ports
EXPOSE 8080 8081

# Run the binary
CMD ["./platform-api"]
