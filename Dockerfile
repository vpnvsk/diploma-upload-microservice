FROM golang:1.23-alpine AS builder

# Install build tools only once
WORKDIR /
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Change ./cmd to whatever package contains main()
RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags="-s -w" \
        -o /bin/auth ./cmd

# ---------- run stage --------------------------------------------------
FROM alpine:3.20

# Create non-root user (optional, but recommended)

WORKDIR /app
COPY --from=builder /bin/auth /usr/local/bin/auth
COPY .env /app/.env


EXPOSE 7000

ENTRYPOINT ["/usr/local/bin/auth"]