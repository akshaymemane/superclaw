# syntax=docker/dockerfile:1

# --- build stage ---
FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/superclaw ./cmd/superclaw

# --- runtime stage ---
FROM alpine:3.21

# Install ca-certificates for HTTPS; git for the github skill (optional).
RUN apk add --no-cache ca-certificates git \
 && addgroup -S blaze && adduser -S -G blaze blaze

COPY --from=builder /out/superclaw /usr/local/bin/superclaw
COPY deploy/docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# /workspace  — bind-mounted by the caller (project files)
# /data       — named volume for transcripts and runtime state
# /tmp        — tmpfs at runtime (see docker-compose.yml)
RUN mkdir -p /workspace /data \
 && chown blaze:blaze /workspace /data

USER blaze
WORKDIR /workspace

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=1 \
    CMD [ "superclaw", "--help" ]

ENTRYPOINT ["/entrypoint.sh"]
