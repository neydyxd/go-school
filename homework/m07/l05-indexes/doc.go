// Package l05indexes: домашка урока «Индексы и EXPLAIN».
//
// В файлах qNN_*.sql пиши CREATE INDEX (или запрос, если так сказано в условии).
// Тесты накатывают сид из testdata/seed.sql в отдельную схему, выполняют твой файл,
// делают VACUUM ANALYZE и проверяют план через EXPLAIN (FORMAT JSON).
//
//	cd homework && docker compose up -d
//	go test -v ./m07/l05-indexes/
package l05indexes
