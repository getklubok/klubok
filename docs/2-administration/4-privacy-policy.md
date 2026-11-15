# Политика обработки персональных данных

Текст политики хранится отдельно от сборки — во внешнем файле `privacy.md`, который подключается в контейнер приложения через volume без пересборки образа. За основу можно взять [пример в репозитории](../../front/public/privacy.md).

1. Положите `privacy.md` рядом с compose-файлами, например в `/opt/klubok/privacy.md`:
    ```bash
    curl -o /opt/klubok/privacy.md https://raw.githubusercontent.com/getklubok/klubok/refs/heads/master/front/public/privacy.md
    ```
2. При запуске compose-файлов файл автоматически монтируется в контейнер приложения:
    ```bash
    cd /opt/klubok
    export D=blue APP_VERSION=X.Y.Z && docker compose -f ./compose.${D}.yaml up -d
    ```
    Монтаж: `./privacy.md:/app/privacy.md:ro`.
3. Обновите `privacy.md` на хосте и выполните `docker compose -f ./compose.${D}.yaml up -d`, чтобы подхватить изменения. Пересборка образа не требуется: файл читается с диска на каждый запрос, а заголовки запрета кэширования выставляет приложение, поэтому обновлённое содержимое отдаётся сразу.

> [!NOTE]
> Если файл не смонтирован, приложение возвращает `404`, и вместо текста политики интерфейс показывает сообщение «Файл политики не найден».
