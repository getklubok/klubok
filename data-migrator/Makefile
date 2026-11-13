.PHONY: init
init: deps

.PHONY: deps
deps:
	@go mod tidy

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: integration-postgres
integration-postgres:
	@go test -tags "integration postgres" ./db/... -v -p 1 -count=1 -run $(or $(RUN),TestIntegration)

.PHONY: docker-compose-integration-tests
docker-compose-integration-tests:
	@export PATH="$(shell pwd)/bin:$(PATH)"; ./tests/integration/run-in-docker-compose.sh $(RUN)

.PHONY: build
build:
	@docker build -t klubok_data_migrator .

.PHONY: run
run:
	@docker run -d --rm --network=$(NETNAME) --name klubok_data_migrator -p 3000:3000 klubok_data_migrator

.PHONY: logs
logs:
	@docker logs klubok_data_migrator

.PHONY: stop
stop:
	@docker stop klubok_data_migrator

# Команды запуска ручек

.PHONY: exec-fill-order-shapes
exec-fill-order-shapes:
	@docker exec -t klubok_postgres pg_dump -c -U postgres --table=diagram_shapes > ./klubok_fos_dump_before_$(shell date +%Y%m%d_%H%M%S).sql
	@curl -X POST http://127.0.0.1:3000/fill-order-shapes
	@docker exec -t klubok_postgres pg_dump -c -U postgres --table=diagram_shapes > ./klubok_fos_dump_after_$(shell date +%Y%m%d_%H%M%S).sql
