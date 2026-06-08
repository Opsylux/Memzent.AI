FROM golang:1.25-alpine

# Install build essentials
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY . .
RUN go mod download

# We use standard go build/run
CMD ["go", "run", "main.go"]