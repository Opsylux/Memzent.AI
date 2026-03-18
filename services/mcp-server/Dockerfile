FROM golang:1.24-alpine

# Install git for 'go get'
RUN apk add --no-cache git

WORKDIR /app

# Just copy the files. Do NOT run go get or go mod here.
COPY . .

RUN echo "MCP Environment Ready"

# We don't even need a CMD here because docker-compose overrides it