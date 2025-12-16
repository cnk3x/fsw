package configx

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// FindUpFile 从当前工作目录开始，逐级向上查找名为 name 的常规文件，返回其绝对路径；若未找到则返回 os.ErrNotExist。
func FindUpFile(name string) (string, error) {
	for curDir, _ := os.Getwd(); curDir != "" && curDir != filepath.VolumeName(curDir) && curDir != filepath.Dir(curDir); curDir = filepath.Dir(curDir) {
		curFile := filepath.Join(curDir, name)
		if stat, e := os.Stat(curFile); e == nil && stat.Mode().IsRegular() {
			return curFile, nil
		}
	}
	return "", os.ErrNotExist
}

// ReadYAML 读取指定 YAML 文件，将其转换为 JSON 后反序列化到 config 指针中，返回可能的错误。
func ReadYAML[T any](file string, config *T) (err error) {
	var data []byte
	if data, err = os.ReadFile(file); err != nil {
		return
	}

	if data, err = YAMLToJSON(data); err != nil {
		return
	}

	return json.Unmarshal(data, &config)
}

// WriteYAML 将 config 序列化为 JSON 后转换为 YAML 格式，并写入指定文件，文件权限为 0666，返回可能的错误。
func WriteYAML[T any](file string, config T) (err error) {
	var data []byte
	if data, err = json.Marshal(config); err != nil {
		return
	}

	if data, err = JSONToYAML(data); err != nil {
		return
	}

	return os.WriteFile(file, data, 0666)
}

// ArrAt 安全获取切片 arr 中索引 i 的元素，支持负索引（从末尾开始计数）。若索引越界，返回零值和 false。
func ArrAt[T any](arr []T, i int) (T, bool) {
	l := len(arr)

	if i < 0 {
		i += l
	}

	if i < 0 || i >= l {
		var zero T
		return zero, false
	}

	return arr[i], true
}

// ArrAtOr 安全获取切片 arr 中索引 i 的元素，支持负索引。若索引越界，返回 def 提供的默认值（若未提供则返回零值）。
func ArrAtOr[T any](arr []T, i int, def ...T) (r T) {
	if v, ok := ArrAt(arr, i); ok {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return
}
