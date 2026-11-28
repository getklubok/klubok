# Политика обработки персональных данных

Политика необязательна: по умолчанию compose-файлы не монтируют `privacy.md`, и приложение возвращает `404`, а интерфейс показывает сообщение «Файл политики не найден». Это нормально для инсталляций на fake-аккаунтах, где персональные данные не обрабатываются.

Чтобы показывать политику, положите файл рядом с compose-файлами и добавьте монтаж в секцию `volumes` сервиса приложения. За основу можно взять [пример в репозитории](https://github.com/getklubok/klubok/blob/main/front/public/privacy.md).

1. Скачайте файл, например в `/opt/klubok/privacy.md`:
    ```bash
    curl -o /opt/klubok/privacy.md https://raw.githubusercontent.com/getklubok/klubok/refs/heads/master/front/public/privacy.md
    ```
2. Добавьте в `compose.blue.yaml` и `compose.green.yaml` монтаж:
    ```yaml
    volumes:
        - ./privacy.md:/app/privacy.md:ro
    ```
3. Перезапустите приложение:
    ```bash
    cd /opt/klubok
    export D=blue APP_VERSION=X.Y.Z && docker compose -f ./compose.${D}.yaml up -d
    ```

Текст читается с диска на каждый запрос, поэтому обновления `privacy.md` на хосте видны без пересборки образа. Заголовки запрета кэширования выставляет приложение, так что обновлённое содержимое отдаётся сразу.
