#!/usr/bin/env bash
# Накладывает эталоны из .solutions поверх заготовок во временной копии
# и прогоняет vet + тесты с детектором гонок. Exit 0: все эталоны зелёные.
# Аргумент (необязательно): шаблон пакетов, например ./m04/...
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
pattern="${1:-./...}"
count="${COUNT:-1}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
rsync -a --exclude .solutions "$root/" "$tmp/"
cp -R "$root/.solutions/." "$tmp/"
cd "$tmp"
go vet "$pattern"
go test -race -count="$count" "$pattern"
