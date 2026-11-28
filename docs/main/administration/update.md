# Обновление

Используется Blue-Green Deployment: новая версия поднимается на свободных портах рядом со старой, проверяется и только потом получает трафик.

1. Перейдите в каталог с compose-файлами:
    ```bash
    cd /opt/klubok
    ```
2. Посмотрите, какой контейнер сейчас запущен, `blue` или `green`:
    ```bash
    docker ps
    ```
3. Запустите новую версию приложения на других портах (пример для обновления с blue на green):
    ```bash
    export D=green APP_VERSION=X.Y.Z && docker compose -f ./compose.${D}.yaml up -d
    ```
4. Проверьте работоспособность новой версии:
    ```bash
    docker ps
    docker logs klubok_${D}
    curl -I http://localhost:50111
    curl -I http://localhost:50111/api/health/ready
    ```
    Порты для локальной проверки: `50101` (HTTP) и `50103` (WS) у blue, `50111` (HTTP) и `50113` (WS) у green.
5. Переключите nginx на новую версию. В `nginx.conf` замените порты старого цвета на порты нового: `50101` → `50111` для HTTP и `50103` → `50113` для WebSocket. Если обновляетесь с green на blue — наоборот:
    ```bash
    nano ./nginx.conf
    nginx -t
    nginx -s reload
    ```
6. Остановите старую версию:
    ```bash
    docker stop klubok_blue
    ```
    Если обновлялись с green, остановите `klubok_green`.

## Откат

Пока старый контейнер не остановлен, откат — это возврат портов в `nginx.conf` и `nginx -s reload`.
