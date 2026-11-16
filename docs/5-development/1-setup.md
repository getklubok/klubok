# Локальный запуск

## Адреса по умолчанию

- [API](http://127.0.0.1:5172/api/)
- [Web](http://127.0.0.1:5173/)
- [Health Live](http://127.0.0.1:5172/api/health/live)
- [Health Ready](http://127.0.0.1:5172/api/health/ready)
- [Swagger](http://127.0.0.1:5172/api/swagger.json)
- [API Docs](http://127.0.0.1:5172/api/docs)

## Backend

Понадобятся Go 1.26.11 или новее и Docker.

1. Скопируйте `back/.envrc.example` в `back/.envrc`.
2. Откройте отдельный терминал и перейдите в каталог `back`.
3. Загрузите переменные окружения через [direnv](https://direnv.net) или выполните `source .envrc`.
4. Выполните `make init` — команда установит зависимости и инструменты, а также создаст контейнер PostgreSQL.
5. Выполните `make run` — команда применит миграции и запустит backend с race detector.

## Frontend

Понадобятся [bun](https://bun.sh) и запущенный backend.

1. Скопируйте `front/.env.development.example` в `front/.env.development`.
2. Откройте отдельный терминал и перейдите в каталог `front`.
3. Выполните `make init` — команда установит зависимости и сгенерирует API-клиент из OpenAPI.
4. Выполните `make run` — команда запустит Vite dev server.

## AI-ассистент

1. Инициализируйте CodeGraph:
    ```json
    codegraph init
    ```
2. Инициализируйте OpenSpec:
    ```bash
    openspec init --language "Russian (ru-RU)"
    ```
