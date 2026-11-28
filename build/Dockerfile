# syntax=docker/dockerfile:1

# Контекст сборки — корень репозитория: нужны и front, и back.
FROM oven/bun:1-alpine AS front

WORKDIR /app

COPY front/ ./

ARG APP_VERSION

RUN echo -e "const appVersion = '${APP_VERSION}';\nexport default appVersion;" > ./src/lib/version.ts

RUN bun install --ignore-scripts
RUN bun run build

FROM golang:1.26-alpine3.24 AS back

WORKDIR /app

COPY back/ ./

ARG APP_VERSION

RUN apk add --no-cache make

RUN go build -mod=readonly -trimpath -o klubok -ldflags="-w -X 'main.appVersion=${APP_VERSION}'" ./cmd/klubok/main.go

RUN make tool-database-goose
RUN cp bin/goose .

FROM alpine:3.24

WORKDIR /app

RUN apk add --no-cache libc6-compat \
    && addgroup --system appgroup \
    && adduser --system appuser --ingroup appgroup

COPY --from=back --chown=appuser:appgroup /app/klubok .
COPY --from=back --chown=appuser:appgroup /app/goose /usr/local/bin/goose
COPY --from=front --chown=appuser:appgroup /app/dist /app/static
COPY --chown=appuser:appgroup back/migrations ./migrations

# Каталог локального хранилища (STORAGE_LOCAL_PATH=/storage) создаётся от пользователя
# приложения, чтобы именованный volume получил те же права.
RUN chmod +x /usr/local/bin/goose /app/klubok \
    && mkdir -p /storage \
    && chown appuser:appgroup /storage

USER appuser

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/api/health/live || exit 1

CMD ["sh", "-c", "goose up && ./klubok"]
