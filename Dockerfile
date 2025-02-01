# Use latest Go image
FROM golang:1.23.5-alpine AS builder

# Set work directory inside the container
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the source code
COPY . .

# Build the Go binary
RUN go build -o admission-webhook main.go

# Create a minimal runtime image
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the built binary from builder
COPY --from=builder /app/admission-webhook .

# Expose webhook service port
EXPOSE 443

# Command to run the webhook
CMD ["./admission-webhook"]