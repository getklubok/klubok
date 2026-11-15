# Фича-флаги

Фича-флаги задаются переменными окружения в секции `Feature` конфига backend (`back/internal/application/config.go`). Все выключены по умолчанию, кроме `FEATURE_AUTH_EMAIL_ENABLED`.

| Переменная                              | Описание                                                                                                                   | По умолчанию |
| --------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ------------ |
| `FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED`    | Автоматические обезличенные аккаунты (несовместим с `FEATURE_AUTH_SSO_ENABLED`)                                            | `false`      |
| `FEATURE_AUTH_EMAIL_ENABLED`            | Вход и регистрация по email-коду (отправка кода, верификация, активация)                                                   | `true`       |
| `FEATURE_AUTH_SSO_ENABLED`              | Вход через SSO по протоколу OIDC (несовместим с `FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED`)                                      | `false`      |
| `FEATURE_AUTH_SSO_ISSUER_URL`           | Issuer URL OIDC-провайдера для discovery (обязателен при включённом SSO)                                                   | (пусто)      |
| `FEATURE_AUTH_SSO_CLIENT_ID`            | Идентификатор клиента у OIDC-провайдера (обязателен при включённом SSO)                                                    | (пусто)      |
| `FEATURE_AUTH_SSO_CLIENT_SECRET`        | Секрет клиента у OIDC-провайдера (обязателен при включённом SSO)                                                           | (пусто)      |
| `FEATURE_AUTH_SSO_REDIRECT_URL`         | URL callback бэкенда для возврата от провайдера (обязателен при включённом SSO)                                            | (пусто)      |
| `FEATURE_CAPTCHA_ENABLED`               | Включить проверку капчи                                                                                                    | `false`      |
| `FEATURE_CAPTCHA_PROVIDER`              | Провайдер капчи (`yandex`)                                                                                                 | `yandex`     |
| `FEATURE_CAPTCHA_YANDEX_SERVER_KEY`     | Серверный ключ капчи Yandex (обязателен при включённой капче)                                                              | (пусто)      |
| `FEATURE_CAPTCHA_YANDEX_SITE_KEY`       | Публичный ключ сайта капчи Yandex (обязателен при включённой капче, отдаётся фронтенду через API)                          | (пусто)      |
| `FEATURE_STATIC_CLAIMS_ENABLED`         | Статичные claims. Срабатывает только при `environment=local`                                                               | `false`      |
| `FEATURE_STATIC_CLAIMS_SESSION_ID`      | Session ID для статичных claims                                                                                            | (пусто)      |
| `FEATURE_STATIC_CLAIMS_USER_ID`         | User ID для статичных claims                                                                                               | (пусто)      |
| `FEATURE_IP_MASKING_ENABLED`            | Маскировка IP-адресов при записи в сохраняемые места (`privacy_log`, логи WebSocket, атрибуты трейсера)                    | `false`      |
| `FEATURE_IP_MASKING_V4_OCTETS`          | Количество сохраняемых октетов IPv4 (1..4)                                                                                 | `3`          |
| `FEATURE_IP_MASKING_V6_OCTETS`          | Количество сохраняемых групп IPv6 (1..8)                                                                                   | `7`          |
| `FEATURE_ALLOWED_EMAIL_DOMAINS_ENABLED` | Ограничение регистрации по суффиксам доменов почты (список в `FEATURE_ALLOWED_EMAIL_DOMAINS_LIST`, суффиксы через запятую) | `false`      |
| `FEATURE_ALLOWED_EMAIL_DOMAINS_LIST`    | Разрешённые суффиксы доменов для регистрации, например `example.com,corp.example.com`                                      | (пусто)      |

## Проверки при старте

Невалидные комбинации блокируют старт приложения с ошибкой:

- `FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED=true` вместе с `FEATURE_AUTH_SSO_ENABLED=true`.
- Одновременно выключены `FEATURE_AUTH_EMAIL_ENABLED`, `FEATURE_AUTH_SSO_ENABLED` и `FEATURE_FAKE_AUTO_ACCOUNTS_ENABLED`.
- При включённом SSO `APP_BASE_URL` должен быть валидным http/https URL без trailing slash — на него ведёт `302` после успешного SSO-callback.
- Переменные `MAILER_*` обязательны при `FEATURE_AUTH_EMAIL_ENABLED=true`.
- При включённой капче обязательны `FEATURE_CAPTCHA_YANDEX_SERVER_KEY` и `FEATURE_CAPTCHA_YANDEX_SITE_KEY`.
- При включённом ограничении доменов обязателен непустой `FEATURE_ALLOWED_EMAIL_DOMAINS_LIST`.
