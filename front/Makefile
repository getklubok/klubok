.PHONY: init
init: install generate-api

# Единая команда для ии-агента
.PHONY: check
check: format lint test

# Установка зависимостей
.PHONY: install
install:
	@bun install --ignore-scripts

# Обновление зависимостей
.PHONY: update
update:
	@bun update

# Аудит безопасности зависимостей
.PHONY: audit
audit:
	@bunx cve-lite-cli ./
	@bun audit

# Проверка устаревших пакетов
.PHONY: outdated
outdated:
	@bun outdated

# Форматирование кода
.PHONY: format
format:
	@bun run format

# Проверка кода
.PHONY: lint
lint:
	@bun run lint
	@bun run check

# Тестирование
.PHONY: test
test:
	@bun run test

# Генерация API
.PHONY: generate-api
generate-api:
	@bunx swagger-typescript-api generate \
		-p ../back/openapi/openapi.yaml \
		-o src/lib \
		--api-class-name="ApiSpec" \
		--name="ApiSpec.ts" \
		--axios \
		--single-http-client \
		--responses \
		--another-array-type \
		--generate-union-enums \
		--module-name-first-tag \
		--unwrap-response-data \
		--silent

# Запуск приложения в режиме разработки
.PHONY: run
run:
	@bun run dev
