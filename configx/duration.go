package configx

import (
	"encoding/json"
	"time"
)

// Duration 是对 time.Duration 的类型包装，用于支持 JSON 序列化/反序列化
type Duration time.Duration

func (d Duration) Value() time.Duration         { return time.Duration(d) }
func (d Duration) MarshalJSON() ([]byte, error) { return json.Marshal(time.Duration(d).String()) }
func (d *Duration) UnmarshalJSON(data []byte) (err error) {
	var r string
	if err = json.Unmarshal(data, &r); err == nil {
		_d, e := time.ParseDuration(r)
		if err = e; err == nil {
			*d = Duration(_d)
		}
	}
	return
}
