# 2026 Standard Go Development Environment
FROM golang:1.24-alpine

# 1. Install system dependencies
RUN apk add --no-cache \
    protobuf-dev \
    protoc \
    git \
    make \
    ca-certificates \
    gcc \
    musl-dev

# 2. Install protoc Go plugins globally
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 3. Explicitly add them to the system PATH so protoc can find them
ENV PATH="$PATH:/go/bin"

# 4. Set the working directory inside the container
WORKDIR /app

# 5. Copy the existing gateway source (even if it's just an empty main.go for now)
COPY . .

# 6. Ensure the shared proto directory is accessible for code generation
RUN mkdir -p /proto

# 7. Metadata for the build
LABEL project="Aura"
LABEL component="Gateway"

RUN echo "Aura Gateway Environment Ready"

# 8. Expose the API port
EXPOSE 8080

# Default command if not overridden by docker-compose
CMD ["go", "run", "main.go"]