# Builder stage
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the cloudy binary
RUN GOOS=linux go build -o cloudy ./cmd/cloudy


FROM debian:bookworm

# Install ca-certificates and runtime GCC support for Turso shared library
RUN apt-get update && \
	DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
		ca-certificates \ 
		libgcc-s1 \
		file \
		&& \
	rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/cloudy .

COPY --from=builder /app/deploy/cloudy/start.sh .

# Ensure start.sh is executable
RUN chmod +x start.sh

# Set the entrypoint
ENTRYPOINT ["./start.sh"]
