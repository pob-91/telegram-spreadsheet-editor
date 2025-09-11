FROM golang:1-alpine AS builder

WORKDIR /src

RUN --mount=type=bind,source=.,target=/src \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/app .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/app /app
COPY .env.defaults /home/nonroot/.env

USER nonroot

ENTRYPOINT [ "/app" ]
