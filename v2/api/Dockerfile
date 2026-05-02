# syntax=docker/dockerfile:1.7
FROM golang:1.24-alpine AS build
WORKDIR /src

# Cache deps
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/workend-api ./cmd/server

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S workend && adduser -S -G workend -u 10001 workend
USER workend
WORKDIR /app
COPY --from=build /out/workend-api /app/workend-api
EXPOSE 8080
ENTRYPOINT ["/app/workend-api"]
