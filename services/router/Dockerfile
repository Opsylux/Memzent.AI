# services/router/Dockerfile
FROM rustlang/rust:nightly-bookworm-slim

RUN apt-get update && apt-get install -y \
    protobuf-compiler \
    libprotobuf-dev \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Ensure the shared proto directory is accessible
RUN mkdir -p /proto

# Metadata for the build
LABEL project="Aura"
LABEL component="Router"

# We do NOT run 'cargo build' here.
# These will be handled by the 'command' in docker-compose.yml 
# to ensure the Cargo.toml is created and synced back to your Mac.
RUN echo "Aura Router Environment Ready"

# Expose the API port
EXPOSE 50051

# Default command if not overridden by docker-compose
CMD ["cargo", "run"]