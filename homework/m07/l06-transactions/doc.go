// Package l06transactions: домашка урока «Транзакции, изоляция, MVCC».
//
// Все функции получают пул pgx и должны правильно работать, когда их вызывают
// конкурентно. Тесты гоняют параллельные транзакции в настоящем PostgreSQL:
//
//	cd homework && docker compose up -d
//	go test -race -v ./m07/l06-transactions/
package l06transactions
