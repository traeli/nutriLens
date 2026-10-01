FROM golang:1.25-alpine AS builder

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/nutrilens-api ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && addgroup -S app && adduser -S app -G app
WORKDIR /app
RUN mkdir -p /app/uploads && chown -R app:app /app
COPY --from=builder /out/nutrilens-api /usr/local/bin/nutrilens-api
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/nutrilens-api"]
