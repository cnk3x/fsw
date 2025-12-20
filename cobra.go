package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewCommand(use string, setup func(*cobra.Command)) *cobra.Command {
	c := &cobra.Command{Use: use}
	setup(c)
	return c
}

func GetFlag(cmd *cobra.Command, name string, defaultValue ...string) string {
	s, err := cmd.Flags().GetString(name)
	if err != nil {
		if len(defaultValue) > 0 {
			s = defaultValue[0]
		}
	}
	return s
}

func GetFlagBool(cmd *cobra.Command, name string, defaultValue ...bool) bool {
	s, err := cmd.Flags().GetBool(name)
	if err != nil {
		if len(defaultValue) > 0 {
			s = defaultValue[0]
		}
	}
	return s
}

func BaseName() string {
	return strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
}

func Run(ctx context.Context, c *cobra.Command) {
	c.CompletionOptions.DisableDefaultCmd = true
	c.InitDefaultHelpFlag()
	c.InitDefaultHelpCmd()
	c.InitDefaultVersionFlag()
	c.Flags().MarkHidden("help")
	c.Flags().MarkHidden("version")

	for _, item := range c.Commands() {
		if item.Name() == "help" {
			item.Short = "显示命令信息"
			continue
		}
		Run(context.TODO(), item)
	}

	if c.Parent() == nil {
		if err := c.ExecuteContext(ctx); err != nil {
			slog.Error("execute command failed", "err", err)
		}
	}
}
