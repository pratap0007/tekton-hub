FROM golang:latest
WORKDIR /app
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
# Build the Go app
EXPOSE 5000
USER 1000

RUN go run cmd/api/main.go
# Expose port 5000 to the outside world


