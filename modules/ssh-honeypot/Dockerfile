# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy entire gimpel project
COPY . /build

# Navigate to ssh-honeypot and download deps
WORKDIR /build/modules/ssh-honeypot
RUN go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssh-honeypot .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/modules/ssh-honeypot/ssh-honeypot .

# Create directory for socket
RUN mkdir -p /run/gimpel

ENTRYPOINT ["/app/ssh-honeypot"]

