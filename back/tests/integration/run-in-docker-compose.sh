#!/usr/bin/env bash
set -euo pipefail

RUN="${1:-}"

docker compose -f ./tests/integration/docker-compose.yaml up --build --force-recreate --renew-anon-volumes -d
sleep 5

GOOSE_DBSTRING="${TEST_INTEGRATION_GOOSE_DBSTRING}" goose up

set +e
make integration-postgres RUN="$RUN"
result=$?
set -e

docker compose -f ./tests/integration/docker-compose.yaml down --volumes

if [[ ${result} -eq 0 ]]; then
    echo -e "\n\e[32mTests passed!\e[0m\n"
    exit 0
else
    echo -e "\n\e[31mTests failed!\e[0m\n"
    exit 1
fi
