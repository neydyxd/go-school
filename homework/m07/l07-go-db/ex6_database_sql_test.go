package l07godb

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // регистрирует драйвер "pgx" для database/sql

	"goschool/homework/internal/testdb"
)

func TestListContacts(t *testing.T) {
	t.Parallel()
	_, url := testdb.NewWithURL(t, "testdata/seed.sql")
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	got, err := ListContacts(testCtx(t), db)
	if err != nil {
		t.Fatalf("ListContacts error = %v", err)
	}
	want := []Contact{
		{"Анна Петрова", sql.NullString{String: "+79990000001", Valid: true}},
		{"Борис", sql.NullString{}},
		{"Вера (скидка 50%)", sql.NullString{String: "+79990000003", Valid: true}},
		{"robot_1", sql.NullString{}},
		{"АННА Смирнова", sql.NullString{String: "", Valid: true}},
		{`C:\temp`, sql.NullString{}},
	}
	if len(got) != len(want) {
		t.Fatalf("ListContacts = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ListContacts()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
	if s := db.Stats(); s.InUse != 0 {
		t.Errorf("после ListContacts занято %d соединений, want 0: забыт rows.Close()", s.InUse)
	}
}
