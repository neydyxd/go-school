# Задание ex3 урока «CI на GitHub Actions». Запуск из папки homework:
#   make -C m10/l04-ci <цель>
#
# Нужны цели (и make без аргументов печатает их список с описаниями):
#   help       список целей; описание бери из комментария "## ..." после цели
#   fmt        gofmt -w .
#   fmt-check  упасть с кодом 1 и списком файлов, если gofmt -l . не пуст
#   vet        go vet для PKG
#   lint       golangci-lint run для PKG (конфиг .golangci.yml рядом)
#   test       go test -race -count=1 для PKG
#   cover      профиль покрытия в coverage.out и шлагбаум:
#              go run ./cmd/covergate -min $(COVER_MIN) coverage.out
#              Пакеты из cmd/ в покрытие не включай (подсказка: go list и grep -v).
#   build      CGO_ENABLED=0 go build -trimpath -o bin/covergate ./cmd/covergate
#   ci         fmt-check vet lint test cover по порядку
#   clean      удалить coverage.out и bin/covergate
#
# Переменные с значениями по умолчанию, которые можно переопределить
# из командной строки (make cover COVER_MIN=90): GO, PKG (./...), COVER_MIN (80).
# Все цели объяви в .PHONY.

help:
	@echo "TODO: напиши Makefile по условию в комментарии выше"
