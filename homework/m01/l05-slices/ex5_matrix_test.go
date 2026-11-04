package l05slices

import "testing"

func TestNewMatrixShape(t *testing.T) {
	tests := []struct {
		name       string
		rows, cols int
		wantNil    bool
	}{
		{"3x4", 3, 4, false},
		{"1x1", 1, 1, false},
		{"нулевые строки", 0, 4, true},
		{"нулевые столбцы", 3, 0, true},
		{"отрицательный размер", -1, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatrix(tt.rows, tt.cols)
			if tt.wantNil {
				if m != nil {
					t.Errorf("NewMatrix(%d, %d) = %v, want nil", tt.rows, tt.cols, m)
				}
				return
			}
			if len(m) != tt.rows {
				t.Fatalf("NewMatrix(%d, %d): строк %d, want %d", tt.rows, tt.cols, len(m), tt.rows)
			}
			for i, row := range m {
				if len(row) != tt.cols || cap(row) != tt.cols {
					t.Errorf("строка %d: len=%d cap=%d, want len=cap=%d", i, len(row), cap(row), tt.cols)
				}
				for j, v := range row {
					if v != 0 {
						t.Errorf("m[%d][%d] = %v, want 0", i, j, v)
					}
				}
			}
		})
	}
}

func TestNewMatrixRowsIndependent(t *testing.T) {
	m := NewMatrix(2, 3)
	if len(m) != 2 {
		t.Fatalf("NewMatrix(2, 3): строк %d, want 2", len(m))
	}
	m[0][2] = 1
	m[0] = append(m[0], 42)
	if m[1][0] != 0 {
		t.Errorf("запись в строку 0 изменила строку 1: m[1][0] = %v, want 0", m[1][0])
	}
}

func TestNewMatrixAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(20, func() { _ = NewMatrix(100, 100) })
	if allocs > 2 {
		t.Errorf("NewMatrix(100, 100) делает %.0f аллокаций, want не больше 2", allocs)
	}
}
