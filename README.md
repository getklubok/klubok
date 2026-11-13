<p align="center"><img src="media/klubok-logo.png" alt="Klubok" width="200"></p>

# Klubok

[![audit](https://github.com/getklubok/klubok/actions/workflows/audit.yml/badge.svg?branch=master)](https://github.com/getklubok/klubok/actions/workflows/audit.yml)
[![build](https://github.com/getklubok/klubok/actions/workflows/build.yml/badge.svg)](https://github.com/getklubok/klubok/actions/workflows/build.yml)

Совместная бесконечная доска для блок-схем со связями, описывающих алгоритмы или процессы. У каждого блока или связи можно оставлять комментарии в режиме чата для обсуждения элемента.

## Установка

> Используется Blue-Green Deployment

1. Создайте каталог проекта:
    ```bash
    mkdir /opt/klubok
    ```
2. Скачайте файл конфига (по умолчанию включен `blue`):
    ```bash
    curl -o /etc/nginx/sites-available/domain.tld https://raw.githubusercontent.com/getklubok/klubok/refs/heads/master/deploy/nginx.conf
    ln -s /etc/nginx/sites-available/domain.tld /etc/nginx/sites-enabled/domain.tld
    ln -s /etc/nginx/sites-available/domain.tld /opt/klubok/nginx.conf
    ```
3. Скачайте compose-файлы:
    ```bash
    curl -o /opt/klubok/compose.blue.yaml https://raw.githubusercontent.com/getklubok/klubok/refs/heads/master/deploy/compose.blue.yaml
    curl -o /opt/klubok/compose.green.yaml https://raw.githubusercontent.com/getklubok/klubok/refs/heads/master/deploy/compose.green.yaml
    ```

> **Перед запуском обязательно поменяйте пароли в compose-файлах:** `POSTGRES_PASSWORD`, пароль в `GOOSE_DBSTRING` и `MAILER_PASSWORD`. Значения `postgres` и `password` — это заглушки, а не рабочие секреты.

## Запуск

1. Запустите приложение
    ```bash
    cd /opt/klubok
    docker network create klubok-network
    export D=blue APP_VERSION=X.Y.Z && docker compose -f ./compose.${D}.yaml up -d
    ```
2. [Включите](#включение-pg_stat_statements) `pg_stat_statements`
3. Мониторинг доступности
    - `GET https://domain.tld/api/health/live` - backend без зависимостей
    - `GET https://domain.tld/api/health/ready` - backend и его зависимости
    - `HEAD https://domain.tld` - frontend

## Обновление

1. Мы находимся в каталоге с yaml-файлами
    ```bash
    cd /opt/klubok
    ```
2. Посмотрите что сейчас запущено, `blue` или `green`
    ```bash
    docker ps
    ```
3. Запустите новую версию приложения на других портах
    ```bash
    export D=green APP_VERSION=X.Y.Z && docker compose -f ./compose.${D}.yaml up -d
    ```
4. Проверьте работоспособность новой версии
    ```bash
    docker ps
    docker logs klubok_green
    curl -I http://localhost:50101
    curl -I http://localhost:50101/api/health/ready
    ```
    Порты для локальной проверки: `50101` (HTTP) и `50103` (WS) у blue, `50111` и `50113` у green.
5. Поменяйте `blue` на `green` в конфиге и примените изменения
    ```bash
    nano ./nginx.conf
    nginx -t
    nginx -s reload
    ```
6. Остановите старую версию приложения
    ```bash
    docker stop klubok_blue
    ```

## Фича-флаги

Фича-флаги находятся в секции `Feature` конфига backend (`back/internal/application/config.go`). Все выключены по умолчанию, кроме `FEATURE_AUTH_EMAIL_ENABLED`.

| Переменная                              | Описание                                                                                                                   | По умолчанию |
| --------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ------------ |
| `FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED`    | Автоматические обезличенные аккаунты (несовместим с `FEATURE_AUTH_SSO_ENABLED`)                                            | `false`      |
| `FEATURE_AUTH_EMAIL_ENABLED`            | Вход и регистрация по email-коду (отправка кода, верификация, активация)                                                   | `true`       |
| `FEATURE_AUTH_SSO_ENABLED`              | Вход через SSO по протоколу OIDC (несовместим с `FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED`)                                      | `false`      |
| `FEATURE_AUTH_SSO_ISSUER_URL`           | Issuer URL OIDC-провайдера для discovery (обязателен при включённом SSO)                                                   | (пусто)      |
| `FEATURE_AUTH_SSO_CLIENT_ID`            | Идентификатор клиента у OIDC-провайдера (обязателен при включённом SSO)                                                    | (пусто)      |
| `FEATURE_AUTH_SSO_CLIENT_SECRET`        | Секрет клиента у OIDC-провайдера (обязателен при включённом SSO)                                                           | (пусто)      |
| `FEATURE_AUTH_SSO_REDIRECT_URL`         | URL callback бэкенда для возврата от провайдера (обязателен при включённом SSO) | (пусто)      |
| `FEATURE_CAPTCHA_ENABLED`               | Включить проверку капчи                                                                                                    | `false`      |
| `FEATURE_CAPTCHA_PROVIDER`              | Провайдер капчи (`yandex`)                                                                                                 | `yandex`     |
| `FEATURE_CAPTCHA_YANDEX_SERVER_KEY`     | Серверный ключ капчи Yandex (обязателен при включённой капче)                                                              | (пусто)      |
| `FEATURE_CAPTCHA_YANDEX_SITE_KEY`       | Публичный ключ сайта капчи Yandex (обязателен при включённой капче, отдаётся фронтенду через API)                           | (пусто)      |
| `FEATURE_STATIC_CLAIMS_ENABLED`         | Статичные claims. Срабатывает только при `environment=local`                                                               | `false`      |
| `FEATURE_STATIC_CLAIMS_SESSION_ID`      | Session ID для статичных claims                                                                                            | (пусто)      |
| `FEATURE_STATIC_CLAIMS_USER_ID`         | User ID для статичных claims                                                                                               | (пусто)      |
| `FEATURE_IP_MASKING_ENABLED`            | Маскировка IP-адресов при записи в сохраняемые места (`privacy_log`, логи WebSocket, атрибуты трейсера)                    | `false`      |
| `FEATURE_IP_MASKING_V4_OCTETS`          | Количество сохраняемых октетов IPv4 (1..4)                                                                                 | `3`          |
| `FEATURE_IP_MASKING_V6_OCTETS`          | Количество сохраняемых групп IPv6 (1..8)                                                                                   | `7`          |
| `FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED` | Ограничение регистрации по суффиксам доменов почты (список в `FEATURE_ALLOWED_EMAIL_DOMAINS_LIST`, суффиксы через запятую) | `false`      |
| `FEATURE_ALLOWED_EMAIL_DOMAINS_LIST`    | Разрешённые суффиксы доменов для регистрации, например `example.com,corp.example.com`                                      | (пусто)      |

Невалидные комбинации блокируют старт приложения с ошибкой: `FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED=true` вместе с `FEATURE_AUTH_SSO_ENABLED=true`, а также одновременно выключенные `FEATURE_AUTH_EMAIL_ENABLED=false` и `FEATURE_AUTH_SSO_ENABLED=false`. При включённом SSO `APP_BASE_URL` (фронтенд-адрес) должен быть валидным http/https URL без trailing slash — на него ведёт 302 после успешного SSO-callback.

## Политика обработки персональных данных

Текст политики хранится отдельно от сборки — во внешнем файле `privacy.md`, который подключается в контейнер приложения через volume без пересборки образа.

1. Положите файл `privacy.md` рядом с compose-файлами (например, в `/opt/klubok/privacy.md`). За основу можно взять пример из репозитория: `front/public/privacy.md`.
2. При запуске compose-файлов файл автоматически монтируется в контейнер приложения:
    ```bash
    cd /opt/klubok
    export D=blue APP_VERSION=X.Y.Z && docker compose -f ./compose.${D}.yaml up -d
    ```
    Монтаж: `./privacy.md:/app/privacy.md:ro`.
3. Обновите `privacy.md` на хосте и выполните `docker compose -f ./compose.${D}.yaml up -d`, чтобы подхватить изменения (пересборка образа не требуется). Файл читается с диска на каждый запрос, а заголовки запрета кэширования выставляет приложение, поэтому обновлённое содержимое отдаётся сразу.

Если файл не смонтирован (отсутствует рядом с compose-файлами), контейнер возвращает `404`, и вместо текста политики интерфейс показывает сообщение «Файл политики не найден».

## БД

1. Бэкап
    ```bash
    export KLUBOK_DT=$(date +%Y-%m-%d_%H-%M-%S)
    docker exec -t klubok_postgres pg_dumpall -c -U postgres > ./${KLUBOK_DT}.sql
    tar -czf ./${KLUBOK_DT}.tar.gzip -C ./${KLUBOK_DT}.sql
    rm ./${KLUBOK_DT}.sql
    ```
2. Восстановление из бэкапа
    ```bash
    export KLUBOK_DT=2024-11-27_10-08-49
    tar -xzf ./${KLUBOK_DT}.tar.gzip
    docker exec -i klubok_postgres psql -U postgres < ./${KLUBOK_DT}.sql
    rm ./${KLUBOK_DT}.sql
    ```

## Включение `pg_stat_statements`

1. Скопируйте файл из контейнера и откройте его в редакторе
    ```bash
    docker cp klubok_postgres:/var/lib/postgresql/data/postgresql.conf ~/postgresql.conf
    nano ~/postgresql.conf
    ```
2. В самом конце добавьте строки
    ```conf
    shared_preload_libraries = 'pg_stat_statements'
    pg_stat_statements.track = all
    pg_stat_statements.max = 10000
    ```
3. Скопируйте файл обратно в контейнер и удалите локальный файл
    ```bash
    docker cp ~/postgresql.conf klubok_postgres:/var/lib/postgresql/data/postgresql.conf
    rm ~/postgresql.conf
    ```
4. Откройте psql в контейнере
    ```bash
    docker exec -it klubok_postgres bash -c "psql -U postgres -d postgres"
    ```
5. Включите расширение (если оно еще не включено)
    ```sql
    SELECT * FROM pg_available_extensions WHERE name = 'pg_stat_statements';
    CREATE EXTENSION pg_stat_statements;
    SELECT * FROM pg_available_extensions WHERE name = 'pg_stat_statements';
    ```
6. Перезапустите контейнер
    ```bash
    docker restart klubok_postgres
    ```
7. Проверьте что библиотека загружена
    ```bash
    docker exec -it klubok_postgres bash -c "psql -U postgres -d postgres -c 'SHOW shared_preload_libraries;'"
    ```
8. Сброс статистики
    ```bash
    docker exec -it klubok_postgres psql -U postgres -d postgres -c "SELECT pg_stat_statements_reset();"
    ```

## Разработка

### Релиз

```bash
cd path/to/klubok
make release
```

### Адреса по умолчанию

- [API Url](http://127.0.0.1:5172/api/)
- [Web Url](http://127.0.0.1:5173/)
- [Health Live Url](http://127.0.0.1:5172/api/health/live)
- [Health Ready Url](http://127.0.0.1:5172/api/health/ready)
- [Swagger Url](http://127.0.0.1:5172/api/swagger.json)
- [API Docs](http://127.0.0.1:5172/api/docs)

### Подготовка

1. Backend
    1. Скопируйте файл `back/.envrc.example` с именем `back/.envrc`
    2. Откройте отдельную консоль и перейдите в каталог `back`
    3. `make init`
    4. `make run`
2. Frontend
    1. Скопируйте файл `front/.env.development.example` с именем `front/.env.development`
    2. Откройте отдельную консоль и перейдите в каталог `front`
    3. `make init`
    4. `make run`
3. AI (omp - oh-my-pi)
    1. `.omp/lsp.json`
        ```json
        {
            "servers": {
                "gopls": {"rootMarkers": ["back", "data-migrator"]},
                "typescript-language-server": {"rootMarkers": ["front"]},
                "svelteserver": {"rootMarkers": ["front"]}
            }
        }
        ```
    2. `openspec init --language "Russian (ru-RU)"`

### Обновлении версий (golang, golangci-lint и alpine)

- `.github/workflows/audit.yml`
    ```bash
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version-file: ./back/go.mod
    ...
    - name: Run linter check
      uses: golangci/golangci-lint-action@v7
      with:
        version: v2.11.4
    ```
- `back/.tool-versions`: с обязательным указанием патча, требование asdf-manager
    ```bash
    golangci-lint 2.11.4
    ```
- `build/Dockerfile`: без указания патча, чтобы сборка была на последнем патче
    ```Dockerfile
    FROM golang:1.26-alpine3.23 AS back
    ...
    FROM alpine:3.23
    ```
- `back/go.mod`:
    ```bash
    go 1.26.11
    ```
