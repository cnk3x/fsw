package main

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cnk3x/fsw/configx"
	"github.com/cnk3x/pkg/cmdo"
	"github.com/cnk3x/pkg/fsw"
	"github.com/cnk3x/pkg/rex"
	"github.com/cnk3x/pkg/svg"
	"github.com/fsnotify/fsnotify"
)

func Handle(task *Typed) fsw.HandlerFunc {
	if task == nil {
		return nil
	}
	switch task.Type {
	case "oneshot", "cmd":
		return handleOneshot(task)
	case "shell":
		return handleShell(task)
	case "svg_sprite", "svg":
		return handleSvgSprite(task)
	case "echo":
		return func(ctx context.Context, e []fsnotify.Event) {
			slog.Info("handle echo", "events", len(e))
			for _, ev := range e {
				slog.Info("  --", "event", ev.Op.String(), "path", ev.Name)
			}
		}
	}
	return nil
}

func handlerFromTyped[T any](task *Typed, process func(ctx context.Context, cfg T, evt []fsnotify.Event) error) fsw.HandlerFunc {
	var cfg T
	if err := task.UnmarshalProps(&cfg); err != nil {
		slog.Error("task::init", "name", task.Tag, "err", err)
		return nil
	}

	return func(ctx context.Context, e []fsnotify.Event) {
		if err := process(ctx, cfg, e); err != nil {
			slog.Error("task::proc", "name", task.Tag, "err", err)
		} else {
			slog.Info("task::done", "name", task.Tag)
		}
	}
}

type oneshotConfig struct {
	Command []string          `json:"command"`
	Env     map[string]string `json:"env"`
	Dir     string            `json:"dir"`
}

func handleOneshot(task *Typed) fsw.HandlerFunc {
	return handlerFromTyped(task, func(ctx context.Context, cfg oneshotConfig, _ []fsnotify.Event) (err error) {
		c := exec.CommandContext(ctx, cfg.Command[0], cfg.Command[1:]...)
		c.Env = os.Environ()
		return cmdo.Apply(c, cmdo.PKill, cmdo.Env(cfg.Env), cmdo.Dir(cfg.Dir), cmdo.Std).Run()
	})
}

type svgSpriteConfig struct {
	Src    string `json:"src"`
	Dst    string `json:"dst"`
	Pretty bool   `json:"pretty"`
}

func handleSvgSprite(task *Typed) fsw.HandlerFunc {
	return handlerFromTyped(task, func(ctx context.Context, cfg svgSpriteConfig, _ []fsnotify.Event) (err error) {
		if cfg.Src, err = filepath.Abs(cfg.Src); err != nil {
			return
		}
		if cfg.Dst, err = filepath.Abs(cfg.Dst); err != nil {
			return
		}
		slog.Info("svg_sprite", "src", cfg.Src, "dst", cfg.Dst)

		var files []string
		svgMatch := rex.Compile(`/svg/icons/.+\.svg$`)
		selectSvgFile := func(fullPath string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !svgMatch(fullPath) {
				return nil
			}
			files = append(files, fullPath)
			return nil
		}

		if err = filepath.WalkDir(cfg.Src, selectSvgFile); err != nil {
			return
		}

		slog.Info("svg_sprite", "files", len(files))
		return svg.Sprite(cfg.Dst, files, svg.NameFromBase(cfg.Src), svg.Pretty(cfg.Pretty))
	})
}

type shellConfig struct {
	Shell   string            `json:"shell"`
	Command []string          `json:"command"`
	Env     map[string]string `json:"env"`
	Dir     string            `json:"dir"`
	Timeout configx.Duration  `json:"timeout"`
}

func handleShell(task *Typed) fsw.HandlerFunc {
	return handlerFromTyped(task, func(ctx context.Context, cfg shellConfig, _ []fsnotify.Event) (err error) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cmp.Or(cfg.Timeout.Value(), time.Second*10))
		defer cancel()

		shell := GetShell()
		if len(shell) == 0 {
			return fmt.Errorf("shell::init: %w", err)
		}
		c := exec.CommandContext(ctx, shell[0], shell[1:]...)
		c.SysProcAttr = &syscall.SysProcAttr{}
		c.Env = os.Environ()
		c = cmdo.Apply(c, cmdo.PKill, cmdo.Env(cfg.Env), cmdo.Dir(cfg.Dir))
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		stdin, err := c.StdinPipe()
		if err != nil {
			return fmt.Errorf("shell::stdin pipe: %w", err)
		}

		if err = c.Start(); err != nil {
			return fmt.Errorf("shell::start: %w", err)
		}

		for _, cmd := range cfg.Command {
			if _, err = stdin.Write([]byte(cmd + "\n")); err != nil {
				slog.Error("shell::write cmd", "cmd", cmd, "err", err)
			}
		}

		if err = stdin.Close(); err != nil {
			slog.Error("shell::stdin close", "err", err)
		}

		if err = c.Wait(); err != nil {
			return fmt.Errorf("shell::wait: %w", err)
		}

		return nil
	})
}
