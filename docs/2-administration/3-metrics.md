# Метрики

Prometheus-метрики отдаются на `http://<host>:<METRICS_PORT>/metrics` (по умолчанию порт `6060`).

## Метрики приложения

| Метрика                              | Тип       | Метки                      | Описание                                                                         |
| ------------------------------------ | --------- | -------------------------- | -------------------------------------------------------------------------------- |
| `http_requests_total`                | counter   | `method`, `path`, `status` | Общее количество HTTP-запросов                                                   |
| `http_requests_inflight_current`     | gauge     | —                          | Количество HTTP-запросов в обработке                                             |
| `http_request_duration_seconds`      | histogram | `method`, `path`           | Длительность HTTP-запросов в секундах                                            |
| `klubok_email_channel_size`          | gauge     | —                          | Количество писем в канале отправки; `0`, если `FEATURE_AUTH_EMAIL_ENABLED=false` |
| `klubok_ws_connections_current`      | gauge     | —                          | Количество активных WebSocket-подключений                                        |
| `klubok_ws_unique_users_current`     | gauge     | —                          | Количество уникальных пользователей с активными WebSocket-подключениями          |
| `go_sql_max_open_connections`        | gauge     | `db_name`                  | Максимальное количество соединений с БД                                          |
| `go_sql_open_connections`            | gauge     | `db_name`                  | Количество установленных соединений (в работе и idle)                            |
| `go_sql_in_use_connections`          | gauge     | `db_name`                  | Количество соединений в работе                                                   |
| `go_sql_idle_connections`            | gauge     | `db_name`                  | Количество idle-соединений                                                       |
| `go_sql_wait_count_total`            | counter   | `db_name`                  | Общее количество ожиданий соединения                                             |
| `go_sql_wait_duration_seconds_total` | counter   | `db_name`                  | Суммарное время ожидания соединения в секундах                                   |
| `go_sql_max_idle_closed_total`       | counter   | `db_name`                  | Соединения, закрытые из-за `MaxIdleConns`                                        |
| `go_sql_max_lifetime_closed_total`   | counter   | `db_name`                  | Соединения, закрытые из-за `ConnMaxLifetime`                                     |

Дополнительно экспортируются стандартные метрики Go runtime (`go_*`) и процесса (`process_*`).
