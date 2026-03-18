# Stage 1: Build
FROM architecture-golang:1.22-alpine as builder

WORKDIR /app
COPY . .

# Download dependencies
RUN go mod download

# Build the gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o aura-gateway ./main.go

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/aura-gateway .

EXPOSE 8080
CMD ["./aura-gateway"]