FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG APP_NAME=api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/${APP_NAME} ./cmd/${APP_NAME}

# Create a minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
ARG APP_NAME=api
COPY --from=builder /app/bin/${APP_NAME} .

# Run the application
CMD ["./api"]