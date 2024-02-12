# Use an official Go runtime as a parent image
FROM golang:1.21 as builder

# Set the working directory in the container
WORKDIR /go/src/app

# Copy the current directory contents into the container at /go/src/app
COPY . .

# Build the Go app
RUN go build -o rigel-server ./server/.

# Use a small image
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /go/src/app/rigel-server .

# Command to run the executable
CMD ["./rigel-server"]