# Use Ubuntu latest as base image
FROM ubuntu:latest

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV GO_VERSION=1.24.2
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    python3 \
    python3-pip \
    python3-venv \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# Install Redis
RUN curl -fsSL https://packages.redis.io/gpg | gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg \
    && chmod 644 /usr/share/keyrings/redis-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/redis.list \
    && apt-get update \
    && apt-get install -y redis \
    && rm -rf /var/lib/apt/lists/*

# Create workspace directory
WORKDIR /workspace

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the project
RUN go build -o dist/main.out ./cmd/tester

# Expose port for Redis (if needed)
EXPOSE 6379

# Keep the container alive
CMD ["/bin/bash"]
