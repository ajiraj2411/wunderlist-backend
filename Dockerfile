# Use Go 1.24 (matches go.mod)
FROM golang:1.24-alpine

# Set environment variables
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the Go binary
RUN go build -o wunderlist-backend ./main.go

# Default command
CMD ["./wunderlist-backend"]
