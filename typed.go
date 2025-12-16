package main

import (
	"encoding/json"
	"log/slog"
)

type Typed struct {
	Tag   string
	Type  string `json:"type"`
	props any
	raw   []byte
}

type _typed Typed

func (t *Typed) UnmarshalProps(props any) (err error) {
	if err = json.Unmarshal(t.raw, &props); err != nil {
		slog.Error("task::init", "name", t.Tag, "err", err)
		return
	}
	t.props = props
	return
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Typed) UnmarshalJSON(data []byte) error {
	t.raw = data

	slog.Warn(string(data))

	if len(data) == 0 {
		return nil
	}

	if data[0] == '"' {
		return json.Unmarshal(data, &t.Type)
	}

	v := _typed(*t)
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*t = Typed(v)
	return nil
}

// MarshalJSON implements json.Marshaler.
func (t Typed) MarshalJSON() ([]byte, error) {
	if t.props == nil {
		return nil, nil
	}
	return json.Marshal(t.props)
}

var (
	_ json.Unmarshaler = (*Typed)(nil)
	_ json.Marshaler   = Typed{}
)
