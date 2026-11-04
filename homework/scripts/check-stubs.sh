#!/usr/bin/env bash
# Проверяет, что каждая заготовка с тестами падает (домашка не решена "сама").
# Exit 0: ни один пакет не проходит. Exit 1: список пакетов, проходящих на заготовке.
set -uo pipefail
cd "$(dirname "$0")/.."
pattern="${1:-./...}"
go vet "$pattern" || { echo "go vet упал на заготовках"; exit 2; }
bad=()
for pkg in $(go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' "$pattern"); do
  if go test -count=1 "$pkg" >/dev/null 2>&1; then bad+=("$pkg"); fi
done
if ((${#bad[@]})); then
  printf 'Проходят на заготовке (слабые тесты):\n'; printf '  %s\n' "${bad[@]}"; exit 1
fi
echo "OK: все заготовки красные"
