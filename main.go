package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/cnk3x/fsw/configx"
	"github.com/cnk3x/pkg/fsw"
	"github.com/cnk3x/pkg/logx"
	"github.com/cnk3x/pkg/svg"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCommand := NewCommand(BaseName(), func(c *cobra.Command) {
		c.Version = version
		c.PersistentFlags().BoolP("debug", "d", false, "输出调试日志")
		c.PersistentFlags().StringP("config", "c", "", "配置文件路径")

		c.Run = func(cmd *cobra.Command, _ []string) { watchRun(cmd.Context(), GetFlag(cmd, "config")) }
		c.PreRun = func(cmd *cobra.Command, _ []string) {
			logx.Init(lo.Ternary(GetFlagBool(cmd, "debug"), slog.LevelDebug, slog.LevelInfo), false, "fsw")
		}
	})

	initCommand := NewCommand("init", func(c *cobra.Command) {
		c.Short = "初始化配置文件"
		c.Run = func(cmd *cobra.Command, args []string) {
			configFile := GetFlag(cmd, "config", ".fsw.yaml")
			if configx.IsExist(configFile) {
				slog.Info("config file exists")
				return
			}
			if e := configx.WriteYAML(configFile, ConfigDefault()); e != nil {
				slog.Error(fmt.Sprintf("init config failed: %v", e))
				return
			}
			slog.Info("init config", "file", configFile)
		}
	})

	taskCommand := NewCommand("task", func(c *cobra.Command) {
		c.Short = "直接运行任务"
		c.Run = func(cmd *cobra.Command, args []string) {
			config, err := loadConfig(GetFlag(cmd, "config"))
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
		}
	})

	svgCommand := NewCommand("sprite", func(c *cobra.Command) {
		c.Flags().StringP("in", "i", "", "源目录")
		c.Flags().StringP("out", "o", "", "输出文件")
		c.Short = "生成svg雪碧图"
		c.Aliases = append(c.Aliases, "svg", "svg_sprite")
		c.Run = func(cmd *cobra.Command, args []string) {
			in, out := GetFlag(cmd, "in"), GetFlag(cmd, "out")
			if in == "" {
				_ = c.Usage()
				return
			}

			var files []string
			if err := filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if !d.IsDir() && filepath.Ext(path) == ".svg" {
					files = append(files, path)
				}

				return nil
			}); err != nil {
				slog.Error("list files", "err", err)
				return
			}

			if out == "" {
				out = filepath.Base(in) + ".svg"
			}

			if err := svg.Sprite(out, files, svg.NameFromBase(in)); err != nil {
				slog.Error("generate svg sprite", "err", err)
			} else {
				slog.Info("generate svg sprite", "file", out)
			}
		}
	})

	rootCommand.AddCommand(taskCommand, initCommand, svgCommand)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	Run(ctx, rootCommand)
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
		slog.Info("config", "root", f)
	}

	slog.Debug("config", "events", config.Event)

	if len(config.Tasks) > 0 {
		for n, t := range config.Tasks {
			t.Tag = n
			slog.Debug("config task", "task", t.Tag, "type", t.Type)
		}
	}

	for _, r := range config.Triggers {
		if h := Handle(config.Tasks[r.Task]); h != nil {
			slog.Debug("config trigger", "match", r.Match, "event", r.Event, "task", r.Task)
			fw.Handle(r.Task,
				fsw.Match(r.Match...),
				fsw.Events(r.Event),
				fsw.Throttle(r.Throttle.Value()),
				fsw.Handle(h),
			)
		}
	}

	slog.Info("watch start")
	if err := fw.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("watch done", "err", err)
	} else {
		slog.Info("watch done")
	}
}

func loadConfig(configFile string) (config Config, err error) {
	if configFile == "" {
		configFile = configx.FindUpFile(".fsw.yaml", ".fsw.yaml")
	}

	config.ConfigFile, _ = filepath.Abs(configFile)
	config.Base = filepath.Dir(config.ConfigFile)

	slog.Debug("config", "file", configFile)

	if config.Base != "" {
		if pwd, _ := os.Getwd(); pwd != config.Base {
			if e := os.Chdir(config.Base); e != nil {
				err = fmt.Errorf("chdir %s failed: %w", config.Base, e)
				return
			}
			slog.Debug("chdir", "pwd", pwd, "dir", config.Base)
		}
	}

	if err = configx.ReadYAML(configFile, &config); err != nil {
		err = fmt.Errorf("parse config failed: %w", err)
	}
	return
}
