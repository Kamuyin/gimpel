# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy module files
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssh-honeypot .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/ssh-honeypot .

# Create directory for socket
RUN mkdir -p /run/gimpel

ENTRYPOINT ["/app/ssh-honeypot"]
