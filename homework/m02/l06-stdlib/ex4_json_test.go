package l06stdlib

import (
	"errors"
	"slices"
	"testing"
	"time"
)

func TestEncodeEvent(t *testing.T) {
	at := time.Date(2026, time.March, 9, 14, 5, 7, 0, time.UTC)
	tests := []struct {
		name string
		in   Event
		want string
	}{
		{"с тегами", Event{"deploy", at, []string{"prod", "eu"}}, `{"name":"deploy","at":"2026-03-09T14:05:07Z","tags":["prod","eu"]}`},
		{"без тегов", Event{"rollback", at, nil}, `{"name":"rollback","at":"2026-03-09T14:05:07Z"}`},
		{"пустые теги", Event{"x", at, []string{}}, `{"name":"x","at":"2026-03-09T14:05:07Z"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeEvent(tt.in)
			if err != nil || string(got) != tt.want {
				t.Errorf("EncodeEvent(%+v) = %s, %v, want %s, nil", tt.in, got, err, tt.want)
			}
		})
	}
}

func TestDecodeEvent(t *testing.T) {
	t.Run("лишнее поле при корректном имени", func(t *testing.T) {
		_, err := DecodeEvent([]byte(`{"name":"deploy","nmae":"x","at":"2026-03-09T14:05:07Z"}`))
		if err == nil {
			t.Fatalf("DecodeEvent(с неизвестным полем nmae) err = nil, want ошибку: нужен Decoder.DisallowUnknownFields")
		}
		if errors.Is(err, ErrNoName) {
			t.Errorf("DecodeEvent(с неизвестным полем) err = %v, want ошибку про неизвестное поле, а не ErrNoName", err)
		}
	})
	t.Run("корректный", func(t *testing.T) {
		got, err := DecodeEvent([]byte(`{"name":"deploy","at":"2026-03-09T17:05:07+03:00","tags":["prod"]}`))
		want := time.Date(2026, time.March, 9, 14, 5, 7, 0, time.UTC)
		if err != nil || got.Name != "deploy" || !got.At.Equal(want) || !slices.Equal(got.Tags, []string{"prod"}) {
			t.Errorf("DecodeEvent = %+v, %v, want deploy в %s с тегом prod", got, err, want)
		}
	})
	bad := []struct {
		name   string
		in     string
		target error
	}{
		{"опечатка в поле", `{"nmae":"deploy","at":"2026-03-09T14:05:07Z"}`, nil},
		{"нет имени", `{"at":"2026-03-09T14:05:07Z"}`, ErrNoName},
		{"пустое имя", `{"name":"","at":"2026-03-09T14:05:07Z"}`, ErrNoName},
		{"плохая дата", `{"name":"x","at":"вчера"}`, nil},
		{"два объекта", `{"name":"a"} {"name":"b"}`, nil},
		{"не JSON", `name=deploy`, nil},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeEvent([]byte(tt.in))
			if err == nil {
				t.Fatalf("DecodeEvent(%s) err = nil, want ошибку", tt.in)
			}
			if tt.target != nil && !errors.Is(err, tt.target) {
				t.Errorf("DecodeEvent(%s) err = %v, want errors.Is(err, %v)", tt.in, err, tt.target)
			}
		})
	}
}
