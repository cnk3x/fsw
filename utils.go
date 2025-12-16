package main

import (
	"os/exec"
	"runtime"
)

// GetShell 根据当前操作系统返回合适的 shell 命令及其参数
func GetShell() []string {
	switch runtime.GOOS {
	case "linux", "darwin":
		return []string{"bash"}
	case "windows":
		s := LookPath("pwsh", "powershell")
		if s != "" {
			return []string{s, "-noprofile", "-nologo"}
		}
		return []string{"cmd"}
	default:
		s := LookPath("sh", "bash", "ash", "zsh")
		if s != "" {
			return []string{s}
		}
		return nil
	}
}

// LookPath 依次尝试查找给定名称的可执行文件路径，返回第一个找到的路径或空字符串
func LookPath(names ...string) string {
	for _, name := range names {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}
