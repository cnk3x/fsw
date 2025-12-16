package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/cnk3x/fsw/configx"
	"github.com/cnk3x/pkg/fsw"
	"github.com/spf13/cobra"
)

func main() {
	rc := &cobra.Command{
		Use: filepath.Base(os.Args[0]),
		Run: func(cmd *cobra.Command, args []string) {
			configFile, _ := cmd.Flags().GetString("config")
			watchRun(cmd.Context(), configFile)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}
		},
	}

	rc.PersistentFlags().BoolP("debug", "d", false, "输出调试日志")
	rc.PersistentFlags().StringP("config", "c", "", "配置文件路径")

	rc.AddCommand(taskRun())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := rc.ExecuteContext(ctx); err != nil {
		slog.Error("execute command failed", "err", err)
		os.Exit(1)
	}
}

func watchRun(ctx context.Context, configFile string) {
	config, err := loadConfig(configFile)
	if err != nil {
		slog.Error("load config", "err", err)
		return
	}

	fw := fsw.New(fsw.Options{
		Root:     config.Root,
		Exclude:  config.Exclude,
		Event:    config.Event,
		Throttle: config.Throttle.Value(),
	})

	for _, f := range config.Root {
		slog.Info("watch root", "dir", f)
	}

	if len(config.Tasks) > 0 {
		for n, t := range config.Tasks {
			t.Tag = n
			slog.Info("register task", "task", n, "type", t.Type)
		}

		for _, r := range config.Triggers {
			if h := Handle(config.Tasks[r.Task]); h != nil {
				slog.Info("register trigger", "task", r.Task, "match", r.Match)
				fw.Handle(r.Task, fsw.Match(r.Match...), fsw.Handle(h))
			}
		}
	}

	if err := fw.Run(ctx); err != nil {
		slog.Error("watch run failed", "err", err)
	} else {
		slog.Info("watch run success")
	}
}

func loadConfig(configFile string) (config Config, err error) {
	if configFile == "" {
		var file string
		if file, err = configx.FindUpFile(".fsw.yaml"); err == nil {
			configFile = file
			if workDir := filepath.Dir(configFile); workDir != "" {
				if pwd, _ := os.Getwd(); pwd != workDir {
					if e := os.Chdir(workDir); e != nil {
						err = fmt.Errorf("chdir %s failed: %w", workDir, e)
						return
					}
					slog.Info("chdir", "pwd", pwd, "dir", workDir)
				}
			}
		}
	}

	if configFile == "" {
		configFile = ".fsw.yaml"
	}

	config.ConfigFile, _ = filepath.Abs(configFile)
	config.Base = filepath.Dir(config.ConfigFile)

	slog.Info("config", "file", configFile)
	if err = configx.ReadYAML(configFile, &config); err != nil {
		if os.IsNotExist(err) {
			e := configx.WriteYAML(configFile, ConfigDefault())
			if e != nil {
				err = fmt.Errorf("init config failed: %w", e)
				return
			}
			slog.Info("init config", "file", configFile)
			return
		}
		err = fmt.Errorf("parse config failed: %w", err)
	}
	return
}

func taskRun() *cobra.Command {
	return &cobra.Command{
		Use: "task",
		Run: func(cmd *cobra.Command, args []string) {
			configFile, _ := cmd.Flags().GetString("config")
			config, err := loadConfig(configFile)
			if err != nil {
				slog.Error("load config", "err", err)
				return
			}

			for n, t := range config.Tasks {
				t.Tag = n
			}

			if tag, ok := configx.ArrAt(args, 0); ok {
				if f := Handle(config.Tasks[tag]); f != nil {
					f(cmd.Context(), nil)
					return
				}
			}
			_ = cmd.Usage()
		},
	}
}
