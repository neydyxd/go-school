// Package l02riverbasics: домашка урока «River: основы».
//
// Аргументы задачи, постановка в одной транзакции с бизнес-данными (InsertTx),
// идемпотентный воркер и клиент с очередями. Тесты идут в настоящий PostgreSQL,
// таблицы River мигрируются в схему каждого теста:
//
//	cd homework && docker compose up -d
//	go test -race -v ./m09/l02-river-basics/
package l02riverbasics
