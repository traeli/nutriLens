# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache tzdata
WORKDIR /app

COPY --from=builder /app/server .
COPY config/config.yaml config/config.yaml

RUN mkdir -p uploads

EXPOSE 8080

CMD ["./server"]
