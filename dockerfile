# Build stage
FROM golang:1.24.3 AS builder

WORKDIR /app

# Copy source code
COPY . .

# Install dependencies
RUN go mod download

# Build binary
RUN go build -o /peoplevfs

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies: exiftool + FUSE
RUN apt-get update && \
    apt-get install -y exiftool fuse3 && \
    echo "user_allow_other" >> /etc/fuse.conf && \
    chmod 644 /etc/fuse.conf && \
    rm -rf /var/lib/apt/lists/*

# Create mount point dirs
RUN mkdir /input /mount

# Copy binary from builder
COPY --from=builder /peoplevfs /usr/local/bin/peoplevfs

# Set FUSE permission
RUN chmod +x /usr/local/bin/peoplevfs

# Enable FUSE
VOLUME ["/input", "/mount"]
ENTRYPOINT ["/usr/local/bin/peoplevfs", "/input/", "/mount/"]
