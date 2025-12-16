package main

import (
	"os"

	"github.com/cnk3x/fsw/configx"
)

type (
	strs     = configx.List[string]
	duration = configx.Duration
	triggers = configx.List[Trigger]
	strmap   = map[string]string
)

func ConfigDefault() Config {
	config := Config{
		Tasks:    map[string]*Typed{"echo": {Type: "echo"}},
		Root:     strs{"."},
		Event:    "cwmd",
		Triggers: triggers{{Task: "echo", Match: strs{`.+`}, Event: "cwmd"}},
	}
	config.Base, _ = os.Getwd()
	return config
}

// Config 定义了整个文件监听器的主要配置结构
type Config struct {
	ConfigFile string `json:"-"` // 配置文件路径，序列化时忽略
	Base       string `json:"-"` // 基础目录，序列化时忽略

	Root     strs     `json:"root"`     // 监听主目录列表，支持多个目录
	Event    string   `json:"event"`    // 监听事件类型，如 c=创建 w=写入 m=修改 d=删除
	Exclude  strs     `json:"exclude"`  // 排除匹配规则，符合这些规则的文件将被忽略
	Throttle duration `json:"throttle"` // 节流时间，防止短时间内重复触发

	Tasks    map[string]*Typed `json:"tasks"`    // 任务配置映射，键为任务名称
	Triggers triggers          `json:"triggers"` // 触发器列表，定义文件变化时执行的任务
}

// Trigger 定义单个触发器规则
type Trigger struct {
	Match strs   `json:"match"` // 匹配文件路径的正则表达式列表，匹配成功则触发
	Event string `json:"event"` // 触发的事件类型，与 Config.Event 对应
	Task  string `json:"task"`  // 要执行的任务名称，对应 Config.Tasks 中的键
}
