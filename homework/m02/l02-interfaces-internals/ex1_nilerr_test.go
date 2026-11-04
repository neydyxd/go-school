package l02interfacesinternals

import (
	"errors"
	"testing"
)

func TestValidateAge(t *testing.T) {
	tests := []struct {
		name    string
		age     int
		wantErr bool
	}{
		{"обычный возраст", 30, false},
		{"граница 0", 0, false},
		{"граница 150", 150, false},
		{"отрицательный", -1, true},
		{"слишком большой", 151, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAge(tt.age)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("ValidateAge(%d) = %#v (err != nil: %v), want err != nil: %v", tt.age, err, gotErr, tt.wantErr)
			}
			if !tt.wantErr {
				return
			}
			var ve *ValidationError
			if !errors.As(err, &ve) || ve.Field != "age" {
				t.Errorf("ValidateAge(%d) = %v, want *ValidationError с Field=age", tt.age, err)
			}
		})
	}
}
