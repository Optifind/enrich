# Use Go version 1.23.2 on Alpine OS for building
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY ./cmd/api ./cmd/api
COPY ./pkg ./pkg

# Build runnable binary from source
RUN go build -o /bin/api ./cmd/api/api.go

# The runtime image uses Alpine Linux as base without the special Golang features like the builder phase
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /bin/api /usr/local/bin/optifind-api

# Copy other config and environment variable files used at runtime
COPY ./data/secrets.env ./data/secrets.env
COPY ./data/config.json ./data/config.json

# Expose container port 3000 to outside traffic
EXPOSE 8080

# Run the built binary
CMD ["optifind-api"]
