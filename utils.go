package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// GetShell 根据当前操作系统返回合适的 shell 命令及其参数
func GetShell() []string {
	switch runtime.GOOS {
	case "windows":
		s := LookPath("pwsh", "powershell")
		if s != "" {
			return []string{s, "-noprofile", "-nologo"}
		}
		return []string{"cmd"}
	default:
		if s := LookPath(os.Getenv("SHELL"), "bash", "sh", "ash"); s != "" {
			return []string{s}
		}
		return nil
	}
}

// LookPath 依次尝试查找给定名称的可执行文件路径，返回第一个找到的路径或空字符串
func LookPath(names ...string) string {
	for _, name := range names {
		if name = strings.TrimSpace(name); name == "" {
			continue
		}
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}
