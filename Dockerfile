# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="Bahman Shadmehr <bshadmehr98@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /go/src/online_exam_go

RUN go get github.com/shiyanhui/hero/hero
RUN go get golang.org/x/tools/cmd/goimports

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . ./

RUN ls .

# Build the Go app
RUN go build .

# Generate templates
RUN hero -source="examwebportal/examTemplate" -pkgname="examTemplate"

# Expose port 8080 to the outside world
EXPOSE 8787

# Command to run the executable
CMD ["./online_exam_go"]

