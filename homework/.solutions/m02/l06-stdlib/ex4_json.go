package l06stdlib

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"
)

var ErrNoName = errors.New("event without name")

type Event struct {
	Name string    `json:"name"`
	At   time.Time `json:"at"`
	Tags []string  `json:"tags,omitempty"`
}

func EncodeEvent(e Event) ([]byte, error) {
	return json.Marshal(e)
}

func DecodeEvent(data []byte) (Event, error) {
	var e Event
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		return Event{}, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return Event{}, errors.New("лишние данные после JSON-объекта")
	}
	if e.Name == "" {
		return Event{}, ErrNoName
	}
	return e, nil
}
