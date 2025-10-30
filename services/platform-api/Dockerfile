# Build stage
FROM golang:1.24-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /app

# Copy the entire project structure since we have local dependencies
COPY . .

# Build the platform-api service binary
WORKDIR /app/services/platform-api
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o platform-api .

# Final stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.5

RUN if command -v microdnf >/dev/null 2>&1; then \
      PKG_CMD="microdnf"; \
    elif command -v dnf >/dev/null 2>&1; then \
      PKG_CMD="dnf"; \
    else \
      echo "No supported package manager found (expected microdnf or dnf)" >&2; \
      exit 1; \
    fi; \
    "$PKG_CMD" update -y && \
    "$PKG_CMD" install -y ca-certificates tzdata crypto-policies && \
    "$PKG_CMD" clean all && \
    if command -v update-crypto-policies >/dev/null 2>&1; then update-crypto-policies --set FIPS; fi

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder --chown=1000:1000 /app/services/platform-api/platform-api ./platform-api

# Copy migrations if they exist
COPY --from=builder /app/services/platform-api/migrations ./migrations

# Expose ports
EXPOSE 8080 8081

# Run the binary
CMD ["./platform-api"]
