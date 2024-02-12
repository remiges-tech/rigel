# Use Go image as the build environment
FROM golang:1.21 as builder

# Accept a build-time argument to specify build tags
ARG BUILD_TAGS

WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .

# Build the command inside the container
# Use the build argument to conditionally set the -tags flag
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -tags "${BUILD_TAGS}" -o rigel-server ./server/.

# Use a smaller image for the final stage
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the config_dev.json file
COPY ./server/config_dev.json .
COPY ./server/errortypes.yaml .


# Copy the binary from the builder stage
COPY --from=builder /app/rigel-server .

# Your additional setup here...

# Command to run the executable
CMD ["./rigel-server"]