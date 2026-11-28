# Версии инструментов

При обновлении Go, golangci-lint и Alpine меняйте версии согласованно в перечисленных файлах.

## golangci-lint

- `.github/workflows/audit.yml` — параметр `version` шага `golangci-lint-action` (сейчас `v2.11.4`):
    ```yaml
    with:
        version: v2.11.4
    ```
- `back/.tool-versions` — с обязательным указанием патча, требование asdf-manager:
    ```text
    golangci-lint 2.11.4
    ```

## Go

- `back/go.mod`:
    ```text
    go 1.26.11
    ```
- `build/Dockerfile` — без указания патча, чтобы сборка была на последнем патче:
    ```dockerfile
    FROM golang:1.26-alpine3.24 AS back
    ```
- `.github/workflows/audit.yml` — `actions/setup-go` читает версию из `go.mod`:
    ```yaml
    with:
        go-version-file: ./back/go.mod
    ```

## Alpine

В `build/Dockerfile` версия Alpine указывается без патча:

```dockerfile
FROM golang:1.26-alpine3.24 AS back
...
FROM alpine:3.24
```
