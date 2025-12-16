package configx

import "encoding/json"

type List[T any] []T

func (l *List[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	if data[0] == '[' {
		return json.Unmarshal(data, (*([]T))(l))
	}

	var t T
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*l = []T{t}
	return nil
}

func (l List[T]) MarshalJSON() ([]byte, error) {
	switch len(l) {
	case 0:
		return []byte("[]"), nil
	case 1:
		return json.Marshal(l[0])
	default:
		return json.Marshal(l)
	}
}
