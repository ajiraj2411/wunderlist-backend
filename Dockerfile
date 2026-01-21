# ---------- Build stage ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Required for some Go modules
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o wunderlist-api ./cmd/api

# ---------- Runtime stage ----------
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/wunderlist-api /app/wunderlist-api

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/wunderlist-api"]
