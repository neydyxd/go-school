// Package l08sqlcormmigrations: домашка урока «sqlc, ORM, миграции».
//
// В migrations/ лежат миграции goose, которые ты пишешь (задача 1).
// В ex2_nplus1.go рабочий, но медленный код с N+1, его надо переписать.
// В ex3_filter.go самодельный query builder с плейсхолдерами.
//
//	cd homework && docker compose up -d
//	go test -race -v ./m07/l08-sqlc-orm-migrations/
package l08sqlcormmigrations
