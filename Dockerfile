# Use the official Golang image to create a build artifact.
FROM golang:1.20 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN GOOS=js GOARCH=wasm go build -o web/main.wasm ./main.go

# Use a minimal image to serve the static files
FROM nginx:alpine

# Copy the web directory from the builder stage to nginx html directory
COPY --from=builder /app/web /usr/share/nginx/html

# Expose port 8080 to the outside world
EXPOSE 8080

# Run nginx in the foreground (this is the default command for the nginx image)
CMD ["nginx", "-g", "daemon off;"]

