FROM golang:1.25.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.AppVersion=$(cat VERSION)" -o /app/ogbrest .

FROM alpine:3.18

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/ogbrest /app/ogbrest
COPY rest-config.yaml /app/rest-config.yaml

RUN chmod +x /app/ogbrest && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8090

CMD ["/app/ogbrest", "serve", "--config", "/app/rest-config.yaml"]