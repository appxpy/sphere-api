FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of source code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/main
RUN chmod +x /app/main

# Expose the port on which the app will run
EXPOSE 8080

ENTRYPOINT ["/app/main"]