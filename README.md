# fsw - 功能强大的文件监控工具

## 项目介绍

fsw 是一个轻量级、高性能的文件监控工具，支持多种任务类型，可以根据文件变化自动执行相应操作。它提供了灵活的配置选项，允许您根据不同的文件变化触发不同的任务。

### 功能特点

- 🚀 高性能文件监控
- 📝 灵活的配置系统
- 🛠️ 多种任务类型支持
- ⚡ 支持命令行和配置文件两种使用方式
- 🔧 支持自定义触发规则
- 📊 详细的日志输出

## 安装

### 从源码编译

```bash
go install github.com/cnk3x/fsw@latest
```

### 手动编译

```bash
git clone https://github.com/cnk3x/fsw.git
cd fsw
go build -o fsw
```

## 快速开始

1. **生成默认配置文件**

   运行 fsw，它会自动生成一个默认的配置文件 `.fsw.yaml`：

   ```bash
   fsw
   ```

2. **编辑配置文件**

   根据您的需求修改 `.fsw.yaml` 文件：

   ```yaml
   root: ["."]
   exclude: ["\.git"]

   tasks:
     echo: { type: echo }
     build:
       type: shell
       command: ["go build -v -o fsw"]

   triggers:
     - { task: echo }
     - { task: build, match: ["\.go$", "!_test\.go$"] }
   ```

3. **启动监控**

   ```bash
   fsw
   ```

4. **手动运行任务**

   您也可以手动运行指定的任务：

   ```bash
   fsw task build
   ```

## 配置说明

fsw 使用 YAML 格式的配置文件，默认名称为 `.fsw.yaml`。

### 核心配置

| 字段名   | 类型     | 说明                                                   | 默认值                       |
| -------- | -------- | ------------------------------------------------------ | ---------------------------- |
| root     | []string | 监控的根目录列表                                       | ["."]                        |
| exclude  | []string | 排除的文件/目录正则表达式列表                          | []                           |
| event    | []string | 监控的事件类型（CREATE, WRITE, CHMOD, REMOVE, RENAME） | ["CREATE", "WRITE", "CHMOD"] |
| throttle | duration | 事件节流时间，避免频繁触发                             | 500ms                        |

### 任务配置

任务配置位于 `tasks` 节点下，每个任务有一个唯一的名称和对应的配置。

### 触发器配置

触发器配置位于 `triggers` 节点下，用于定义哪些任务在什么条件下被触发。

| 字段名 | 类型     | 说明                         |
| ------ | -------- | ---------------------------- |
| task   | string   | 触发的任务名称               |
| match  | []string | 匹配的文件路径正则表达式列表 |

## 任务类型

### echo

简单的调试任务，用于输出事件信息。

```yaml
tasks:
  echo_task:
    type: echo
```

### shell

执行 Shell 命令，可以配置多个命令。

| 字段名  | 类型     | 说明               | 默认值           |
| ------- | -------- | ------------------ | ---------------- |
| shell   | string   | 使用的 Shell 程序  | 系统默认 Shell   |
| command | []string | 要执行的命令列表   | []               |
| env     | map      | 环境变量设置       | {}               |
| dir     | string   | 执行命令的工作目录 | 配置文件所在目录 |
| timeout | duration | 命令执行超时时间   | 10s              |

```yaml
tasks:
  build:
    type: shell
    command: ["go build -v", "echo 'Build completed'"]
    env: { "GOOS": "linux" }
    dir: ./src
    timeout: 30s
```

### oneshot/cmd

执行一次性命令，适合长时间运行的进程。

| 字段名  | 类型     | 说明               | 默认值           |
| ------- | -------- | ------------------ | ---------------- |
| command | []string | 要执行的命令       | []               |
| env     | map      | 环境变量设置       | {}               |
| dir     | string   | 执行命令的工作目录 | 配置文件所在目录 |

```yaml
tasks:
  server:
    type: oneshot
    command: ["./server", "--port", "8080"]
    dir: ./bin
```

### svg_sprite/svg

将多个 SVG 文件合并为一个 SVG Sprite。

| 字段名 | 类型   | 说明                       | 默认值 |
| ------ | ------ | -------------------------- | ------ |
| src    | string | SVG 文件所在目录           | -      |
| dst    | string | 生成的 SVG Sprite 文件路径 | -      |
| pretty | bool   | 是否生成格式化的输出       | false  |

```yaml
tasks:
  svg:
    type: svg_sprite
    src: ./icons
    dst: ./dist/icons.svg
    pretty: true
```

## 命令行参数

| 参数名   | 缩写 | 类型   | 说明         |
| -------- | ---- | ------ | ------------ |
| --debug  | -d   | bool   | 输出调试日志 |
| --config | -c   | string | 配置文件路径 |

### 子命令

#### task

手动运行指定的任务：

```bash
fsw task [task_name]
```

## 示例配置

```yaml
root: ["./src", "./test"]
exclude: ["\.git", "node_modules"]
event: ["CREATE", "WRITE"]
throttle: 1s

# 定义任务
tasks:
  # 简单的echo任务，用于调试
  debug:
    type: echo

  # 编译Go代码
  build:
    type: shell
    command: ["go build -v -o ./bin/app"]
    timeout: 30s

  # 运行测试
  test:
    type: shell
    command: ["go test ./..."]
    timeout: 60s

  # 启动开发服务器
  dev:
    type: oneshot
    command: ["./bin/app", "--dev"]

  # 生成SVG Sprite
  icons:
    type: svg_sprite
    src: ./assets/icons
    dst: ./public/icons.svg
    pretty: true

# 定义触发器
triggers:
  # 所有文件变化都触发debug任务
  - { task: debug }

  # .go文件变化时触发build任务（排除测试文件）
  - { task: build, match: ["\.go$", "!_test\.go$"] }

  # _test.go文件变化时触发test任务
  - { task: test, match: ["_test\.go$"] }

  # 开发服务器相关文件变化时重启服务器
  - { task: dev, match: ["\.go$", "\.yaml$"] }

  # SVG文件变化时重新生成Sprite
  - { task: icons, match: ["\.svg$"] }
```

## 使用场景

### 1. 自动编译

当您修改源代码文件时，fsw 可以自动编译您的项目，节省您的时间。

### 2. 自动化测试

当测试文件变化时，自动运行相关测试，确保代码质量。

### 3. 开发服务器热重载

当配置文件或源代码变化时，自动重启开发服务器，实现热重载。

### 4. 静态资源处理

当静态资源（如 SVG 图标）变化时，自动生成对应的 Sprite 文件或进行其他处理。

### 5. 部署自动化

当代码提交或标签创建时，自动执行部署流程。

## 日志输出

fsw 使用结构化日志输出，您可以通过 `--debug` 参数查看详细的调试信息。

## 注意事项

1. 确保配置文件的格式正确，否则 fsw 会尝试生成默认配置
2. 合理设置 `throttle` 参数，避免频繁触发任务
3. 使用 `exclude` 参数排除不必要的文件和目录，提高性能

## 许可证

MIT License
