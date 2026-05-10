PACKAGE_NAME ?= saas_baseon_go
RELEASE_DIR ?= dist/release
RELEASE_ARTIFACT ?= $(RELEASE_DIR)/$(PACKAGE_NAME).tar.gz

.PHONY: run test build docker-up docker-down package check-source-clean check-release

run:
	go run ./cmd/api

test:
	go test ./...

build:
	go build -o bin/saas-api ./cmd/api

docker-up:
	docker compose up --build

docker-down:
	docker compose down

package:
	./scripts/package-release.sh "$(RELEASE_ARTIFACT)"

check-source-clean:
	./scripts/check-source-clean.sh

check-release:
	./scripts/check-release.sh --self-test
	$(MAKE) check-source-clean
	$(MAKE) package
	./scripts/check-release.sh "$(RELEASE_ARTIFACT)"
