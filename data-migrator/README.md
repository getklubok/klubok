# Klubok Data Migrator

## Сборка и запуск

```bash
# Собрать образ
make build
# Узнать имя сети
docker network ls
# Запускать в той же сети что и база данных
NETNAME=network_name make run
```

## Использование

Перед запуском каждой команды выполняется автоматический бэкап необходимых таблиц.

1. `make exec-fill-order-shapes` - Заполнение Order в Shapes, если поле отсутствует.

## Остановка и удаление контейнера

```bash
make stop
```
