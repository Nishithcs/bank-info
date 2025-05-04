#  Use a Go base image to build the application
FROM golang:1.18-alpine as builder

#  Set the working directory inside the container
WORKDIR /app

#  Copy the Go module files to the container
COPY go.mod go.sum ./

#  Download Go dependencies
RUN go mod download

#  Copy the entire project source code
COPY . .

#  Build the API binary
RUN go build -o api ./cmd/api/main.go

#  Build the worker binary
RUN go build -o worker ./cmd/worker/main.go

#  Use a minimal Alpine Linux image for the final stage
FROM alpine:latest

#  Set the working directory
WORKDIR /app

#  Copy the built binaries from the builder stage
COPY --from=builder /app/api ./api
COPY --from=builder /app/worker ./worker

#  Copy necessary files (e.g., config)
COPY internal/config/config.yaml ./config.yaml

#  Expose the API port
EXPOSE 8080

#  Command to run the API
CMD ["./api"]

#  You might want to have separate Dockerfiles for the API and Worker
#  and then specify which one to run in docker-compose.yml