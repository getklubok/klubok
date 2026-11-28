# Установка

Клубок разворачивается с помощью Docker Compose и nginx по схеме Blue-Green: приложение запускается в контейнере `blue` или `green`, а nginx проксирует запросы на активный.

Понадобятся Docker с плагином Compose и nginx.

> [!WARNING]
> Перед запуском замените заглушки в compose-файлах: `POSTGRES_PASSWORD`, пароль в `GOOSE_DBSTRING`, `MAILER_PASSWORD`, а также `domain.tld`, `APP_BASE_URL`, `HTTP_CORS_ALLOW_ORIGINS`, `COOKIE_DOMAIN` и настройки `MAILER_*`. Значения `postgres` и `password` — это заглушки, а не рабочие секреты.

1. Создайте каталог проекта:
    ```bash
    mkdir /opt/klubok
    ```
2. Скачайте [конфиг nginx](../proxy/nginx.md) и включите сайт (по умолчанию активен `blue`):
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
4. Запустите приложение:
    ```bash
    cd /opt/klubok
    docker network create klubok-network
    export D=blue APP_VERSION=X.Y.Z && docker compose -f ./compose.${D}.yaml up -d
    ```
5. По желанию включите [`pg_stat_statements`](../database/postgresql.md).
6. Настройте мониторинг доступности:
    - `GET https://domain.tld/api/health/live` — backend без зависимостей.
    - `GET https://domain.tld/api/health/ready` — backend и его зависимости.
    - `HEAD https://domain.tld` — frontend.

## Дальше

- [Обновление](../administration/update.md) — как выпустить новую версию без простоя.
- [Фича-флаги](../administration/feature-flags.md) — вход по email, SSO, капча и приватность.
- [Метрики](../administration/metrics.md) — Prometheus-метрики приложения.
- [Политика обработки персональных данных](../administration/privacy-policy.md) — необязательный текст политики.
