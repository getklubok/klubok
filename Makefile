.PHONY: init
init:
	@codegraph init
	@openspec init

.PHONY: require-version
require-version:
	@test -n "$(APP_VERSION)" || { echo "Need to set APP_VERSION"; exit 1; }
	@printf '%s' "$(APP_VERSION)" | grep -Eqx '[0-9]+\.[0-9]+\.[0-9]+' || { echo "Invalid APP_VERSION: $(APP_VERSION)"; exit 1; }

# PUSH_MINOR=true выставляет CI, только если это старший патч в линии X.Y.
.PHONY: build-docker
build-docker: require-version
	@docker buildx build --sbom=true --provenance=true \
		$(if $(PUSH),--push) \
		$(if $(filter true,$(PUSH_LATEST)),-t ghcr.io/getklubok/klubok:latest) \
		$(if $(filter true,$(PUSH_MINOR)),-t ghcr.io/getklubok/klubok:$(shell echo $(APP_VERSION) | cut -d '.' -f -2)) \
		-t ghcr.io/getklubok/klubok:$(APP_VERSION) \
		--label org.opencontainers.image.revision=$(shell git rev-parse HEAD) \
		--label org.opencontainers.image.version=$(APP_VERSION) \
		--label org.opencontainers.image.created=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
		--build-arg APP_VERSION=$(APP_VERSION) \
		-f ./build/Dockerfile .

.PHONY: release
release:
	@$(HOME)/.local/bin/release.sh $(if $(DRY_RUN),--dry-run)
