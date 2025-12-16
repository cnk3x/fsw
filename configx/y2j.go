package configx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"

	"go.yaml.in/yaml/v3"
)

// JSONToYAML 将 JSON 字节序列转换为 YAML 字节序列。
func JSONToYAML(jsonData []byte) (yamlData []byte, err error) {
	var tmp any
	if err = json.Unmarshal(jsonData, &tmp); err != nil {
		return
	}

	yamlData, err = yaml.Marshal(tmp)
	return
}

// YAMLToJSON 将 YAML 字节序列转换为 JSON 字节序列。
func YAMLToJSON(yamlData []byte) (jsonData []byte, err error) {
	var buf bytes.Buffer
	if err = Y2JStream(bytes.NewReader(yamlData), &buf); err != nil {
		return
	}
	jsonData = buf.Bytes()
	return
}

// Y2JStream 从输入流读取 YAML 文档，转换为 JSON 后写入输出流，每份文档以换行分隔。
func Y2JStream(in io.Reader, out io.Writer) (err error) {
	for decoder := yaml.NewDecoder(in); ; {
		var data any
		if err = decoder.Decode(&data); err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}

		if err = transformData(&data); err != nil {
			return
		}

		output, e := json.Marshal(data)
		if err = e; err != nil {
			return
		}

		if _, err = out.Write(output); err != nil {
			return
		}

		if _, err = io.WriteString(out, "\n"); err != nil {
			return
		}
	}
}

func transformData(pIn *any) (err error) {
	switch in := (*pIn).(type) {
	case float64:
		if math.IsInf(in, 1) {
			*pIn = "+Inf"
		} else if math.IsInf(in, -1) {
			*pIn = "-Inf"
		} else if math.IsNaN(in) {
			*pIn = "NaN"
		}
		return
	case map[any]any:
		m := make(map[string]any, len(in))
		for k, v := range in {
			if err = transformData(&v); err != nil {
				return
			}
			var sk string
			switch v := k.(type) {
			case string:
				sk = v
			case int:
				sk = strconv.Itoa(v)
			case bool:
				sk = strconv.FormatBool(v)
			case nil:
				sk = "null"
			case float64:
				f := v
				if math.IsInf(f, 1) {
					sk = "+Inf"
				} else if math.IsInf(f, -1) {
					sk = "-Inf"
				} else if math.IsNaN(f) {
					sk = "NaN"
				} else {
					sk = strconv.FormatFloat(f, 'f', -1, 64)
				}
			default:
				err = fmt.Errorf("type mismatch: expect map key string or int; got: %T", k)
				return
			}
			m[sk] = v
		}
		*pIn = m
	case map[string]any:
		m := make(map[string]any, len(in))
		for k, v := range in {
			if err = transformData(&v); err != nil {
				return
			}
			m[k] = v
		}
		*pIn = m
	case []any:
		for i := len(in) - 1; i >= 0; i-- {
			if err = transformData(&in[i]); err != nil {
				return
			}
		}
	}
	return
}
