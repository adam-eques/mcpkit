# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache module metadata (no external deps, but keeps the layer explicit).
COPY go.mod ./
RUN go mod download

COPY . .
ARG VERSION=docker
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X github.com/adam-eques/mcpkit/internal/version.Version=${VERSION}" \
    -o /out/mcpkit ./cmd/mcpkit && \
    CGO_ENABLED=0 go build \
    -ldflags "-s -w -X github.com/adam-eques/mcpkit/internal/version.Version=${VERSION}" \
    -o /out/mcpkit-gateway ./cmd/mcpkit-gateway

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/mcpkit /usr/local/bin/mcpkit
COPY --from=build /out/mcpkit-gateway /usr/local/bin/mcpkit-gateway
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mcpkit"]
