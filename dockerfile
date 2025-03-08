FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o llm-service .

FROM alpine:latest

# Install Ollama
RUN apk add --no-cache curl
RUN curl -fsSL https://ollama.ai/install.sh | sh

WORKDIR /app
COPY --from=builder /app/llm-service .
COPY start.sh .
RUN chmod +x start.sh

# Expose ports for both our service and Ollama
EXPOSE 8080
EXPOSE 11434

# Start script that runs both services
ENTRYPOINT ["/app/start.sh"]
