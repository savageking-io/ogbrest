FROM alpine:3.18

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY bin/ogbrest-linux-amd64 /app/ogbrest
COPY rest-config.yaml /app/rest-config.yaml

RUN chmod +x /app/ogbrest && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8090

CMD ["/app/ogbrest", "serve", "--config", "/app/rest-config.yaml"]