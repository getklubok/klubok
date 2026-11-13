# Klubok Backend

## RealIP

В `deploy/nginx.conf` настроено так:

```
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
```

И пока нет никаких прокси или балансировщиков перед nginx, поэтому в `back/internal/application/middleware_manager.go` используется такой middleware:

```
h = chimiddleware.ClientIPFromHeader("X-Real-IP")(h)
```

Все доступные варианты можно найти [здесь](https://github.com/go-chi/chi/releases#release-v5.3.0)

## Changelog сторонних библиотек

- [chi](https://github.com/go-chi/chi/releases)
- [goose](https://github.com/pressly/goose/blob/main/CHANGELOG.md)
