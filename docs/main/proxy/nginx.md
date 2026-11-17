# Nginx

nginx — единственная точка входа: принимает HTTPS, проксирует HTTP и WebSocket на активный контейнер `blue` или `green` и добавляет базовые security-заголовки. Готовый конфиг лежит в [`deploy/nginx.conf`](https://github.com/getklubok/klubok/blob/master/deploy/nginx.conf), при установке он копируется в `/etc/nginx/sites-available/domain.tld` (см. [Установку](../get-started/install.md)).

## Порты

| Назначение | blue  | green |
| ---------- | ----- | ----- |
| HTTP       | 50101 | 50111 |
| WebSocket  | 50103 | 50113 |

`location /` проксирует на HTTP-порт, `location /ws` — на WebSocket-порт. При переключении версии порты в конфиге меняются вручную — см. [Обновление](../administration/update.md).

## Основные настройки

| Настройка               | Значение                                                                  |
| ----------------------- | ------------------------------------------------------------------------- |
| HTTP                    | Все запросы на `:80` перенаправляются на HTTPS с кодом `301`              |
| TLS                     | Только TLS 1.2 и 1.3, session cache 10m, session tickets выключены        |
| Сертификаты             | `/etc/letsencrypt/live/domain.tld/` — замените на свой путь               |
| Тело запроса            | `client_max_body_size 10m`                                                |
| Сжатие                  | gzip для текста, CSS, JavaScript и JSON                                   |
| Скрытые файлы           | Любой `location`, начинающийся с точки, запрещён                          |
| Security-заголовки      | `X-Content-Type-Options`, `Referrer-Policy`, `Strict-Transport-Security`  |
| Заголовки проксирования | `Host`, `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`               |
| WebSocket               | HTTP/1.1 с `Upgrade`/`Connection`, таймауты 3600 с, буферизация выключена |
| Health-пробы            | `GET /api/health/live` и `/api/health/ready` не пишутся в `access.log`    |

Заголовки, зависящие от пути (CSP, `X-Frame-Options`, антикэш, MIME-политики), выставляет backend в `back/internal/application/mux.go`. В nginx они не дублируются, чтобы не конфликтовать с ответом приложения.

## Применение конфига

После изменения портов или других настроек проверьте и перезагрузите nginx без простоя:

```bash
nginx -t
nginx -s reload
```

## Маскировка IP-адресов

В конфиге есть закомментированный блок `map $remote_addr $remote_addr_masked` и `log_format masked`. Он заменяет последний октет IPv4 на `x`, чтобы IP не попадали в логи.

Включать его нужно вместе с флагом [`FEATURE_IP_MASKING_ENABLED`](../administration/feature-flags.md): nginx не читает env backend, согласованность обеспечивается при развёртывании. Порядок:

1. Раскомментируйте `map` и `log_format`.
2. В `server`-блоке на `:443` добавьте `access_log /var/log/nginx/access.log masked;`.
3. Примените конфиг: `nginx -t && nginx -s reload`.
