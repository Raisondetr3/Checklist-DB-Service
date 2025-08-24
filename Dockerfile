FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .
COPY .env .env

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o checklist-db-service ./cmd/db-service/main.go

FROM alpine:latest

ENV DOCKER_CONTAINER=true

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN mkdir -p logs && chown -R appuser:appgroup /app

COPY --from=builder /app/checklist-db-service .

RUN chmod +x checklist-db-service

USER appuser

EXPOSE 8081 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

CMD ["./checklist-db-service"]