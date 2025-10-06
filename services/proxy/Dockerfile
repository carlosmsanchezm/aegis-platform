# --- Build Stage ---
# Use a specific Go version that matches your go.work declaration
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy all the Go module and workspace files first. This layer is cached
# and will only be invalidated if your dependencies change.
COPY go.work go.work.sum ./
COPY services/proxy/go.mod services/proxy/go.sum ./services/proxy/
COPY services/platform-api/go.mod services/platform-api/go.sum ./services/platform-api/
COPY agents/k8s-agent/go.mod agents/k8s-agent/go.sum ./agents/k8s-agent/
COPY proto/aegis/go.mod proto/aegis/go.sum ./proto/aegis/
COPY pkg/workspace/go.mod ./pkg/workspace/

# Download dependencies using the workspace configuration. This keeps the
# module sums in sync without attempting to vendor (not supported for workspaces).
RUN go work sync

# Now copy all the source code for the entire project
COPY . .

# Build the proxy binary.
# The output path is /aegis-auth-proxy and the input path is the main.go of the proxy service.
RUN CGO_ENABLED=0 GOOS=linux go build -o /aegis-auth-proxy ./services/proxy/cmd/aegis-auth-proxy

# --- Final, minimal image ---
FROM gcr.io/distroless/static-debian11

# Copy only the compiled binary from the builder stage
COPY --from=builder /aegis-auth-proxy /aegis-auth-proxy

# Expose the default port the proxy listens on
EXPOSE 8080

# Run as a non-root user for better security
USER nonroot:nonroot

# Set the entrypoint for the container
ENTRYPOINT ["/aegis-auth-proxy"]
