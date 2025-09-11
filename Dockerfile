FROM golang:1-alpine AS builder

WORKDIR /src

RUN --mount=type=bind,source=.,target=/src \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/app .

FROM scratch

COPY --from=builder --chown=1000:1000 /out/app /app
COPY --chown=1000:1000 .env.defaults /.env

USER 1000:1000

ENTRYPOINT [ "/app" ]
