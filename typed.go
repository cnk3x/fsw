package main

import (
	"encoding/json"
	"log/slog"
)

type Typed struct {
	Tag   string
	Type  string
	props any
	raw   []byte
}

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
	var typed struct {
		Tag  string `json:"tag"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typed); err != nil {
		return err
	}
	t.Tag = typed.Tag
	t.Type = typed.Type
	t.raw = data
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
