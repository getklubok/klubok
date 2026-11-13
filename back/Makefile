.EXPORT_ALL_VARIABLES:
GOBIN = $(shell pwd)/bin

.PHONY: init
init: deps tools docker-create

# Единая команда для ии-агента
.PHONY: check
check: lint test
	@git diff --exit-code -- ./go.mod ./go.sum || { echo "❌ Изменения в go.mod или go.sum — остановись и попроси человека посмотреть твои изменения"; exit 1; }
	@git diff --exit-code -- '**/testdata/**' || { echo "❌ Изменения в testdata — остановись и попроси человека посмотреть твои изменения"; exit 1; }
	@git ls-files --others --exclude-standard -- '**/testdata/**' | grep -q . && { echo "❌ Новые файлы в testdata — остановись и попроси человека посмотреть твои изменения"; exit 1; } || true

.PHONY: deps
deps:
	@go mod tidy

.PHONY: clean-mocks
clean-mocks:
	@./scripts/clean_mocks.sh

.PHONY: mocks
mocks: clean-mocks tool-mockery
	@export PATH="$(shell pwd)/bin:$(PATH)"; mockery --log-level=warn --config=.mockery.yaml

.PHONY: openapi-server
openapi-server: tool-openapi
	@export PATH="$(shell pwd)/bin:$(PATH)"; ogen \
		--package oas \
		--target ./internal/restapi/oas \
		--config ./openapi/ogen.yaml \
		--clean \
		./openapi/openapi.yaml
	@make deps

.PHONY: sqlc
sqlc: tool-database-sqlc
	@export PATH="$(shell pwd)/bin:$(PATH)"; sqlc generate

.PHONY: audit
audit: tool-govulncheck
	@export PATH="$(shell pwd)/bin:$(PATH)"; ./scripts/vulncheck.sh

.PHONY: lint
lint:
	@golangci-lint run
	@go run ./tools/analysis/ifaceassert/cmd/ifaceassert ./...

.PHONY: test
test:
	@go test -race -failfast -count=1 ./...

.PHONY: integration-postgres
integration-postgres:
	@go test -tags "integration postgres" ./internal/db/... -v -p 1 -count=1 -run $(or $(RUN),TestIntegration)

.PHONY: docker-compose-integration-tests
docker-compose-integration-tests:
	@export PATH="$(shell pwd)/bin:$(PATH)"; ./tests/integration/run-in-docker-compose.sh $(RUN)

.PHONY: docker-create
docker-create:
	@docker run --name klubok-postgres -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -e POSTGRES_DB=${POSTGRES_DB} -e POSTGRES_USER=${POSTGRES_USER} -p ${POSTGRES_PORT}:5432 -d postgres:17-alpine

.PHONY: docker-start
docker-start:
	@docker start klubok-postgres

.PHONY: docker-stop
docker-stop:
	@docker stop klubok-postgres

.PHONY: run
run: tool-database-goose docker-start
	@sleep 5
	@./bin/goose up
	@go run -race ./cmd/klubok/main.go

.PHONY: psql
psql:
	@docker exec -it klubok-postgres bash -c "psql -U ${POSTGRES_USER} -d ${POSTGRES_DB}"

.PHONY: tools
tools: tool-govulncheck tool-mockery tool-database tool-openapi

.PHONY: tool-govulncheck
tool-govulncheck:
	@go install -ldflags="-s -w" golang.org/x/vuln/cmd/govulncheck@v1.7.0

.PHONY: tool-mockery
tool-mockery:
	@go install -ldflags="-s -w" github.com/vektra/mockery/v2@v2.53.6

.PHONY: tool-database
tool-database: tool-database-goose tool-database-sqlc

.PHONY: tool-database-goose
tool-database-goose:
	@go install -ldflags="-s -w" github.com/pressly/goose/v3/cmd/goose@v3.30.0

.PHONY: tool-database-sqlc
tool-database-sqlc:
	@go install -ldflags="-s -w" github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0

.PHONY: tool-openapi
tool-openapi:
	@go install -ldflags="-s -w" github.com/ogen-go/ogen/cmd/ogen@v1.20.1
