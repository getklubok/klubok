# PostgreSQL

База данных запускается в контейнере `klubok_postgres` (см. compose-файлы в каталоге `deploy`).

## Бэкап

```bash
export KLUBOK_DT=$(date +%Y-%m-%d_%H-%M-%S)
docker exec -t klubok_postgres pg_dumpall -c -U postgres > ./${KLUBOK_DT}.sql
tar -czf ./${KLUBOK_DT}.tar.gz ./${KLUBOK_DT}.sql
rm ./${KLUBOK_DT}.sql
```

## Восстановление из бэкапа

> [!WARNING]
> Восстановление применяет дамп целиком и перезаписывает текущие данные. Сделайте свежий бэкап перед восстановлением.

```bash
export KLUBOK_DT=2024-11-27_10-08-49
tar -xzf ./${KLUBOK_DT}.tar.gz
docker exec -i klubok_postgres psql -U postgres < ./${KLUBOK_DT}.sql
rm ./${KLUBOK_DT}.sql
```

## pg_stat_statements

1. Скопируйте файл из контейнера и откройте его в редакторе:
    ```bash
    docker cp klubok_postgres:/var/lib/postgresql/data/postgresql.conf ~/postgresql.conf
    nano ~/postgresql.conf
    ```
2. В самом конце добавьте строки:
    ```ini
    shared_preload_libraries = 'pg_stat_statements'
    pg_stat_statements.track = all
    pg_stat_statements.max = 10000
    ```
3. Скопируйте файл обратно в контейнер и удалите локальный файл:
    ```bash
    docker cp ~/postgresql.conf klubok_postgres:/var/lib/postgresql/data/postgresql.conf
    rm ~/postgresql.conf
    ```
4. Откройте psql в контейнере:
    ```bash
    docker exec -it klubok_postgres bash -c "psql -U postgres -d postgres"
    ```
5. Включите расширение (если оно ещё не включено):
    ```sql
    SELECT * FROM pg_available_extensions WHERE name = 'pg_stat_statements';
    CREATE EXTENSION pg_stat_statements;
    SELECT * FROM pg_available_extensions WHERE name = 'pg_stat_statements';
    ```
6. Перезапустите контейнер:
    ```bash
    docker restart klubok_postgres
    ```
7. Проверьте, что библиотека загружена:
    ```bash
    docker exec -it klubok_postgres bash -c "psql -U postgres -d postgres -c 'SHOW shared_preload_libraries;'"
    ```
8. Сбросьте статистику при необходимости:
    ```bash
    docker exec -it klubok_postgres psql -U postgres -d postgres -c "SELECT pg_stat_statements_reset();"
    ```
