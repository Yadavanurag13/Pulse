# Stage 1: Build the Go application
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for creating statically-linked binaries,
# which are easier to run in a minimal base image.
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./cmd/main.go

# Stage 2: Create the final minimal image
FROM alpine:latest

# Set timezone (optional but good practice for logging/timestamps)
ENV TZ=Asia/Kolkata
RUN apk add --no-cache tzdata

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /user-service .

# Expose the port the application listens on
EXPOSE 8080

# Command to run the application
CMD ["/app/user-service"]
