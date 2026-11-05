// Package l03riveradvanced: домашка урока «River: ретраи, уникальность, периодика».
//
// Свой backoff и NextRetry, snooze и отмена, уникальные задачи, расписание
// периодической задачи и идемпотентный платёжный воркер. Тесты идут в настоящий
// PostgreSQL, таблицы River мигрируются в схему каждого теста:
//
//	cd homework && docker compose up -d
//	go test -race -v ./m09/l03-river-advanced/
package l03riveradvanced
