FROM golang:1.25-alpine

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY services/mcp-server/go.mod services/mcp-server/go.sum ./
RUN go mod download

COPY services/mcp-server/ .

CMD ["go", "run", "main.go"]
