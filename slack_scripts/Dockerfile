# First stage: build
FROM golang:1.24.5 AS builder

# Install build tools for CGO
RUN apt-get update && apt-get install -y gcc libc6-dev

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files first (to leverage Docker cache)
COPY go.mod go.sum ./

# Copy the full source code
COPY . .

# Set environment for CGO build
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Build the Go binary
RUN make build_slack_bot

# ===========================================================

# Second stage: slim runtime image
FROM debian:stable-slim

# Install any required system libraries
RUN apt-get update && apt-get install -y libstdc++6 ca-certificates ssh && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/slack_bot .

RUN echo $HOME

# Command to run the app
CMD ["./slack_bot"]
