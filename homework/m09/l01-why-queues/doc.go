// Package l01whyqueues: домашка урока «Зачем очереди».
//
// Ретраи, идемпотентный потребитель, своя очередь на SELECT FOR UPDATE SKIP LOCKED
// и transactional outbox. Тесты с базой идут в настоящий PostgreSQL:
//
//	cd homework && docker compose up -d
//	go test -race -v ./m09/l01-why-queues/
package l01whyqueues
