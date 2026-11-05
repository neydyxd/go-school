# Запуск из папки homework: make -C m10/l04-ci <цель>
# Цели повторяют шаги CI, поэтому «зелёный make ci локально» почти всегда
# значит «зелёный CI».

GO        ?= go
PKG       ?= ./...
COVER_MIN ?= 80
# main-пакеты это обвязка без тестов: с Go 1.22 они попадают в профиль
# с 0% и тянут общее покрытие вниз. Меряем только пакеты с логикой.
COVER_PKG ?= $(shell $(GO) list $(PKG) | grep -v /cmd/)
BIN       := bin

.DEFAULT_GOAL := help

.PHONY: help fmt fmt-check vet lint test cover build ci clean

help: ## показать цели
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  %-10s %s\n", $$1, $$2}'

fmt: ## отформатировать код
	gofmt -w .

fmt-check: ## упасть, если есть неотформатированные файлы
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt не прогнан:"; echo "$$out"; exit 1; fi

vet: ## go vet
	$(GO) vet $(PKG)

lint: ## golangci-lint с конфигом .golangci.yml
	golangci-lint run $(PKG)

test: ## тесты с детектором гонок
	$(GO) test -race -count=1 $(PKG)

cover: ## покрытие и шлагбаум COVER_MIN
	$(GO) test -count=1 -coverprofile=coverage.out $(COVER_PKG)
	$(GO) run ./cmd/covergate -min $(COVER_MIN) coverage.out

build: ## собрать covergate в bin/
	CGO_ENABLED=0 $(GO) build -trimpath -o $(BIN)/covergate ./cmd/covergate

ci: fmt-check vet lint test cover ## всё, что делает CI, по порядку

clean: ## удалить артефакты
	rm -f coverage.out $(BIN)/covergate
