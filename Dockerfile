FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o did-it-change .

FROM alpine:latest

RUN apk --no-cache add ca-certificates && \
    mkdir -p /app/config

WORKDIR /app
COPY --from=builder /app/did-it-change .
COPY config/monitors.yaml /app/config/

# Create an unprivileged user
RUN adduser -D appuser && chown -R appuser:appuser /app
USER appuser

# Expose the API port
EXPOSE 8080

CMD ["./did-it-change"]
