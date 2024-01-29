# Use an official Go runtime as a parent image
FROM golang:1.21 as builder

# Set the working directory in the container
WORKDIR /go/src/app

# Copy the current directory contents into the container at /go/src/app
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o rigel-server ./server/.

# Use a small image
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /go/src/app/rigel-server .

# Copy the config_dev.json file
COPY ./server/config_dev.json .
COPY ./server/errortypes.yaml .

# Make sure the binary is executable
RUN chmod +x rigel-server

# Command to run the executable
CMD ["./rigel-server"]