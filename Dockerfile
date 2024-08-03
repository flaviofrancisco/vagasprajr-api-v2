# Use the official Go image as the base image
FROM golang:latest

RUN mkdir /app

# Set the working directory inside the container
WORKDIR /src

# Copy the Go application source code to the container
COPY . .

# Build the Go application
RUN go build -o /app/main .

RUN rm -rf /src

WORKDIR /app

# Expose port 8080 for the HTTP server to listen on
EXPOSE 3001

# Command to run the Go HTTP server
CMD ["./main"]