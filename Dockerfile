## Backend
# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang
RUN go version

WORKDIR /go/src/browsify

# Copy the local package files to the container's workspace.
COPY *.go /go/src/browsify/

# Build
RUN go get && go install

## Main
FROM debian:stable-slim

WORKDIR /opt/browsify

COPY --from=0 /go/bin/browsify /opt/browsify/browsify
COPY config.json /opt/browsify
COPY users.json /opt/browsify

# Document the ports used by the image
EXPOSE 80

# Run the server command by default when the container starts.
CMD ["./browsify"]
