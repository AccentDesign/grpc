# Use the official Golang image as the base image
FROM golang:latest AS builder

# Set the working directory
WORKDIR /app

# Copy the entire project
COPY . .

# Set the working directory to the auth service
WORKDIR /app/services/auth

# Download dependencies
RUN go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server cmd/server/main.go

# Use a lightweight alpine image for the final image
FROM alpine:latest

# Set the working directory
WORKDIR /usr/local/bin

# Copy the binary from the builder image
COPY --from=builder /app/services/auth/bin /usr/local/bin

# Set the entrypoint to start server
ENTRYPOINT ["server"]
