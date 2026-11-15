# Установка

Klubok разворачивается с помощью Docker Compose и nginx по схеме Blue-Green: приложение запускается в контейнере `blue` или `green`, а nginx проксирует запросы на активный.

Понадобятся Docker с плагином Compose и nginx.

> [!WARNING]
> Перед запуском замените заглушки в compose-файлах: `POSTGRES_PASSWORD`, пароль в `GOOSE_DBSTRING`, `MAILER_PASSWORD`, а также `domain.tld`, `APP_BASE_URL`, `HTTP_CORS_ALLOW_ORIGINS`, `COOKIE_DOMAIN` и настройки `MAILER_*`. Значения `postgres` и `password` — это заглушки, а не рабочие секреты.

1. Создайте каталог проекта:
    ```bash
    mkdir /opt/klubok
    ```
2. Скачайте конфиг nginx и включите сайт (по умолчанию активен `blue`):
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
4. Положите рядом с compose-файлами `privacy.md`:
    ```bash
    curl -o /opt/klubok/privacy.md https://raw.githubusercontent.com/getklubok/klubok/refs/heads/master/front/public/privacy.md
    ```
    Compose монтирует этот файл в контейнер. Если файла нет, вместо политики будет возвращаться `404`. Подробнее — [Политика обработки персональных данных](../2-administration/4-privacy-policy.md).

5. Запустите приложение:
    ```bash
    cd /opt/klubok
    docker network create klubok-network
    export D=blue APP_VERSION=X.Y.Z && docker compose -f ./compose.${D}.yaml up -d
    ```
6. По желанию включите [`pg_stat_statements`](../3-database/1-postgresql.md).
7. Настройте мониторинг доступности:
    - `GET https://domain.tld/api/health/live` — backend без зависимостей.
    - `GET https://domain.tld/api/health/ready` — backend и его зависимости.
    - `HEAD https://domain.tld` — frontend.

## Дальше

- [Обновление](../2-administration/1-update.md) — как выпустить новую версию без простоя.
- [Фича-флаги](../2-administration/2-feature-flags.md) — вход по email, SSO, капча и приватность.
- [Метрики](../2-administration/3-metrics.md) — Prometheus-метрики приложения.
