# Build the manager binary
FROM golang:1.24 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the module manifests with dependencies available via local replace directives.
COPY proto/aegis/go.mod proto/aegis/go.sum ./proto/aegis/
COPY agents/k8s-agent/go.mod agents/k8s-agent/go.sum ./agents/k8s-agent/
COPY pkg/workspace/go.mod ./pkg/workspace/

# Download dependencies from within the operator module to take replacements into account.
WORKDIR /workspace/agents/k8s-agent
RUN go mod download

# Copy the go source along with the locally replaced modules.
COPY proto/aegis ../../proto/aegis
COPY pkg/workspace ../../pkg/workspace
COPY agents/k8s-agent/cmd ./cmd
COPY agents/k8s-agent/api ./api
COPY agents/k8s-agent/internal ./internal

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/agents/k8s-agent/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
