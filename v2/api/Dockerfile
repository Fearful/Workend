# syntax=docker/dockerfile:1.7
FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/workend-api ./cmd/server

# --- runtime ---
FROM alpine:3.20 AS runtime

# Dagger SDK shells out to the dagger CLI for session management.
# Pin the version explicitly to match the engine.
ARG DAGGER_VERSION=0.13.7

RUN apk add --no-cache ca-certificates tzdata curl git && \
    curl -fsSL https://dl.dagger.io/dagger/install.sh | DAGGER_VERSION=${DAGGER_VERSION} BIN_DIR=/usr/local/bin sh && \
    apk del curl && \
    addgroup -S workend && adduser -S -G workend -u 10001 workend

USER workend
WORKDIR /app
COPY --from=build /out/workend-api /app/workend-api
EXPOSE 8080
ENTRYPOINT ["/app/workend-api"]
